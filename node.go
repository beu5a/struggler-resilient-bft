package main

import (
	"crypto/ed25519"
	"encoding/json"
	"fmt"
	"sync"
	"time"
)

const ViewID = 0

type Node struct {
	nodeID     int
	info       *NodeInfo
	privateKey *ed25519.PrivateKey
	knownNodes []*NodeInfo
	clientNode *ClientNodeInfo
	sequenceID int
	View       int
	msgQueue   chan []byte
	hub        *NetworkingHub
	//stateTransferMsgQ chan []byte
	//clientMsgQ        chan []byte adding and removing messages from the queue will be handled by the hub
	msgLog      *MsgLog // https://github.com/uber-go/zap
	requestPool map[string]*RequestMsg
	mutex       sync.Mutex
}

type MsgLog struct {
	preprepareLog map[string]map[int]bool
	prepareLog    map[string]map[int]bool
	commitLog     map[string]map[int]bool
	replyLog      map[string]bool
}

func NewNode(nodeID int) *Node {
	return &Node{
		nodeID,
		Replicas[nodeID],
		PrivateKey,
		Replicas,
		ClientNode,
		0,
		ViewID,
		make(chan []byte),
		nil,
		&MsgLog{
			make(map[string]map[int]bool),
			make(map[string]map[int]bool),
			make(map[string]map[int]bool),
			make(map[string]bool),
		},
		make(map[string]*RequestMsg),
		sync.Mutex{},
	}
}

func (node *Node) getSequenceID() int {
	seq := node.sequenceID
	node.sequenceID++
	return seq
}

func (node *Node) Start() {
	go node.handleMsg()
}

// message handler function, create a handler for each Queue
func (node *Node) handleMsg() {
	for {
		msg := <-node.msgQueue
		header, payload, sign := SplitMsg(msg)
		switch header {
		case hRequest:
			node.handleRequest(payload, sign)
		case hPrePrepare:
			node.handlePrePrepare(payload, sign)
		case hPrepare:
			node.handlePrepare(payload, sign)
		case hCommit:
			node.handleCommit(payload, sign)
		}
	}
}

func (node *Node) handleRequest(payload []byte, sig []byte) {
	var request RequestMsg
	var prePrepareMsg PrePrepareMsg
	err := json.Unmarshal(payload, &request)
	if err != nil {
		Logger.Error("Error in Request Handling:%v", err)
		return
	}
	logHandleMsg(hRequest, request, request.ClientID)
	// verify request's digest
	vdig := verifyDigest(request.CRequest.Message, request.CRequest.Digest)
	if !vdig {
		Logger.Error("Verify digest failed in Request Handling")
		return
	}

	//TODO verify request's signature
	_ = sig
	if err != nil {
		Logger.Error("Verify request signature failed in Request Handling:%v", err)
		return
	}

	node.mutex.Lock()
	node.requestPool[request.CRequest.Digest] = &request
	seqID := node.getSequenceID()
	node.mutex.Unlock()

	prePrepareMsg = PrePrepareMsg{
		request,
		request.CRequest.Digest,
		ViewID,
		seqID,
	}
	//sign prePrepareMsg
	msgSig, err := node.signMessage(prePrepareMsg)
	if err != nil {
		Logger.Error("Sign prePrepareMsg failed in Request Handling:%v", err)
		return
	}
	msg := ComposeMsg(hPrePrepare, prePrepareMsg, msgSig)
	node.mutex.Lock()

	// put preprepare msg into log
	if node.msgLog.preprepareLog[prePrepareMsg.Digest] == nil {
		node.msgLog.preprepareLog[prePrepareMsg.Digest] = make(map[int]bool)
	}
	node.msgLog.preprepareLog[prePrepareMsg.Digest][node.nodeID] = true
	node.mutex.Unlock()
	logBroadcastMsg(hPrePrepare, prePrepareMsg)
	node.broadcast(msg)
}

// should be moved to consensus
func (node *Node) handlePrePrepare(payload []byte, sig []byte) {
	var prePrepareMsg PrePrepareMsg
	err := json.Unmarshal(payload, &prePrepareMsg)
	if err != nil {
		Logger.Error("Error happened in handle PrePrepare:%v", err)
		return
	}
	pnodeId := node.findPrimaryNode()
	logHandleMsg(hPrePrepare, prePrepareMsg, pnodeId)
	msgPubkey := node.findNodePubkey(pnodeId)
	if msgPubkey == nil {
		Logger.Error("Find node pubkey failed in handle PrePrepare\n")
		return
	}
	// verify msg's signature

	v := verifySignatrue(prePrepareMsg, sig, msgPubkey)
	if !v {
		Logger.Error("Verify signature failed in handle PrePrepare\n")
		return
	}

	// verify prePrepare's digest is equal to request's digest
	if prePrepareMsg.Digest != prePrepareMsg.Request.CRequest.Digest {
		Logger.Error("Verify digest failed in handle PrePrepare\n")
		return
	}
	node.mutex.Lock()
	node.requestPool[prePrepareMsg.Request.CRequest.Digest] = &prePrepareMsg.Request
	node.mutex.Unlock()
	err = node.verifyRequestDigest(prePrepareMsg.Digest)
	if err != nil {
		Logger.Error("Verify request digest failed in handle PrePrepare:%v\n", err)
		return
	}
	// put preprepare's msg into log
	node.mutex.Lock()
	if node.msgLog.preprepareLog[prePrepareMsg.Digest] == nil {
		node.msgLog.preprepareLog[prePrepareMsg.Digest] = make(map[int]bool)
	}
	node.msgLog.preprepareLog[prePrepareMsg.Digest][pnodeId] = true
	node.mutex.Unlock()
	prepareMsg := PrepareMsg{
		prePrepareMsg.Digest,
		ViewID,
		prePrepareMsg.SequenceID,
		node.nodeID,
	}
	//Logger.Debug("Create PrepareMsg:%v\n", prepareMsg)
	// sign prepare msg
	msgSig, err := signMessage(prepareMsg, node.privateKey)
	if err != nil {
		Logger.Error("Sign prepare msg failed in handle Preprepare:%v\n", err)
		return
	}
	sendMsg := ComposeMsg(hPrepare, prepareMsg, msgSig)
	node.mutex.Lock()
	// put prepare msg into log
	if node.msgLog.prepareLog[prepareMsg.Digest] == nil {
		node.msgLog.prepareLog[prepareMsg.Digest] = make(map[int]bool)
	}
	node.msgLog.prepareLog[prepareMsg.Digest][node.nodeID] = true
	node.mutex.Unlock()
	logBroadcastMsg(hPrepare, prepareMsg)
	node.broadcast(sendMsg)
}

func (node *Node) handlePrepare(payload []byte, sig []byte) {
	var prepareMsg PrepareMsg
	err := json.Unmarshal(payload, &prepareMsg)
	if err != nil {
		Logger.Error("Error happened in handle Prepare:%v", err)
		return
	}
	logHandleMsg(hPrepare, prepareMsg, prepareMsg.NodeID)
	pubkey := node.findNodePubkey(prepareMsg.NodeID)
	v := verifySignatrue(prepareMsg, sig, pubkey)
	if !v {
		Logger.Error("Verify signature failed in handle Prepare\n")
		return
	}

	// verify request's digest
	err = node.verifyRequestDigest(prepareMsg.Digest)
	if err != nil {
		Logger.Error("Verify request digest failed in handle Prepare:%v\n", err)
		return
	}
	// verify prepareMsg's digest is equal to preprepareMsg's digest
	pnodeId := node.findPrimaryNode()
	exist := node.msgLog.preprepareLog[prepareMsg.Digest][pnodeId]
	if !exist {
		Logger.Error("this digest's preprepare msg by %d not existed\n", pnodeId)
		return
	}
	// put prepareMsg into log
	node.mutex.Lock()
	if node.msgLog.prepareLog[prepareMsg.Digest] == nil {
		node.msgLog.prepareLog[prepareMsg.Digest] = make(map[int]bool)
	}
	node.msgLog.prepareLog[prepareMsg.Digest][prepareMsg.NodeID] = true
	node.mutex.Unlock()

	// if receive prepare msg >= 2f +1, then broadcast commit msg
	limit := node.countNeedReceiveMsgAmount()
	sum, err := node.findVerifiedPrepareMsgCount(prepareMsg.Digest)
	if err != nil {
		Logger.Error("Find Verified PrepareMsg Count Failed in handle Prepare:%v", err)
		return
	}
	if sum >= limit {
		// if already send commit msg, then do nothing
		node.mutex.Lock()
		exist := node.msgLog.commitLog[prepareMsg.Digest][node.nodeID]
		node.mutex.Unlock()
		if !exist {
			return
		}
		//send commit msg

		commitMsg := CommitMsg{
			prepareMsg.Digest,
			prepareMsg.ViewID,
			prepareMsg.SequenceID,
			node.nodeID,
		}
		sig, err := node.signMessage(commitMsg)
		if err != nil {
			fmt.Printf("sign message happened error:%v\n", err)
		}
		sendMsg := ComposeMsg(hCommit, commitMsg, sig)
		// put commit msg to log
		node.mutex.Lock()
		if node.msgLog.commitLog[commitMsg.Digest] == nil {
			node.msgLog.commitLog[commitMsg.Digest] = make(map[int]bool)
		}
		node.msgLog.commitLog[commitMsg.Digest][node.nodeID] = true
		node.mutex.Unlock()
		logBroadcastMsg(hCommit, commitMsg)
		node.broadcast(sendMsg)
	}
}

func (node *Node) handleCommit(payload []byte, sig []byte) {
	var commitMsg CommitMsg
	err := json.Unmarshal(payload, &commitMsg)
	if err != nil {
		Logger.Error("Error happened in handle Commit:%v", err)
	}
	logHandleMsg(hCommit, commitMsg, commitMsg.NodeID)
	//verify commitMsg's signature
	msgPubKey := node.findNodePubkey(commitMsg.NodeID)
	v := verifySignatrue(commitMsg, sig, msgPubKey)
	if !v {
		Logger.Error("Verify signature failed in handle Commit\n")
		return
	}

	err = node.verifyRequestDigest(commitMsg.Digest)
	if err != nil {
		Logger.Error("Verify request digest failed in handle Commit:%v\n", err)
		return
	}
	// put commitMsg into log
	node.mutex.Lock()
	if node.msgLog.commitLog[commitMsg.Digest] == nil {
		node.msgLog.commitLog[commitMsg.Digest] = make(map[int]bool)
	}
	node.msgLog.commitLog[commitMsg.Digest][commitMsg.NodeID] = true
	node.mutex.Unlock()
	// if receive commit msg >= 2f +1, then send reply msg to client
	limit := node.countNeedReceiveMsgAmount()

	sum, err := node.findVerifiedCommitMsgCount(commitMsg.Digest)
	if err != nil {
		Logger.Error("Find Verified CommitMsg Count Failed in handle Commit:%v", err)
		return
	}
	if sum >= limit {
		// if already send reply msg, then do nothing
		node.mutex.Lock()
		exist := node.msgLog.replyLog[commitMsg.Digest]
		node.mutex.Unlock()
		if exist {
			return
		}
		// send reply msg
		node.mutex.Lock()
		requestMsg := node.requestPool[commitMsg.Digest]
		node.mutex.Unlock()
		//fmt.Printf("operstion:%s  message:%s executed... \n", requestMsg.Operation, requestMsg.CRequest.Message)
		done := fmt.Sprintf("operstion:%s  message:%s done ", requestMsg.Operation, requestMsg.CRequest.Message)
		replyMsg := ReplyMsg{
			node.View,
			int(time.Now().Unix()),
			requestMsg.ClientID,
			node.nodeID,
			done,
		}
		logBroadcastMsg(hReply, replyMsg)
		//send(ComposeMsg(hReply, replyMsg, []byte{}), node.clientNode.url)
		node.mutex.Lock()
		node.msgLog.replyLog[commitMsg.Digest] = true
		node.mutex.Unlock()
	}
}

func (node *Node) verifyRequestDigest(digest string) error {
	node.mutex.Lock()
	_, ok := node.requestPool[digest]
	if !ok {
		node.mutex.Unlock()
		return fmt.Errorf("verify request digest failed")

	}
	node.mutex.Unlock()
	return nil
}

func (node *Node) findVerifiedPrepareMsgCount(digest string) (int, error) {
	sum := 0
	node.mutex.Lock()
	for _, exist := range node.msgLog.prepareLog[digest] {
		if exist {
			sum++
		}
	}
	node.mutex.Unlock()
	return sum, nil
}

func (node *Node) findVerifiedCommitMsgCount(digest string) (int, error) {
	sum := 0
	node.mutex.Lock()
	for _, exist := range node.msgLog.commitLog[digest] {

		if exist {
			sum++
		}
	}
	node.mutex.Unlock()
	return sum, nil
}

// should be moved to networking
func (node *Node) broadcast(data []byte) {
	node.hub.broadcast(data)
}

// do we need fast access to the public key of a node?
func (node *Node) findNodePubkey(nodeId int) *ed25519.PublicKey {
	for _, knownNode := range node.knownNodes {
		if knownNode.nodeID == nodeId {
			return knownNode.pubKey
		}
	}
	return nil
}

// Useless function ??
func (node *Node) signMessage(msg interface{}) ([]byte, error) {
	sig, err := signMessage(msg, node.privateKey)
	if err != nil {
		fmt.Printf("sign message happened error:%v\n", err)
		return nil, err
	}
	return sig, nil

}

// find leader
func (node *Node) findPrimaryNode() int {
	return ViewID % len(node.knownNodes)
}

// this is part of system config
func (node *Node) countTolerateFaultNode() int {
	return (len(node.knownNodes) - 1) / 3
}

// this is part of system config
func (node *Node) countNeedReceiveMsgAmount() int {
	f := node.countTolerateFaultNode()
	return 2*f + 1
}
