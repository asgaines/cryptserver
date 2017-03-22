package server

import (
	"fmt"
	"net"
	"time"
	"testing"
	"net/http"
	"net/url"
	"io/ioutil"
	"strings"
	"github.com/asgaines/cryptserver/utils"
)

var passwordEncodings = []struct {
	password string
	passHash string
}{
	// {"password", "correct-hash-of-password"}
	{"angryMonkey", "ZEHhWB65gUlzdVwtDQArEyx+KVLzp/aTaRaPlBzYRIFj6vjFdqEb0Q5B8zVKCZ0vKbZPZklJz0Fd7su2A+gf7Q=="},
	{"blowfish1234", "CG9ZxAdtMgJfBbBjtTmznVrAH/bIKMYG9AOvLx/+P/4kIaCkXzhSi7K6TYfEnHCB/cicK2A6BBfZL6q48V25SA=="},
	{"a87&1hkA!l*Q12n6i2&Q", "RRBeSqawrv0y1LrVZb13RhHneaHkSvzAvacPttI+j+SQEcri19wr+fD2qOqzcw7C404jaYXSne0sg39/eO7eaA=="},
}

func TestCryptValidPasswords(t *testing.T) {
	port := 8081
	shutdown := make(chan bool)
	passHashes := utils.LoadPassHashes("../test/etc/shadow")
	delay := 0 * time.Second

	httpServer := Create(port, delay, shutdown, passHashes)
	defer httpServer.Close()

	listener, err := net.Listen("tcp", fmt.Sprintf(":%v", port))
	if err != nil {
		t.Error(err)
	}
	defer listener.Close()

	go func() {
		httpServer.Serve(listener)
	}()

	for _, c := range passwordEncodings {
		resp, _ := http.PostForm(fmt.Sprintf("http://localhost:%d/", port), url.Values{"password": {c.password}})
		defer resp.Body.Close()

		data, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			t.Errorf("Failed to read data")
		}

		result := strings.TrimSpace(string(data))

		if result != c.passHash {
			t.Errorf("Password %q returned %q, expected %q", c.password, result, c.passHash)
		}
	}
}

func TestCryptInvalidPasswords(t *testing.T) {
	port := 8082
	shutdown := make(chan bool)
	passHashes := utils.LoadPassHashes("../test/etc/shadow")
	delay := 0 * time.Second

	httpServer := Create(port, delay, shutdown, passHashes)
	defer httpServer.Close()

	listener, err := net.Listen("tcp", fmt.Sprintf(":%v", port))
	if err != nil {
		t.Error(err)
	}
	defer listener.Close()

	go func() {
		httpServer.Serve(listener)
	}()

	resp, err := http.PostForm(fmt.Sprintf("http://localhost:%d/", port), url.Values{"password": {""}})
	if err != nil {
		t.Error(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 400 {
		t.Errorf("Response code %v is not 400", resp.StatusCode)
	}
}

func TestShutdownValidPassword(t *testing.T) {
	port := 8083
	shutdown := make(chan bool)
	end := make(chan bool)
	passHashes := utils.LoadPassHashes("../test/etc/shadow")
	delay := 1 * time.Second

	httpServer := Create(port, delay, shutdown, passHashes)

	listener, err := net.Listen("tcp", fmt.Sprintf(":%v", port))
	if err != nil {
		t.Error(err)
	}
	defer listener.Close()

	go func() {
		httpServer.Serve(listener)
	}()

	password := "angryMonkey"
	var resp *http.Response

	go func() {
		resp, _ = http.PostForm(fmt.Sprintf("http://localhost:%d/shutdown", port), url.Values{"password": {password}})
		defer resp.Body.Close()
		end <- true
	}()

	// Channel receives from shutdown handler
	<-shutdown
	<-end
	GracefulShutdown(&httpServer, delay)

	if resp.StatusCode != 200 {
		t.Errorf("Status code was %v when expecting 200", resp.StatusCode)
	}
}

func TestShutdownIsGraceful(t *testing.T) {
	port := 8084
	shutdown := make(chan bool)
	passHashes := utils.LoadPassHashes("../test/etc/shadow")
	delay := 200 * time.Millisecond

	httpServer := Create(port, delay, shutdown, passHashes)

	listener, err := net.Listen("tcp", fmt.Sprintf(":%v", port))
	if err != nil {
		t.Error(err)
	}
	defer listener.Close()

	go func() {
		httpServer.Serve(listener)
	}()

	validPassword := "angryMonkey"

	var resp *http.Response

	go func() {
		resp, _ = http.PostForm(fmt.Sprintf("http://localhost:%d/", port), url.Values{"password": {"password123"}})
		defer resp.Body.Close()
	}()

	// Wait halfway through request response time, then issue a shutdown request
	time.Sleep(delay / 2)

	go func() {
		resp, _ := http.PostForm(fmt.Sprintf("http://localhost:%d/shutdown", port), url.Values{"password": {validPassword}})
		defer resp.Body.Close()
	}()

	<-shutdown // Receives from server shutdown handler
	GracefulShutdown(&httpServer, delay)

	if resp.StatusCode != 200 {
		t.Error("Server was shutdown before processing request completed")
	}
}
