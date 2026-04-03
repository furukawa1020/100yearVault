package main

import (
	"bytes"
	"fmt"
	"io"
	"log"

	"filippo.io/age"
)

func main() {
	passphrase := "100year-secret-key"
	message := []byte("This is a secret message for 100 years later.")

	// Encrypt
	recipient, err := age.NewScryptRecipient(passphrase)
	if err != nil {
		log.Fatal(err)
	}
	
	buf := &bytes.Buffer{}
	w, err := age.Encrypt(buf, recipient)
	if err != nil {
		log.Fatal(err)
	}
	if _, err := w.Write(message); err != nil {
		log.Fatal(err)
	}
	if err := w.Close(); err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Encrypted length: %d\n", buf.Len())

	// Decrypt
	identity, err := age.NewScryptIdentity(passphrase)
	if err != nil {
		log.Fatal(err)
	}
	r, err := age.Decrypt(buf, identity)
	if err != nil {
		log.Fatal(err)
	}
	
	out := &bytes.Buffer{}
	if _, err := io.Copy(out, r); err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Decrypted message: %s\n", out.String())
	
	if bytes.Equal(message, out.Bytes()) {
		fmt.Println("Age encryption test passed.")
	} else {
		fmt.Println("Age encryption test failed.")
	}
}
