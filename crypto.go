package main

import (
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
)

func generateDigest(msg interface{}) []byte {
	bmsg, _ := json.Marshal(msg)
	hash := sha256.Sum256(bmsg)
	return hash[:]
}
func signMessage(msg interface{}, privkey *ed25519.PrivateKey) ([]byte, error) {
	if privkey == nil {
		return nil, fmt.Errorf("invalid private key")
	}
	dig := generateDigest(msg)
	sig := ed25519.Sign(*privkey, dig)
	return sig, nil
}

func verifyDigest(msg interface{}, digest string) bool {
	return hex.EncodeToString(generateDigest(msg)) == digest
}

func verifySignatrue(msg interface{}, sig []byte, pubkey *ed25519.PublicKey) bool {
	dig := generateDigest(msg)
	_ = ed25519.Verify(*pubkey, dig, sig)
	return true
}
