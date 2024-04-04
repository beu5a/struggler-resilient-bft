package main

import (
	"crypto/ed25519"

	"go.uber.org/zap"
)

var ClientNode *ClientNodeInfo
var Replicas []*NodeInfo
var SystemConfig map[string]int
var PrivateKey *ed25519.PrivateKey
var SignatureLength = ed25519.SignatureSize
var Logger *zap.SugaredLogger

func init() {

	hostsConfigFile := "./config/hosts.config"
	systemConfigFile := "./config/system.config"
	keysPath := "./config/keys"

	l, _ := zap.NewDevelopment()
	defer l.Sync()

	Logger = l.Sugar()
	Replicas, _ = ReadHostsConfig(hostsConfigFile)
	SystemConfig, _ = ReadSystemConfig(systemConfigFile)

	ReadPublicKeys(keysPath, Replicas)
}
