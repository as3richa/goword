package main

import (
	"flag"
	"os"
	"os/signal"
	"syscall"

	"internal/log"
	"internal/server"
)

var addressFlag = flag.String("address", ":8080", "address to listen on")

func main() {
	flag.Parse()

	wait := make(chan error)
	go signalHandler(func() { wait <- nil })
	go func() { wait <- server.Server(*addressFlag) }()

	if err := <-wait; err != nil {
		log.Fields{"error": err}.Fatal("server crashed unexpectedly")
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
