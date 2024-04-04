package networking

type Node struct {
	nodeID int
	//privateKey  *ecdsa.PrivateKey
	//knownNodes []*NodeInfo
	//clientNode *ClientNodeInfo
	sequenceID int
	View       int
	msgQueue   chan []byte
	//clientMsgQ        chan []byte adding and removing messages from the queue will be handled by the hub
	//stateTransferMsgQ chan []byte
	//msgLog      *MsgLog // https://github.com/uber-go/zap
	//requestPool map[string]*RequestMsg
	//mutex       sync.Mutex
}
