package main

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"log"
)

func main() {
	// Generate RSA key pair
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		log.Fatalf("Error generating key pair: %v", err)
	}
	publicKey := &privateKey.PublicKey

	// Log private key and public key
	fmt.Printf("Private Key: %v\n", privateKey)
	fmt.Printf("Public Key: %v\n", publicKey)

	// Save private key as PEM
	privateKeyBytes := x509.MarshalPKCS1PrivateKey(privateKey)
	privateKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: privateKeyBytes,
	})
	err = ioutil.WriteFile("private_key.pem", privateKeyPEM, 0600)
	if err != nil {
		log.Fatalf("Error saving private key: %v", err)
	}

	// Log private key bytes and PEM
	fmt.Printf("Private Key Bytes: %x\n", privateKeyBytes)
	fmt.Printf("Private Key PEM: %s\n", privateKeyPEM)

	// Save public key as PEM
	publicKeyBytes, err := x509.MarshalPKIXPublicKey(publicKey)
	if err != nil {
		log.Fatalf("Error marshaling public key: %v", err)
	}
	publicKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PUBLIC KEY",
		Bytes: publicKeyBytes,
	})
	err = ioutil.WriteFile("public_key.pem", publicKeyPEM, 0644)
	if err != nil {
		log.Fatalf("Error saving public key: %v", err)
	}

	// Log public key bytes and PEM
	fmt.Printf("Public Key Bytes: %x\n", publicKeyBytes)
	fmt.Printf("Public Key PEM: %s\n", publicKeyPEM)

	// Sign data
	message := []byte("Hello, world!")
	fmt.Printf("Message: %s\n", message)
	hash := sha256.New()
	_, err = hash.Write(message)
	if err != nil {
		log.Fatalf("Error hashing message: %v", err)
	}
	hashedMessage := hash.Sum(nil)
	fmt.Printf("Hashed Message: %x\n", hashedMessage)
	signature, err := rsa.SignPKCS1v15(rand.Reader, privateKey, crypto.SHA256, hashedMessage)
	if err != nil {
		log.Fatalf("Error signing message: %v", err)
	}

	// Verify signature
	err = rsa.VerifyPKCS1v15(publicKey, crypto.SHA256, hashedMessage, signature)
	if err != nil {
		log.Fatalf("Error verifying signature: %v", err)
	} else {
		fmt.Println("Signature verified.")
	}

	// Encrypt data
	ciphertext, err := rsa.EncryptPKCS1v15(rand.Reader, publicKey, message)
	if err != nil {
		log.Fatalf("Error encrypting message: %v", err)
	}

	// Decrypt data
	plaintext, err := rsa.DecryptPKCS1v15(rand.Reader, privateKey, ciphertext)
	if err != nil {
		log.Fatalf("Error decrypting message: %v", err)
	}

	if string(plaintext) == string(message) {
		fmt.Println("Decrypted message matches original.")
	}
}
