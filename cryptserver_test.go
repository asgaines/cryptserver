package main

import (
	"fmt"
	"os"
	"os/exec"
	"time"
	"testing"
	"net/http"
	"net/url"
)


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
		t.Errorf("Received status %d, expected 200", status)
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

