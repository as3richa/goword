package main

import (
  "fmt"
  "net/http"

  "github.com/julienschmidt/httprouter"
)

func index(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
    r.Header.Add("Content-Type", "text/plain")
    fmt.Fprint(w, "Hello world!")
}
