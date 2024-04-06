package main

import (
	"fmt"
	"sync"

	getty "github.com/apache/dubbo-getty"
)

type NetworkingHub struct {
	node                 *Node // Use the fully qualified type name
	consensusConnections []*getty.Session
	//stateTransferConnections []*getty.Session
	clientConnections []*getty.Session
	mu                sync.Mutex // Protects connections
}

func NewNetworkingHub(node *Node) *NetworkingHub {
	hub := &NetworkingHub{
		node:                 node,
		consensusConnections: []*getty.Session{},
		mu:                   sync.Mutex{},
	}
	node.hub = hub

	return hub
}

func (h *NetworkingHub) ConnectToPeers() {
	h.listenForConsensusConnections()
	h.establishConsensusConnections()

	h.listenForStateTransferConnections()
	h.establishStateTransferConnections()

	h.listenForClientConnections()
}

func (h *NetworkingHub) establishConsensusConnections() {
	for _, peer := range h.node.knownNodes {
		if h.node.nodeID > peer.nodeID {
			// establish getty sessions
			address := fmt.Sprintf("%s:%d", peer.ip, peer.consensusPort)
			Logger.Infof("Establishing connection to %s", address)
			client := getty.NewTCPClient(
				getty.WithServerAddress(address),
				getty.WithConnectionNumber(1),  // Assuming one connection per peer for simplicity
				getty.WithReconnectInterval(3), //this should be setup in the system config file
			)

			client.RunEventLoop(func(session getty.Session) error {

				session.SetEventListener(
					&ConsensusSessionHandler{
						hub: h,
					},
				)
				session.SetPkgHandler(&DefaultPackageHandler{})

				return nil
			})
		}
	}
}

func (h *NetworkingHub) listenForConsensusConnections() {

	Logger.Infof("Listening for consensus connections on port %d", h.node.info.consensusPort)
	server := getty.NewTCPServer(
		getty.WithLocalAddress(fmt.Sprintf(":%d", h.node.info.consensusPort)))

	server.RunEventLoop(func(session getty.Session) error {
		/*
			_, ok := session.Conn().(*net.TCPConn)
			if !ok {
				Logger.Errorf("Connection Failed")
			}
		*/

		session.SetEventListener(
			&ConsensusSessionHandler{
				hub: h,
			},
		)
		session.SetPkgHandler(&DefaultPackageHandler{})

		return nil
	})
}

func (h *NetworkingHub) establishStateTransferConnections() {}
func (h *NetworkingHub) listenForStateTransferConnections() {}

// now we only support one client connection
func (h *NetworkingHub) listenForClientConnections() {
	server := getty.NewTCPServer(
		getty.WithLocalAddress(fmt.Sprintf(":%d", h.node.info.clientPort)))

	server.RunEventLoop(func(session getty.Session) error {
		session.SetEventListener(
			&ClientSessionHandler{
				hub: h,
			},
		)
		session.SetPkgHandler(&DefaultPackageHandler{})

		return nil
	})
}

//func (h *NetworkingHub) sendConsensusMsg(msg ConsensusMsg, dest int) {}
//func (h *NetworkingHub) broadcastConsensusMsg(msg ConsensusMsg)      {}

/*
func (h *NetworkingHub) sendStateTransferMsg(msg StateTransferMsg, dest int) {
	// Send the message one peer
}

func (h *NetworkingHub) broadcastStateTransferMsg(msg StateTransferMsg) {
	// Broadcast the message to all peers
}
*/

func (h *NetworkingHub) broadcast(bytes []byte) {
	// do we need to lock the hub here?
	h.mu.Lock()
	defer h.mu.Unlock()

	for _, session := range h.consensusConnections {
		s := *session
		s.Send(bytes)
	}
}

func (h *NetworkingHub) sendToClient(bytes []byte) {
	// do we need to lock the hub here?
	for _, session := range h.clientConnections {
		s := *session
		s.Send(bytes)
	}
}
