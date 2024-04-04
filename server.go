package main

import (
	"encoding/hex"
	"fmt"
	"math/rand"
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
		go s.embeddedClient()
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

func (s *Server) embeddedClient() {
	for i := 0; i < 1; i++ {
		msg := fmt.Sprintf("%d work to do!", i)
		req := Request{
			msg,
			hex.EncodeToString(generateDigest(msg)),
		}
		reqmsg := &RequestMsg{
			"solve",
			int(time.Now().Unix()),
			10,
			req,
		}
		sig := make([]byte, SignatureLength)
		_, err := rand.Read(sig)
		if err != nil {
			panic(err)
		}
		req_msg := ComposeMsg(hRequest, reqmsg, sig)
		s.node.msgQueue <- req_msg
	}

}
