package fs

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"
)

type Version struct {
	Version string `json:"version"`
}

var targetDir string
var proxyHost string

func ServerInitialize(dir string, proxy string) {
	targetDir = dir
	proxyHost = proxy
}

func ServerGet(w http.ResponseWriter, r *http.Request) {
	log.Println(r.URL.Path)
	params := r.URL.Query()
	dir := params.Get(":dir")
	key := params.Get(":key")
	version := params.Get(":version")

	target := formatPath(dir, key, version)

	if _, err := os.Stat(target); os.IsNotExist(err) {
		log.Printf("No such file: %s\n", target)
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
	}

	http.ServeFile(w, r, target)
}

func ServerPut(w http.ResponseWriter, r *http.Request) {
	log.Println(r.URL.Path)
	params := r.URL.Query()
	dir := params.Get(":dir")
	key := params.Get(":key")
	v := params.Get(":version")

	r.ParseMultipartForm(32 << 20)
	file, _, err := r.FormFile("file")
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
	defer file.Close()

	if _, err := os.Stat(targetDir + "/" + dir); os.IsNotExist(err) {
		os.Mkdir(targetDir+"/"+dir, 0700)
	}

	out, err := os.Create(formatPath(dir, key, v))
	if err != nil {
		log.Println(err)
	}
	defer out.Close()

	io.Copy(out, file)

	result, _ := json.Marshal(&Version{Version: v})
	fmt.Fprintf(w, string(result))
}

func formatPath(dir string, key string, version string) string {
	if version == "0" {
		return targetDir + "/" + dir + "/" + key
	}
	return targetDir + "/" + dir + "/" + key + "-" + version
}

func ServerMigration() {
	filepath.Walk(targetDir, func(p string, info os.FileInfo, err error) (e error) {
		if info.IsDir() {
			return nil
		}

		p = strings.Replace(p, path.Clean(targetDir)+"/", "", 1)
		log.Println(p)
		res, err := http.PostForm(
			proxyHost+"/api/migration",
			url.Values{"key": {p}},
		)

		for res.StatusCode != http.StatusOK {
			res, err = http.PostForm(
				proxyHost+"/api/migration",
				url.Values{"key": {p}},
			)
		}

		return nil
	})
}
