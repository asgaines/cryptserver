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

	// Channel used to issue shutdown command from HTTP handler
	shutdown := make(chan bool)

	// Channel to block program exit until http.Serve message logged
	complete := make(chan bool)

	// Capture SIGINT (<Ctrl-C> or `kill -2` signals)
	interrupt := make(chan os.Signal)
	signal.Notify(interrupt, os.Interrupt)

	server := createServer(port, shutdown)

	go func() {
		// ListenAndServe always returns non-nil error
		log.Println(server.ListenAndServe())
		// Logging complete: unblock program exit
		complete <- true
	}()

	// Block until shutdown request issued
	select {
	case <- interrupt:
		// Shutdown issued through SIGINT
		gracefulShutdown(&server)
	case <- shutdown:
		// Shutdown issued from HTTP handler
		gracefulShutdown(&server)
	}

	// Block until http.Serve message logged
	<-complete
}

