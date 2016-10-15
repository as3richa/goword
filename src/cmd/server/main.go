package main

import (
	"flag"
	"math/rand"
	"os"
	"os/signal"
	"syscall"
	"time"

	"internal/engine"
	"internal/log"
	"internal/server"
)

var addressFlag = flag.String("address", ":8080", "address to listen on")
var debugFlag = flag.Bool("debug", false, "enable debug output")

func main() {
	rand.Seed(time.Now().UTC().UnixNano())

	flag.Parse()

	if *debugFlag {
		log.EnableDebug()
	}

	engine := engine.New()
	go engine.Run()

	wait := make(chan error)
	go signalHandler(wait)
	go func() { wait <- server.Server(*addressFlag, engine) }()

	if err := <-wait; err != nil {
		log.Fields{"error": err}.Fatal("unexpected top-level crash")
	}
}

func signalHandler(die chan error) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, os.Kill, syscall.SIGTERM)
	for sig := range c {
		log.Fields{"signal": sig}.Info("received signal - terminating")
		close(die)
	}
}
