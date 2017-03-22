package main

import (
	"log"
	"net/http"
	"context"
	"time"
)

func addLogging(next http.Handler) http.Handler {
	return http.HandlerFunc(func (w http.ResponseWriter, r *http.Request) {
		// Log basic information
		log.Printf("%s %s %s", r.RemoteAddr, r.Method, r.URL)
		next.ServeHTTP(w, r)
	})
}

func addDelayContext(next http.Handler, delay time.Duration) http.Handler {
	return http.HandlerFunc(func (w http.ResponseWriter, r *http.Request) {
		// Inject response delay time into context
		ctx := context.WithValue(r.Context(), "delay", delay)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func addPassHashesContext(next http.Handler, passHashes map[string]bool) http.Handler {
	return http.HandlerFunc(func (w http.ResponseWriter, r *http.Request) {
		// Inject password hashes into context
		ctx := context.WithValue(r.Context(), "passHashes", passHashes)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func addShutdownContext(next http.Handler, shutdown chan<- bool) http.Handler {
	return http.HandlerFunc(func (w http.ResponseWriter, r *http.Request) {
		// Inject shutdown channel into context
		ctx := context.WithValue(r.Context(), "signalChan", shutdown)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

