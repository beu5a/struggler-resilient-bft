package main

import (
	"encoding/hex"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type Server struct {
	node *Node
	hub  *NetworkingHub
}

func NewServer(nodeId int) *Server {
	// A server has a node and a communication hub
	PrivateKey = ReadPrivateKey("./config/keys", nodeId)
	newNode := NewNode(nodeId)
	newHub := NewNetworkingHub(newNode)

	server := &Server{
		newNode,
		newHub,
	}
	return server
}

func (s *Server) Start() {

	s.node.Start()

	//sleep a duration proportional to the node id
	Logger.Info("Sleeping before connecting to peers...")
	deltaT := time.Duration(3*(s.node.nodeID+1)) * time.Second
	time.Sleep(deltaT)

	Logger.Info("Connecting to peers...")
	s.hub.ConnectToPeers()

	sigs := make(chan os.Signal, 1)
	// Create a channel to signal the main program to stop.
	done := make(chan bool, 1)

	// Register the channel to receive notifications for specific signals, namely Interrupt and SIGTERM.
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	// wait until all the nodes are connected
	time.Sleep(10 * time.Second)

	if s.node.nodeID == 0 {
		go s.testClient()
	}

	// Start a goroutine that blocks on waiting for signals.
	// Once a signal is received, it prints a message and sends a value to the 'done' channel to indicate the program should stop.
	go func() {
		sig := <-sigs
		fmt.Println()
		fmt.Println(sig, "Signal received. Exiting...")
		done <- true
	}()

	fmt.Println("Press CTRL+C to stop the replica...")

	// Block the main goroutine until a value is received on the 'done' channel.
	<-done
	fmt.Println("Program stopped.")
}

func (s *Server) testClient() {
	// Generate 100 requests
	req_size := 1024
	msgb := make([]byte, req_size)
	for i := 0; i < 1024; i++ {
		msgb[i] = "a"[0]
	}

	msg := "Hello, World!"
	digest := hex.EncodeToString(generateDigest(msg))

	req := Request{
		msg,
		digest,
	}
	reqmsg := &RequestMsg{
		"Operation",
		0,
		10,
		req,
	}
	sig, _ := signMessage(reqmsg, PrivateKey)
	//req_msg := ComposeMsg(hRequest, reqmsg, sig)
	for i := 0; i < 4; i++ {
		reqmsg.Timestamp = int(time.Now().Unix())
		req_msg := ComposeMsg(hRequest, reqmsg, sig)
		s.node.msgQueue <- req_msg
		//time.Sleep(1000 * time.Microsecond)

	}
}
