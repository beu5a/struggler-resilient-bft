package main

import (
	"fmt"
	"log"
	"os"
	"srbft/network"
)

// Hard-coded for test.
var viewID = int64(10000000000)

func main() {
	var nodeTable []*network.NodeInfo

	if len(os.Args) < 2 {
		fmt.Println("Usage:", os.Args[0], "<nodeID> [node.list]")
		return
	}

	nodeID := os.Args[1]
	hostsConfigFile := "config/hosts.config"
	systemConfigFile := "config/system.config"
	keysPath := "config/keys"

	nodeTable, err := ReadHostsConfig(hostsConfigFile)
	AssertError(err)
	systemConfig, err := ReadSystemConfig(systemConfigFile)
	AssertError(err)
	_ = systemConfig

	// extract value from map

	// Load public key for each node -- make this part of config too
	ReadPublicKeys(keysPath, nodeTable)
	privateKey := ReadPrivateKey(keysPath, nodeID)

	server := network.NewServer(nodeID, nodeTable, viewID, privateKey)

	if server != nil {
		server.Start()
	}
}

func AssertError(err error) {
	if err == nil {
		return
	}

	log.Println(err)
	os.Exit(1)
}
