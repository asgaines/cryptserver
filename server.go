package main

import (
	"fmt"
	"net/http"
	"context"
	"time"
	"log"
	"os"
)

func addContext(next http.Handler, shutdown chan bool) http.Handler {
	return http.HandlerFunc(func (w http.ResponseWriter, r *http.Request) {
		// Load the map of accepted password hashes enumerated in
		// the path passed to the function
		passHashes := loadPassHash("./etc/shadow")

		ctx := context.WithValue(r.Context(), "signalChan", shutdown)
		ctx = context.WithValue(ctx, "passHashes", passHashes)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func createServer(port int, shutdown chan bool) http.Server {
	mux := http.NewServeMux()

	mux.Handle("/shutdown", addContext(http.HandlerFunc(handleShutdown), shutdown))
	mux.HandleFunc("/", handleCrypt)

	return http.Server{
		Handler: mux,
		Addr: fmt.Sprintf(":%d", port),
	}
}

func gracefulShutdown(server *http.Server, complete chan bool) error {
	// Cap the shutdown at 5 seconds, the amount of time allowed
	// for a request to be handled gracefully
	ctx, _ := context.WithTimeout(context.Background(), 5 * time.Second)
	// Start the graceful shutdown
	err := server.Shutdown(ctx)
	complete <-true
	return err
}

func listenShutdown(server *http.Server, interrupt chan os.Signal, shutdown chan bool, complete chan bool) {
	select {
	case <- interrupt:
		if err := gracefulShutdown(server, complete); err != nil {
			log.Println(err)
		}
	case <- shutdown:
		if err := gracefulShutdown(server, complete); err != nil {
			log.Println(err)
		}
	}
}

