package utils

import (
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"log"
	
	"golang.org/x/crypto/scrypt"
)

// Compute a base64 digest for password
func Crypt(password string) string {
    if password == "" {
        return ""
	}

	salt := make([]byte, 24)
	rand.Read(salt)
    return base64.StdEncoding.EncodeToString(append(salt, Digest(password, salt)...))
}

// Verify a candidate password against a reference base64 digest
func Verify(candidate string, reference string) error {
	decoded, err := base64.StdEncoding.DecodeString(reference)
	if err != nil {
		return err
	}

	// salt = b64decode(reference)[:24]
    salt := decoded[:24]
	// saved_digest = b64decode(reference)[24:]
    saved_digest := decoded[24:]
    if !bytes.Equal(saved_digest, Digest(candidate, salt)) {
		return errors.New("creds: Password mismatch")
	}

	return nil
}

func Digest(password string, salt []byte) []byte {
	dk, err := scrypt.Key([]byte(password), salt, 1<<13, 8, 2, 32)
	if err != nil {
		log.Fatal(err)
	}
	return dk
}