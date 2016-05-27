package server

import (
	"fmt"
	"net/http"
	"path"
)

type errorPage struct {
	Path, ContentType string
	Status            int
}

func errorHandler(status int) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		serveStaticFile(w, r, path.Join("static", fmt.Sprintf("%d.html", status)), status, "text/html")
	}
}
