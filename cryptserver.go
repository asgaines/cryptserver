package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
)

func main() {
	var port int
	// Optionally set port, default is 8080
	flag.IntVar(&port, "port", 8080, "Port number for server connection")
	// Load the variables with values from command line
	flag.Parse()

	interrupt := make(chan os.Signal)
	shutdown := make(chan bool)
	complete := make(chan bool)
	signal.Notify(interrupt, os.Interrupt)

	server := createServer(port, shutdown)

	go listenShutdown(&server, interrupt, shutdown, complete)

	// ListenAndServe always returns non-nil error
	log.Println(server.ListenAndServe())
	<-complete
}

