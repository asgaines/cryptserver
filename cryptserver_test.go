package main

import (
	"fmt"
	"os"
	"os/exec"
	"net"
	"time"
	"testing"
	"net/http"
	"net/url"
	"io/ioutil"
	"strings"
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

func TestEncodeValidPasswords(t *testing.T) {
	passwordEncodings := append(
		passwordEncodings,
		struct {
			password string
			passHash string
		}{"", "z4PhNX7vuL3xVChQ1m2AB9Yg5AULVxXcg/SpIdNs6c5H0NE8XYXysP+DGNKHfuwvY7kxvUdBeoGlODJ6+SfaPg=="})

	for _, c := range passwordEncodings {
		result := encode(c.password)
		if result != c.passHash {
			t.Errorf("encode(%q) returned %q, wanted %q", c.password, result, c.passHash)
		}
	}
}

func TestLoadPasswordHashes(t *testing.T) {
	testFilename := "./test/etc/shadow"
	passHashes := loadPassHashes(testFilename)

	successCases := []struct {
		passHash string
	}{
		{"ZEHhWB65gUlzdVwtDQArEyx+KVLzp/aTaRaPlBzYRIFj6vjFdqEb0Q5B8zVKCZ0vKbZPZklJz0Fd7su2A+gf7Q=="},
		{"e9hKlUdVXmDsA0o+4xs3tmyhvSV1LZ+Yx5DtCKkpX1A9TzZvZiWANcufAQEzJcsKXlpyeSzQ+CoLaLYGJR8uzg=="},
		{"OuhrA2sIlN6RNFbYH9PtT+VP3yDKdcmDzzRgAQxPe0KNmWWj/AxmP8y1dVYrzJ7BISnEx9edk9Vu1E6if+05vg=="},
	}

	// Test all password hashes in file are recognized
	for _, c := range successCases {
		if _, ok := passHashes[c.passHash]; !ok {
			t.Errorf("%q should be present in hash map", c.passHash)
		}
	}
}

func TestLoadPasswordHashesFail(t *testing.T) {
	testFilename := "./test/etc/shadow"
	passHashes := loadPassHashes(testFilename)

	failCases := []struct {
		passHash string
	}{
		{"XEHhWB65gUlzdVwtDQArEyx+KVLzp/aTaRaPlBzYRIFX6vjFdqEb0Q5B8zVKCZ0vKbZPZklJz0Fd7su2A+gf7X=="},
		{"X9hKlUdVXmDsA0o+4xs3tmyhvSV1LZ+Yx5DtCKkpX1AXTzZvZiWANcufAQEzJcsKXlpyeSzQ+CoLaLYGJR8uzX=="},
		{"XuhrA2sIlN6RNFbYH9PtT+VP3yDKdcmDzzRgAQxPe0KXmWWj/AxmP8y1dVYrzJ7BISnEx9edk9Vu1E6if+05vX=="},
	}

	// Test password hashes not in file are not recognized
	for _, c := range failCases {
		if _, ok := passHashes[c.passHash]; ok {
			t.Errorf("%q should not be present in hash map", c.passHash)
		}
	}
}

func TestCryptValidPasswords(t *testing.T) {
	port := 8081
	shutdown := make(chan bool)
	passHashes := loadPassHashes("./etc/shadow")
	delay := 0 * time.Second

	server := createServer(port, delay, shutdown, passHashes)
	defer server.Close()

	listener, err := net.Listen("tcp", fmt.Sprintf(":%v", port))
	if err != nil {
		t.Error(err)
	}
	defer listener.Close()

	go func() {
		fmt.Println(server.Serve(listener))
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
	passHashes := loadPassHashes("./etc/shadow")
	delay := 0 * time.Second

	server := createServer(port, delay, shutdown, passHashes)
	defer server.Close()

	listener, err := net.Listen("tcp", fmt.Sprintf(":%v", port))
	if err != nil {
		t.Error(err)
	}
	defer listener.Close()

	go func() {
		fmt.Println(server.Serve(listener))
	}()

	resp, err := http.PostForm(fmt.Sprintf("http://localhost:%d/", port), url.Values{"password": {""}})
	if err != nil {
		t.Error(err)
		fmt.Println(err)
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
	passHashes := loadPassHashes("./etc/shadow")
	delay := 1 * time.Second

	server := createServer(port, delay, shutdown, passHashes)

	listener, err := net.Listen("tcp", fmt.Sprintf(":%v", port))
	if err != nil {
		t.Error(err)
	}
	defer listener.Close()

	go func() {
		fmt.Println(server.Serve(listener))
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
	gracefulShutdown(&server, delay)

	if resp.StatusCode != 200 {
		t.Errorf("Status code was %v when expecting 200", resp.StatusCode)
	}
}

func TestShutdownIsGraceful(t *testing.T) {
	port := 8084
	shutdown := make(chan bool)
	passHashes := loadPassHashes("./etc/shadow")
	delay := 200 * time.Millisecond

	server := createServer(port, delay, shutdown, passHashes)

	listener, err := net.Listen("tcp", fmt.Sprintf(":%v", port))
	if err != nil {
		t.Error(err)
	}
	defer listener.Close()

	go func() {
		fmt.Println(server.Serve(listener))
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
	gracefulShutdown(&server, delay)

	if resp.StatusCode != 200 {
		t.Error("Server was shutdown before processing request completed")
	}
}

func TestSIGINTHandledGracefully(t *testing.T) {
	var status int
	delay := 100 * time.Millisecond
	port := 8080

	cmd := exec.Command("./cryptserver",
		"--delay", fmt.Sprintf("%v", delay),
		"--port", fmt.Sprintf("%v", port))

	if err := cmd.Start(); err != nil {
		t.Error(err)
	}

	// Give the server a little time to spool up
	time.Sleep(100 * time.Millisecond)

	go func() {
		resp, err := http.PostForm(fmt.Sprintf("http://localhost:%v", port), url.Values{"password": {"encrypt_this"}})
		if err != nil {
			t.Error(err)
		}
		status = resp.StatusCode
	}()

	// Issue SIGINT halfway through request response time
	time.Sleep(delay / 2)
	cmd.Process.Signal(os.Interrupt)

	cmd.Wait()

	if status != 200 {
		t.Error("Received status %d, expected 200", status)
	}
}

func TestSIGTERMNotHandledGracefully(t *testing.T) {
	delay := 100 * time.Millisecond
	port := 8080

	cmd := exec.Command("./cryptserver",
		"--delay", fmt.Sprintf("%v", delay),
		"--port", fmt.Sprintf("%v", port))

	if err := cmd.Start(); err != nil {
		t.Error(err)
	}

	// Give the server a little time to spool up
	time.Sleep(100 * time.Millisecond)

	go func() {
		resp, _ := http.PostForm(fmt.Sprintf("http://localhost:%v", port), url.Values{"password": {"encrypt_this"}})
		if resp != nil {
			t.Error("Response did not fail")
		}
	}()

	// Issue SIGINT halfway through request response time
	time.Sleep(delay / 2)
	cmd.Process.Signal(os.Kill)

	cmd.Wait()
}

