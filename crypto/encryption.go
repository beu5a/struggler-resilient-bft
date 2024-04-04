package crypto

import (
	"crypto/rand"
	"crypto/rsa"
)

// bad file, rsa is not secure

// Encrypt encrypts the given plaintext using RSA public key.
func Encrypt(publicKey *rsa.PublicKey, plaintext []byte) ([]byte, error) {
	ciphertext, err := rsa.EncryptPKCS1v15(rand.Reader, publicKey, plaintext)
	if err != nil {
		return nil, err
	}

	return ciphertext, nil
}

// Decrypt decrypts the given ciphertext using RSA private key.
func Decrypt(privateKey *rsa.PrivateKey, ciphertext []byte) ([]byte, error) {
	plaintext, err := rsa.DecryptPKCS1v15(rand.Reader, privateKey, ciphertext)
	if err != nil {
		return nil, err
	}

	return plaintext, nil
}
