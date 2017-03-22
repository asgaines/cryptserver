package main

import (
	"fmt"
	"net/http"
	"context"
	"time"
)

func createServer(port int, delay time.Duration, shutdown chan bool, passHashes map[string]bool) http.Server {
	mux := http.NewServeMux()

	// Create handlers from handle functions
	cryptHandleFunc := http.HandlerFunc(handleCrypt)
	shutdownHandleFunc := http.HandlerFunc(handleShutdown)

	// Inject time to delay into handler context
	cryptHandler := addDelayContext(cryptHandleFunc, delay)
	shutdownHandler := addDelayContext(shutdownHandleFunc, delay)

	// Inject shutdown channel
	shutdownHandler = addShutdownContext(shutdownHandler, shutdown)
	// Inject password hashes map loaded from file
	shutdownHandler = addPassHashesContext(shutdownHandler, passHashes)

	// Add logging to request handling
	cryptHandler = addLogging(cryptHandler)
	shutdownHandler = addLogging(shutdownHandler)

	// Create request routing
	mux.Handle("/shutdown", shutdownHandler)
	mux.Handle("/", cryptHandler)

	return http.Server{
		Handler: mux,
		Addr: fmt.Sprintf(":%d", port),
	}
}

func gracefulShutdown(server *http.Server, delay time.Duration) {
	// Cap the shutdown timeout, the amount of time allowed
	// for a request to be handled gracefully
	ctx, _ := context.WithTimeout(context.Background(), delay * 3)
	// Start the graceful shutdown
	if err := server.Shutdown(ctx); err != nil {
		fmt.Println(err)
	}
}

