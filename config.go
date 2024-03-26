package main

import (
	"bufio"
	"crypto/ecdsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
	"srbft/network"
	"strings"
)

func ReadHostsConfig(filePath string) ([]*network.NodeInfo, error) {
	var replicas []*network.NodeInfo

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
			// Skip comments and empty lines
			continue
		}
		parts := strings.Split(line, " ")

		id := parts[0]
		ip := parts[1]
		port := parts[2]

		replicas = append(replicas, &network.NodeInfo{
			NodeID: id,
			Url:    ip + ":" + port,
		})
	}
	// Print out the replicas to verify
	fmt.Println("Replicas:")
	for _, replica := range replicas {
		fmt.Printf("NodeID: %s, Url: %s\n", replica.NodeID, replica.Url)
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

func PrivateKeyDecode(pemEncoded []byte) *ecdsa.PrivateKey {
	blockPriv, _ := pem.Decode(pemEncoded)
	x509Encoded := blockPriv.Bytes
	privateKey, _ := x509.ParseECPrivateKey(x509Encoded)

	return privateKey
}

func PublicKeyDecode(pemEncoded []byte) *ecdsa.PublicKey {
	blockPub, _ := pem.Decode(pemEncoded)
	x509EncodedPub := blockPub.Bytes
	genericPublicKey, _ := x509.ParsePKIXPublicKey(x509EncodedPub)
	publicKey := genericPublicKey.(*ecdsa.PublicKey)

	return publicKey
}

func ReadPublicKeys(path string, nodeTable []*network.NodeInfo) {
	for _, nodeInfo := range nodeTable {
		pubKeyFile := fmt.Sprintf("%s/%s.pub", path, nodeInfo.NodeID)
		pubBytes, err := os.ReadFile(pubKeyFile)

		if err != nil {
			fmt.Println("Error reading public keys", err)
			return
		}

		decodePubKey := PublicKeyDecode(pubBytes)
		nodeInfo.PubKey = decodePubKey
	}
}

func ReadPrivateKey(path string, nodeID string) *ecdsa.PrivateKey {
	privKeyFile := fmt.Sprintf("%s/%s.priv", path, nodeID)
	privbytes, err := os.ReadFile(privKeyFile)
	if err != nil {
		fmt.Println("Error reading private key", err)
		return nil
	}
	decodePrivKey := PrivateKeyDecode(privbytes)

	return decodePrivKey
}
