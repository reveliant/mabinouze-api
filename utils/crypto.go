package utils

import (
	"crypto/rand"
	"crypto/subtle"
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

	// Check saved hash length (24 bytes salt + 32 bytes digest)
	if len(decoded) != 24+32 {
		return errors.New("creds: Invalid password hash")
	}

	// salt = b64decode(reference)[:24]
    salt := decoded[:24]
	// saved_digest = b64decode(reference)[24:]
    saved_digest := decoded[24:]
	if subtle.ConstantTimeCompare(saved_digest, Digest(candidate, salt)) != 1 {
		return errors.New("creds: Password mismatch")
	}

	return nil
}

func Digest(password string, salt []byte) []byte {
	// Scrypt parameters are set to N = 8192, r = 8, p = 2 to ensure retrocompatibility with previous API server in Python
	// and are sufficient for the low security stakes of this application 
	dk, err := scrypt.Key([]byte(password), salt, 1<<13, 8, 2, 32)
	if err != nil {
		log.Fatal(err)
	}
	return dk
}