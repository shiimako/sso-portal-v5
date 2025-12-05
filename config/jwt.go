package config

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
)

var (
	PrivateKey *rsa.PrivateKey
	PublicKey  *rsa.PublicKey
	Issuer     string
	Audience   string
)

func LoadKeys() error {
	// Load private key
	privData, err := os.ReadFile(os.Getenv("JWT_PRIVATE_KEY_PATH"))
	if err != nil {
		return err
	}

	privBlock, _ := pem.Decode(privData)
	PrivateKey, err = x509.ParsePKCS1PrivateKey(privBlock.Bytes)
	if err != nil {
		return err
	}

	// Load public key
	pubData, err := os.ReadFile(os.Getenv("JWT_PUBLIC_KEY_PATH"))
	if err != nil {
		return err
	}

	pubBlock, _ := pem.Decode(pubData)
	PublicKey, err = x509.ParsePKCS1PublicKey(pubBlock.Bytes)
	if err != nil {
		return err
	}

	// Load issuer
	Issuer = os.Getenv("JWT_ISSUER")

	if Issuer == "" {
		return fmt.Errorf("JWT issuer kosong")
	}

	return nil
}