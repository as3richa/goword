package server

import (
	stdlog "log"
	"net/http"
	"time"

	"internal/engine"
	"internal/log"
)

func Server(address string) error {
	log.Fields{"address": address}.Info("starting http server")

	engine := engine.New()
	go engine.Run()
	defer engine.Terminate()

	w := log.Writer()
	defer w.Close()
	s := &http.Server{
		Addr:           address,
		Handler:        router(engine),
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
		ErrorLog:       stdlog.New(w, "", 0),
	}
	if err := s.ListenAndServe(); err != nil {
		log.Fields{"error": err}.Error("unexpected error from http server")
		return err
	}
	return nil
}
