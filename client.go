package main

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net"
	"sync"
	"time"
)

type Client struct {
	nodeId     int
	url        string
	knownNodes []*NodeInfo
	request    *RequestMsg
	replyLog   map[int]*ReplyMsg
	mutex      sync.Mutex
}

func NewClient() *Client {
	client := &Client{
		ClientNode.nodeID,
		ClientNode.url,
		[]*NodeInfo{},
		nil,
		make(map[int]*ReplyMsg),
		sync.Mutex{},
	}
	return client
}

func (c *Client) Start() {
	go c.sendManyRequest()
	ln, err := net.Listen("tcp", c.url)
	if err != nil {
		panic(err)
	}
	defer ln.Close()
	for {
		conn, err := ln.Accept()
		if err != nil {
			panic(err)
		}
		go c.handleConnection(conn)
	}
}

func (c *Client) handleConnection(conn net.Conn) {
	req, err := ioutil.ReadAll(conn)
	header, payload, _ := SplitMsg(req)
	if err != nil {
		panic(err)
	}
	switch header {
	case hReply:
		c.handleReply(payload)
	}
}

func (c *Client) sendManyRequest() {
	for {
		c.sendRequest()
		time.Sleep(time.Millisecond * 10)
	}
}

func (c *Client) sendRequest() {
	msg := fmt.Sprintf("%d work to do!", rand.Int())
	req := Request{
		msg,
		hex.EncodeToString(generateDigest(msg)),
	}
	reqmsg := &RequestMsg{
		"solve",
		int(time.Now().Unix()),
		c.nodeId,
		req,
	}

	/* TODO FIX THIS
	sig, err := c.signMessage(reqmsg)
	if err != nil {
		fmt.Printf("%v\n", err)
	}
	logBroadcastMsg(hRequest, reqmsg)
	send(ComposeMsg(hRequest, reqmsg, sig), c.findPrimaryNode().url)
	*/
	c.request = reqmsg
}

func (c *Client) handleReply(payload []byte) {
	var replyMsg ReplyMsg
	err := json.Unmarshal(payload, &replyMsg)
	if err != nil {
		fmt.Printf("error happened:%v", err)
		return
	}
	logHandleMsg(hReply, replyMsg, replyMsg.NodeID)
	c.mutex.Lock()
	c.replyLog[replyMsg.NodeID] = &replyMsg
	rlen := len(c.replyLog)
	c.mutex.Unlock()
	if rlen >= c.countNeedReceiveMsgAmount() {
		fmt.Println("request success!!")
	}
}

func (c *Client) signMessage(msg interface{}) ([]byte, error) {
	/* TODO FIX THIS
	sig, err := signMessage(msg, c.keypair.privkey)
	if err != nil {
		return nil, err
	}
	*/
	return nil, nil
}

func (c *Client) findPrimaryNode() *NodeInfo {
	nodeId := ViewID % len(c.knownNodes)
	for _, knownNode := range c.knownNodes {
		if knownNode.nodeID == nodeId {
			return knownNode
		}
	}
	return nil
}

func (c *Client) countTolerateFaultNode() int {
	return (len(c.knownNodes) - 1) / 3
}

func (c *Client) countNeedReceiveMsgAmount() int {
	f := c.countTolerateFaultNode()
	return f + 1
}
