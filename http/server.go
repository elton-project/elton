package http

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"../api"
)

type Result struct {
	Version int64 `json:"version"`
	Length  int64 `json:"length"`
}

type Path struct {
	Path []string `json:"path"`
}

var server api.Server
var isMigration bool

func InitServer(dir string, flag bool) {
	server = api.NewServer(dir, "")
	isMigration = flag
}

func ServerGetHandler(w http.ResponseWriter, r *http.Request) {
	params := r.URL.Query()
	dir := params.Get(":dir")
	key := params.Get(":key")
	version := params.Get(":version")

	target, err := server.Read(dir, key, version)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	}

	http.ServeFile(w, r, target)
}

func ServerPutHandler(w http.ResponseWriter, r *http.Request) {
	params := r.URL.Query()
	dir := params.Get(":dir")
	key := params.Get(":key")
	version := params.Get(":version")

	r.ParseMultipartForm(32 << 20)
	file, _, err := r.FormFile("file")
	if err != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
	defer file.Close()

	err = server.Create(dir, key, version, file)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}

	v, _ := strconv.ParseInt(version, 10, 64)
	result, _ := json.Marshal(&Result{Version: v, Length: r.ContentLength})
	fmt.Fprintf(w, string(result))
}

func ServerDeleteHandler(w http.ResponseWriter, r *http.Request) {
	params := r.URL.Query()
	dir := params.Get(":dir")
	key := params.Get(":key")

	err := server.Delete(dir, key)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
}

func ServerMigrationHandler(w http.ResponseWriter, r *http.Request) {
	if isMigration {
		result, _ := json.Marshal(&Path{server.Migration()})
		fmt.Fprintf(w, string(result))
		return
	}
}
