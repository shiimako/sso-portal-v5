package main

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
)

func main() {
	b := make([]byte, 32) // 32 bytes = 256 bits untuk SHA-256
	_, err := rand.Read(b)
	if err != nil {
		panic(err)
	}

	fmt.Println("WEBHOOK_SECRET=" + hex.EncodeToString(b))
}
