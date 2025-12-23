package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
)

func main() {
	if _, err := os.Stat("keys"); os.IsNotExist(err) {
		err := os.Mkdir("keys", 0755)
		if err != nil {
			panic("Gagal membuat folder keys: " + err.Error())
		}
	}

	fmt.Println("Menghasilkan RSA keypair...")

	// Generate RSA 2048-bit keypair
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		panic("Gagal generate private key: " + err.Error())
	}

	// Encode private key ke format PEM
	privBytes := x509.MarshalPKCS1PrivateKey(privateKey)
	privPem := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: privBytes,
	})

	err = os.WriteFile("keys/private.pem", privPem, 0600)
	if err != nil {
		panic("Gagal menulis private.pem: " + err.Error())
	}

	// Encode public key ke format PEM
	pubBytes := x509.MarshalPKCS1PublicKey(&privateKey.PublicKey)
	pubPem := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PUBLIC KEY",
		Bytes: pubBytes,
	})

	err = os.WriteFile("keys/public.pem", pubPem, 0644)
	if err != nil {
		panic("Gagal menulis public.pem: " + err.Error())
	}

	fmt.Println("RSA Keypair berhasil dibuat di folder /keys")
	fmt.Println("- keys/private.pem")
	fmt.Println("- keys/public.pem")
}
