package main

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
)
func randomBase64(n int) string {
	b := make([]byte, n)
	_, err := rand.Read(b)
	if err != nil {
		panic("Gagal menghasilkan random bytes: " + err.Error())
	}
	return base64.StdEncoding.EncodeToString(b)
}
func main() {
	fmt.Println("--- COPY KE .ENV ---")
	fmt.Printf("SESSION_KEY=\"%s\"\n\n", randomBase64(32))
	fmt.Println("--------------------")
}