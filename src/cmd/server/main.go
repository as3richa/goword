package main

import (
	"flag"
	"os"
	"os/signal"
	"syscall"

	"internal/engine"
	"internal/log"
	"internal/server"
)

var addressFlag = flag.String("address", ":8080", "address to listen on")
var debugFlag = flag.Bool("debug", false, "enable debug output")

func main() {
	flag.Parse()

	if *debugFlag {
		log.EnableDebug()
	}

	wait := make(chan error)
	go signalHandler(func() { wait <- nil })
	go func() { wait <- server.Server(*addressFlag) }()

	engine := engine.New()
	go engine.Run()
	client := engine.NewClient()
	response := <-client.Pipe
	log.Fields{"response": response}.Debug("received resposne from engine")

	if err := <-wait; err != nil {
		log.Fields{"error": err}.Fatal("unexpected top-level crash")
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
