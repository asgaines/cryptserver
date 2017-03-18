package main

import (
	"os"
	"crypto/sha512"
	"encoding/base64"
	"bufio"
)

func encode(password string) string {
	hasher := sha512.New()
	hasher.Write([]byte(password))
	return base64.StdEncoding.EncodeToString(hasher.Sum(nil))
}

func loadPassHash(filename string) map[string]bool {
	f, err := os.Open(filename)
	if err != nil {
		panic(err)
	}

	defer f.Close()

	passHashMap := make(map[string]bool)
	scanner := bufio.NewScanner(f)

	for scanner.Scan() {
		passHashMap[scanner.Text()] = true
	}

	return passHashMap
}
