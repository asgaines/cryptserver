package server

import (
	"fmt"
	"time"
	"net/http"
	"log"
	"github.com/asgaines/cryptserver/utils"
)

func handleCrypt(w http.ResponseWriter, req *http.Request) {
	// "/" in router matches all paths
	// Guard against extraneous paths
	if req.URL.Path == "/" {
		if unhashedPass := req.PostFormValue("password"); unhashedPass != "" {
			if _, err := fmt.Fprintf(w, "%v\n", utils.Encode(unhashedPass)); err != nil {
				log.Println(err)
			}
		} else {
			msg := "Request must provide password to be hashed"
			http.Error(w, msg, http.StatusBadRequest)  // 400
		}
	} else {
		http.Error(w, "Invalid path", http.StatusNotFound)  // 404
	}

	// Delay response to client
	if delay := req.Context().Value("delay"); delay != nil {
		time.Sleep(delay.(time.Duration))
	} else {
		log.Fatal("Delay time not successfully passed into context")
	}
}

func handleShutdown(w http.ResponseWriter, req *http.Request) {
	// Collection of password hashes which are accepted as authorized
	// to shutdown the server
	passHashes := req.Context().Value("passHashes")
	if passHashes == nil {
		log.Fatal("Password hashes not successfully passed into context")
	}

	if _, ok := passHashes.(map[string]bool)[utils.Encode(req.PostFormValue("password"))]; ok {
		// User provided password which hashed to an acceptable value
		// Send the shutdown signal through channel
		if shutdown := req.Context().Value("signalChan"); shutdown != nil {
			shutdown.(chan<- bool) <- true
		} else {
			log.Fatal("Shutdown channel not successfully passed into context")
		}
	} else {
		http.Error(w, "Invalid password", http.StatusUnauthorized)  // 401

		// This keeps a user from brute-forcing password requests
		if delay := req.Context().Value("delay"); delay != nil {
			time.Sleep(delay.(time.Duration))
		} else {
			log.Fatal("Delay time not successfully passed into context")
		}
	}
}

