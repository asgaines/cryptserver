package main

import (
	"fmt"
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
	port := 8080
	shutdown := make(chan bool)
	server := createServer(port, shutdown)
	go server.ListenAndServe()

	for _, c := range passwordEncodings {
		resp, _ := http.PostForm(fmt.Sprintf("http://localhost:%d/", port), url.Values{"password": {c.password}})
		data, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			t.Errorf("Failed to read data")
		}
		defer resp.Body.Close()

		result := strings.TrimSpace(string(data))

		if result != c.passHash {
			t.Errorf("Password %q returned %q, expected %q", c.password, result, c.passHash)
		}

	}
}

func TestCryptInvalidPasswords(t *testing.T) {
	port := 8080
	shutdown := make(chan bool)
	server := createServer(port, shutdown)
	go server.ListenAndServe()

	invalidPasswords := []string{""}
	for _, password := range invalidPasswords {
		resp, err := http.PostForm(fmt.Sprintf("http://localhost:%d/", port), url.Values{"password": {password}})
		if err != nil {
			t.Errorf("Failed to read data")
		}
		if resp.StatusCode != 400 {
			t.Errorf("Response code %v is not 400", resp.StatusCode)
		}
	}
}

