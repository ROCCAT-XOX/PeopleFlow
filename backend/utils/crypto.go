// backend/utils/crypto.go
package utils

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"io"
	"os"
)

// Verwende einen Umgebungsvariablenwert oder einen Default-Wert
var rawKey = []byte(getEncryptionKey())

// Verwende SHA-256, um einen Schlüssel mit konstanter Länge zu erhalten
var encryptionKey = sha256.Sum256(rawKey)

// getEncryptionKey liest den Verschlüsselungsschlüssel aus der Umgebungsvariable oder verwendet den Default-Wert
func getEncryptionKey() string {
	key := os.Getenv("PEOPLEFLOW_ENCRYPTION_KEY")
	if key == "" {
		return "PeopleFlow-Default-Secret-Key-Do-Not-Use-In-Production"
	}
	return key
}

// EncryptString verschlüsselt einen String
func EncryptString(plaintext string) (string, error) {
	block, err := aes.NewCipher(encryptionKey[:]) // Slice des SHA-256 Hashes verwenden
	if err != nil {
		return "", err
	}

	// Zufällige Bytes für den Initialisierungsvektor generieren
	ciphertext := make([]byte, aes.BlockSize+len(plaintext))
	iv := ciphertext[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return "", err
	}

	// Verschlüsselung durchführen
	stream := cipher.NewCFBEncrypter(block, iv)
	stream.XORKeyStream(ciphertext[aes.BlockSize:], []byte(plaintext))

	// In Base64 kodieren für die Speicherung
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// DecryptString entschlüsselt einen String
func DecryptString(encrypted string) (string, error) {
	ciphertext, err := base64.StdEncoding.DecodeString(encrypted)
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher(encryptionKey[:]) // Slice des SHA-256 Hashes verwenden
	if err != nil {
		return "", err
	}

	// Prüfen, ob der Ciphertext größer ist als der IV
	if len(ciphertext) < aes.BlockSize {
		return "", errors.New("ciphertext zu kurz")
	}

	// IV und Ciphertext extrahieren
	iv := ciphertext[:aes.BlockSize]
	ciphertext = ciphertext[aes.BlockSize:]

	// Entschlüsselung durchführen
	stream := cipher.NewCFBDecrypter(block, iv)
	stream.XORKeyStream(ciphertext, ciphertext)

	return string(ciphertext), nil
}
