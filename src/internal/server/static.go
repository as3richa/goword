package server

import (
	"io"
	"net/http"
	"os"
	"path"

	"internal/log"

	"github.com/julienschmidt/httprouter"
)

type staticFile struct {
	Path, ContentType string
}

var staticRoutes = map[string]staticFile{
	"/":      {path.Join("static", "index.html"), "text/html"},
	"/shell": {path.Join("static", "shell.html"), "text/html"},

	"/favicon.ico": {path.Join("static", "favicon.ico"), "image/x-icon"},

	"/style.css": {path.Join("static", "style.css"), "text/css"},
	"/shell.css": {path.Join("static", "shell.css"), "text/css"},

	"/game.js": {path.Join("static", "game.js"), "application/javascript"},

	"/compass.svg": {path.Join("static", "compass.svg"), "image/svg+xml"},
	"/skull.svg":   {path.Join("static", "skull.svg"), "image/svg+xml"},

	"/cubes.json": {path.Join("config", "cubes.json"), "application/json"},
	"/words.list": {path.Join("config", "words.list"), "text/plain"},
}

func staticHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	serveStaticFile(w, r, staticRoutes[r.URL.Path].Path, http.StatusOK, staticRoutes[r.URL.Path].ContentType)
}

func serveStaticFile(w http.ResponseWriter, r *http.Request, path string, status int, contentType string) {
	file, err := os.Open(path)
	if err != nil {
		log.Fields{"error": err, "path": r.URL.Path}.Panic("failed to serve static asset")
		return
	}
	defer file.Close()

	w.Header().Set("Content-Type", contentType)
	w.WriteHeader(status)
	io.Copy(w, file)
}
