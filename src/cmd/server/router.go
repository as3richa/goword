package main

import (
	"net/http"

	"internal/log"

	"github.com/julienschmidt/httprouter"
)

func router() http.Handler {
	router := httprouter.New()
	router.GET("/", index)
	return loggingHandler(router)
}

type loggingResponseWriter struct {
	http.ResponseWriter
	code int
}

func (l *loggingResponseWriter) WriteHeader(code int) {
	l.ResponseWriter.WriteHeader(code)
	l.code = code
}

func (l *loggingResponseWriter) Write(data []byte) (int, error) {
	return l.ResponseWriter.Write(data)
}

func loggingHandler(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		lw := &loggingResponseWriter{ResponseWriter: w, code: http.StatusOK}
		h.ServeHTTP(lw, r)
		log.Fields{"url": r.URL.String(), "method": r.Method, "source": r.RemoteAddr, "response": lw.code}.Info("handled request")
	})
}
