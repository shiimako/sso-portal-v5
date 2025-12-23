package main

import (
	"fmt"

	"github.com/SherClockHolmes/webpush-go"
)

func main() {
    privateKey, publicKey, err := webpush.GenerateVAPIDKeys()
    if err != nil {
        panic(err)
    }
	admin := "{{isi dengan email admin}}"

    fmt.Println("--- COPY KE .ENV ---")
    fmt.Printf("VAPID_PRIVATE_KEY=\"%s\"\n", privateKey)
    fmt.Printf("VAPID_PUBLIC_KEY=\"%s\"\n", publicKey)
    fmt.Printf("VAPID_SUBJECT=\"mailto:%s\"\n", admin)
    fmt.Println("--------------------")
}