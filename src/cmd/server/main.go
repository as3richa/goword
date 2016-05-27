package main

import (
  "flag"
  "net/http"
  "time"
  "os"
  "os/signal"
  stdlog "log"
  "syscall"

  "internal/log"
)

var addressFlag = flag.String("address", ":8080", "address to listen on")

func main() {
	wait := make(chan struct{})
  go signalHandler(func() { close(wait) })
	go func() { server(); close(wait) }()
	<-wait
}

func server() {
	log.Fields{"address": *addressFlag}.Info("starting http server")
	w := log.Writer()
	defer w.Close()
	s := &http.Server{
		Addr:           *addressFlag,
		Handler:        router(),
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
		ErrorLog:       stdlog.New(w, "", 0),
	}
	if err := s.ListenAndServe(); err != nil {
		log.Fields{"error": err}.Info("unexpected error from http server")
	}
}

func signalHandler(cancel func()) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, os.Kill, syscall.SIGTERM)
	for sig := range c {
		log.Fields{"signal": sig}.Info("received signal - terminating")
		cancel()
	}
}
