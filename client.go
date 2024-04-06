package main

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"sync"
	"time"
)

type Client struct {
	nodeId      int
	url         string
	knownNodes  []*NodeInfo
	connections []*net.Conn
	replyLog    map[int]*ReplyMsg
	throuput    int
	mutex       sync.Mutex
}

var ClientNode = ClientNodeInfo{
	0,
	"localhost:28080",
	nil,
}

func NewClient() *Client {
	client := &Client{
		ClientNode.nodeID,
		ClientNode.url,
		Replicas,
		[]*net.Conn{},
		make(map[int]*ReplyMsg),
		0,
		sync.Mutex{},
	}
	return client
}

func (c *Client) Start() {
	c.connectToNodes()
	time.Sleep(3 * time.Second)
	c.pollConnections()
	go c.sendManyRequest()

	for {
		select {
		case <-time.After(1 * time.Second):
			fmt.Printf("Current Throughput: %d ops/s \n", c.throuput)
			c.mutex.Lock()
			c.throuput = 0
			c.mutex.Unlock()

		}
	}
}

func (c *Client) sendManyRequest() {
	for {
		c.handleRequest()
		time.Sleep(time.Millisecond * 10)
	}
}

func (c *Client) handleRequest() {
	req_size := 1024
	msgb := make([]byte, req_size)
	_, err := rand.Read(msgb)
	if err != nil {
		fmt.Printf("Error in client Request : %v", err)
		panic(err)
	}

	msg := string(msgb)
	digest := string(generateDigest(msg))

	req := Request{
		msg,
		digest,
	}
	reqmsg := &RequestMsg{
		"Operation",
		int(time.Now().Unix()),
		10,
		req,
	}
	sig := make([]byte, SignatureLength)
	_, err = rand.Read(sig)
	if err != nil {
		fmt.Printf("Error in client Signature : %v", err)
		panic(err)
	}
	req_msg := ComposeMsg(hRequest, reqmsg, sig)
	c.sendRequest(req_msg)
}

func (c *Client) handleReply(payload []byte) {
	var replyMsg ReplyMsg
	fmt.Printf("Received reply: %s\n", payload)
	err := json.Unmarshal(payload, &replyMsg)
	if err != nil {
		fmt.Printf("error happened:%v", err)
		panic(err)
	}

	c.mutex.Lock()
	c.replyLog[replyMsg.NodeID] = &replyMsg
	rlen := len(c.replyLog)
	c.mutex.Unlock()
	if rlen >= c.countNeedReceiveMsgAmount() {
		c.throuput++
	}
}

func (c *Client) connectToNodes() {

	var url string
	for _, node := range c.knownNodes {
		url = fmt.Sprintf("%s:%d", node.ip, node.clientPort)
		conn, err := net.Dial("tcp", url)
		if err != nil {
			fmt.Printf("Error in client connecting to node: %v", err)
			panic(err)
		}
		fmt.Printf("Connected to %s\n", conn.RemoteAddr())
		c.connections = append(c.connections, &conn)
	}
}

func (c *Client) pollConnections() {
	for _, conn := range c.connections {
		go c.pollConnection(conn)
	}
}

func (c *Client) pollConnection(co *net.Conn) {
	//var buf []byte
	for {
		//_, err := (*co).Read(buf)
		buf, err := io.ReadAll(*co)
		if err != nil {
			fmt.Printf("Error in client reading msg: %v", err)
			panic(err)
		}
		c.handleReply(buf)
	}
}

func (c *Client) signMessage(msg interface{}) ([]byte, error) {
	// TODO
	return nil, nil
}

func (c *Client) sendRequest(msg []byte) {
	for _, conn := range c.connections {
		_, err := (*conn).Write(msg)
		if err != nil {
			fmt.Printf("Error in client sending msg: %v", err)
			panic(err)
		}
	}
}

func (c *Client) countTolerateFaultNode() int {
	return (len(c.knownNodes) - 1) / 3
}

func (c *Client) countNeedReceiveMsgAmount() int {
	f := c.countTolerateFaultNode()
	return f + 1
}
