package main

import (
	"crypto/rand"
	"encoding/hex"
	"log"
)

// Generate a random ID
func generateRandomID() string {
	bytes := make([]byte, 16) // 16 bytes = 128 bits
	if _, err := rand.Read(bytes); err != nil {
		log.Fatal("Error generating random ID:", err)
	}
	return hex.EncodeToString(bytes)
}
