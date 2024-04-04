package main

import (
	"encoding/json"
	"fmt"
	"reflect"
)

type ConsensusPkg struct {
	header    ConsensusHeaderMsg
	payload   ConsensusMsg
	signature []byte
}

type ConsensusHeaderMsg string

const (
	consHPrePrepare ConsensusHeaderMsg = "PrePrepare"
	consHPrepare    ConsensusHeaderMsg = "Prepare"
	consHCommit     ConsensusHeaderMsg = "Commit"
)

type ConsensusMsg interface {
	String() string
}

// <<PRE-PREPARE,v,n,d>,m>
type _PrePrepareMsg struct {
	Request    RequestMsg `json:"request"`
	Digest     string     `json:"digest"`
	ViewID     int        `json:"viewID"`
	SequenceID int        `json:"sequenceID"`
}

func (msg _PrePrepareMsg) String() string {
	bmsg, _ := json.MarshalIndent(msg, "", "	")
	return string(bmsg) + "\n"
}

// <PREPARE, v, n, d, i>
type _PrepareMsg struct {
	Digest     string `json:"digest"`
	ViewID     int    `json:"viewID"`
	SequenceID int    `json:"sequenceID"`
	NodeID     int    `json:"nodeid"`
}

func (msg _PrepareMsg) String() string {
	bmsg, _ := json.MarshalIndent(msg, "", "	")
	return string(bmsg) + "\n"
}

// <COMMIT, v, n, d, i>
type _CommitMsg struct {
	Digest     string `json:"digest"`
	ViewID     int    `json:"viewID"`
	SequenceID int    `json:"sequenceID"`
	NodeID     int    `json:"nodeid"`
}

func (msg _CommitMsg) String() string {
	bmsg, _ := json.MarshalIndent(msg, "", "	")
	return string(bmsg) + "\n"
}

func ComposeConsensusPkg(header ConsensusHeaderMsg, payload ConsensusMsg, sig []byte) ConsensusPkg {
	return ConsensusPkg{
		header:    header,
		payload:   payload,
		signature: sig,
	}
}

func _ComposeMsg(header HeaderMsg, payload interface{}, sig []byte) []byte {
	var bpayload []byte
	var err error
	t := reflect.TypeOf(payload)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	switch t.Kind() {
	case reflect.Struct:
		bpayload, err = json.Marshal(payload)
		if err != nil {
			panic(err)
		}
	case reflect.Slice:
		bpayload = payload.([]byte)
	default:
		panic(fmt.Errorf("not support type"))
	}

	b := make([]byte, headerLength)
	for i, h := range []byte(header) {
		b[i] = h
	}
	res := make([]byte, headerLength+len(bpayload)+len(sig))
	copy(res[:headerLength], b)
	copy(res[headerLength:], bpayload)
	if len(sig) > 0 {
		copy(res[headerLength+len(bpayload):], sig)
	}
	return res
}

func _SplitMsg(bmsg []byte) (HeaderMsg, []byte, []byte) {
	var header HeaderMsg
	var payload []byte
	var signature []byte
	hbyte := bmsg[:headerLength]
	hhbyte := make([]byte, 0)
	for _, h := range hbyte {
		if h != byte(0) {
			hhbyte = append(hhbyte, h)
		}
	}
	header = HeaderMsg(hhbyte)
	switch header {
	case hRequest, hPrePrepare, hPrepare, hCommit:
		payload = bmsg[headerLength : len(bmsg)-256]
		signature = bmsg[len(bmsg)-256:]
	case hReply:
		payload = bmsg[headerLength:]
		signature = []byte{}
	}
	return header, payload, signature
}
