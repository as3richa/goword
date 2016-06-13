package server

import (
	"bufio"
	"net"
	"net/http"

	"internal/log"

	"github.com/julienschmidt/httprouter"
)

func router() http.Handler {
	router := httprouter.New()

	router.GET("/internal/grid/", gridHandler)
	router.GET("/internal/grid/:seed", gridHandler)
	for route := range staticRoutes {
		router.GET(route, staticHandler)
	}
	router.GET("/engine", engineHandler)

	router.RedirectTrailingSlash = true
	router.RedirectFixedPath = true
	router.HandleMethodNotAllowed = true
	router.NotFound = errorHandler(404)
	router.MethodNotAllowed = errorHandler(405)
	router.PanicHandler = func(w http.ResponseWriter, r *http.Request, _ interface{}) {
		errorHandler(500)(w, r)
	}

	return loggingHandler(router)
}

type loggingResponseWriter struct {
	http.ResponseWriter
	code     int
	hijacked bool
}

func (l *loggingResponseWriter) WriteHeader(code int) {
	l.code = code
	l.ResponseWriter.WriteHeader(code)
}

func (l *loggingResponseWriter) Write(data []byte) (int, error) {
	return l.ResponseWriter.Write(data)
}

func (l *loggingResponseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	l.hijacked = true
	return l.ResponseWriter.(http.Hijacker).Hijack()
}

func loggingHandler(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		lw := &loggingResponseWriter{ResponseWriter: w, code: http.StatusOK}
		h.ServeHTTP(lw, r)
		log.Fields{
			"url":      r.URL.String(),
			"method":   r.Method,
			"source":   r.RemoteAddr,
			"response": lw.code,
			"hijacked": lw.hijacked,
		}.Info("handled request")
	})
}
