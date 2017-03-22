package main

import (
	"fmt"
	"net/http"
	"context"
	"time"
)

func createServer(port int, shutdown chan bool) http.Server {
	mux := http.NewServeMux()

	mux.Handle("/shutdown", addContext(http.HandlerFunc(handleShutdown), shutdown))
	mux.HandleFunc("/", handleCrypt)

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

