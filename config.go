package main

import (
	"bufio"
	"crypto/ed25519"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
	"strconv"
	"strings"
)

func ReadHostsConfig(filePath string) ([]*NodeInfo, error) {
	var replicas []*NodeInfo

	file, err := os.Open(filePath)
	if err != nil {
		fmt.Println("Error opening file:", err)
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "#") || len(strings.TrimSpace(line)) == 0 {
			continue
		}
		parts := strings.Split(line, " ")

		id, err := strconv.Atoi(parts[0])
		if err != nil {
			fmt.Println("Error parsing hosts file, Invalid node ID: ", parts[0])
			os.Exit(1)
		}
		ip := parts[1]

		cPort, err := strconv.Atoi(parts[2])
		if err != nil {
			fmt.Println("Error parsing hosts file, Invalid client port: ", parts[2])
			os.Exit(1)
		}

		sPort, err := strconv.Atoi(parts[3])
		if err != nil {
			fmt.Println("Error parsing hosts file, Invalid consensus port: ", parts[3])
			os.Exit(1)
		}

		stPort, err := strconv.Atoi(parts[4])
		if err != nil {
			fmt.Println("Error parsing hosts file, Invalid state transfer port: ", parts[4])
			os.Exit(1)
		}

		replicas = append(replicas, &NodeInfo{
			nodeID:            id,
			ip:                ip,
			clientPort:        cPort,
			consensusPort:     sPort,
			stateTransferPort: stPort,
		})
	}
	// Print out the replicas to verify
	fmt.Println("Replicas:")
	for _, replica := range replicas {
		fmt.Printf("NodeID: %d, IP: %s, Ports: %d %d %d\n", replica.nodeID, replica.ip, replica.clientPort, replica.consensusPort, replica.stateTransferPort)
	}
	return replicas, nil
}

func ReadSystemConfig(filePath string) (map[string]int, error) {
	config := make(map[string]int)

	file, err := os.Open(filePath)
	if err != nil {
		fmt.Println("Error opening file:", err)
		return config, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "#") || len(strings.TrimSpace(line)) == 0 {
			// Skip comments and empty lines
			continue
		}

		parts := strings.Split(line, "=")
		key := strings.TrimSpace(parts[0])
		var value int
		_, err := fmt.Sscanf(parts[1], "%d", &value)
		if err != nil {
			fmt.Printf("Error parsing system config file, Invalid value for key %s: %s\n", key, parts[1])
			continue
		}
		config[key] = value

	}

	if err := scanner.Err(); err != nil {
		fmt.Println("Error reading system config file:", err)
	}

	// Print out the config map to verify
	fmt.Println("System Config:")
	for key, value := range config {
		fmt.Printf("%s = %d\n", key, value)
	}

	return config, err
}

func ReadPublicKeys(path string, nodeTable []*NodeInfo) {
	for _, nodeInfo := range nodeTable {
		pubKeyFile := fmt.Sprintf("%s/%d.pub", path, nodeInfo.nodeID)
		pubBytes, err := os.ReadFile(pubKeyFile)

		if err != nil {
			fmt.Println("Error reading public keys", err)
			return
		}

		decodePubKey := PublicKeyDecode(pubBytes)
		nodeInfo.pubKey = &decodePubKey
	}
}

func ReadPrivateKey(path string, nodeID int) *ed25519.PrivateKey {
	privKeyFile := fmt.Sprintf("%s/%d.priv", path, nodeID)
	privbytes, err := os.ReadFile(privKeyFile)
	if err != nil {
		fmt.Println("Error reading private key", err)
		return nil
	}
	decodePrivKey := PrivateKeyDecode(privbytes)

	// Convert ed25519.PrivateKey to *ecdsa.PrivateKey
	ecdsaPrivKey := &decodePrivKey

	return ecdsaPrivKey
}

func PrivateKeyDecode(pemEncoded []byte) ed25519.PrivateKey {
	blockPriv, _ := pem.Decode(pemEncoded)
	x509Encoded := blockPriv.Bytes

	bytes, _ := x509.ParsePKCS8PrivateKey(x509Encoded)
	privateKey, ok := bytes.(ed25519.PrivateKey)
	if !ok {
		fmt.Printf("Not an Ed25519 private key")
	}
	return privateKey
}

func PublicKeyDecode(pemEncoded []byte) ed25519.PublicKey {
	blockPub, _ := pem.Decode(pemEncoded)
	x509EncodedPub := blockPub.Bytes

	bytes, _ := x509.ParsePKIXPublicKey(x509EncodedPub)
	publicKey, ok := bytes.(ed25519.PublicKey)
	if !ok {
		fmt.Printf("Not an Ed25519 private key")
	}
	return publicKey
}
