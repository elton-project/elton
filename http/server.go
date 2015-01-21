package http

import (
	"encoding/json"
	"fmt"
	"net/http"

	"nashio-lab.info/elton/api"
)

type Version struct {
	Version string `json:"version"`
}

var server api.Server

func InitServer(dir string, host string) {
	server = api.NewServer(dir, host)
}

func ServerGetHandler(w http.ResponseWriter, r *http.Request) {
	params := r.URL.Query()
	dir := params.Get(":dir")
	key := params.Get(":key")
	version := params.Get(":version")

	target, err := server.Read(dir, key, version)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
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
	}
	defer file.Close()

	err = server.Create(dir, key, version, file)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}

	result, _ := json.Marshal(&Version{Version: version})
	fmt.Fprintf(w, string(result))
}

func ServerMigration() {
	server.Migration()
}
