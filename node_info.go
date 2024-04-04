package main

// move thsi to networking

import (
	"crypto/ed25519"
)

type NodeInfo struct {
	nodeID            int
	ip                string
	clientPort        int
	consensusPort     int
	stateTransferPort int
	pubKey            *ed25519.PublicKey
}

type ClientNodeInfo struct {
	nodeID int
	url    string
	pubkey *ed25519.PublicKey
}
