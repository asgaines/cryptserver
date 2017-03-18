package main

import (
	"fmt"
	"time"
	"net/http"
	"log"
)

func handleCrypt(w http.ResponseWriter, req *http.Request) {
	if req.URL.Path == "/" {
		if unhashedPass := req.PostFormValue("password"); unhashedPass != "" {
			if _, err := fmt.Fprintf(w, "%v\n", encode(unhashedPass)); err != nil {
				log.Println(err)
			}
		} else {
			msg := "Request must specify value to be hashed as password"
			http.Error(w, msg, http.StatusBadRequest)  // 400
		}
	} else {
		http.Error(w, "Invalid endpoint", http.StatusNotFound)  // 404
	}

	time.Sleep(5 * time.Second)
}

func handleShutdown(w http.ResponseWriter, req *http.Request) {
	passHashes := req.Context().Value("passHashes")
	if passHashes == nil {
		log.Fatal("Value not successfully passed")
	}

	if _, ok := passHashes.(map[string]bool)[encode(req.PostFormValue("password"))]; ok {
		// Send the shutdown signal through channel
		if shutdown := req.Context().Value("signalChan"); shutdown != nil {
			shutdown.(chan bool) <- true
		} else {
			fmt.Println("no signal...")
		}
	} else {
		http.Error(w, "Invalid password", http.StatusUnauthorized)  // 401
		// This keeps a user from brute-forcing password requests
		time.Sleep(5 * time.Second)
	}
}

