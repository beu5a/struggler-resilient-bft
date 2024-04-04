package main

import (
	getty "github.com/apache/dubbo-getty"
)

// -------------------------------------------------Consensus Session Handlers ---------------------------------------------------------------------------------
type ConsensusSessionHandler struct {
	hub *NetworkingHub
}

func (h *ConsensusSessionHandler) OnOpen(session getty.Session) error {
	Logger.Infof("New consensus connection from %s", session.RemoteAddr())
	h.hub.consensusConnections = append(h.hub.consensusConnections, &session)
	return nil
}

func (h *ConsensusSessionHandler) OnError(session getty.Session, err error) {
	Logger.Errorf("Error on consensus connection from %s: %v", session.RemoteAddr(), err)
}

func (h *ConsensusSessionHandler) OnClose(session getty.Session) {
	Logger.Infof("Consensus connection from %s closed", session.RemoteAddr())
	// Remove the session from the hub or replace it will nil?
	h.hub.mu.Lock()
	defer h.hub.mu.Unlock()
	for i, s := range h.hub.consensusConnections {
		if s == &session {
			h.hub.consensusConnections[i] = nil
			break
		}
	}
}

func (h *ConsensusSessionHandler) OnMessage(session getty.Session, pkg interface{}) {
	// Debug:  Logger.Debugf("Received message from %s", session.RemoteAddr())
	msg := pkg.([]byte)
	h.hub.node.msgQueue <- msg
}

func (h *ConsensusSessionHandler) OnCron(session getty.Session) {}

// -------------------------------------------------State Transfer Session Handlers ---------------------------------------------------------------------------------
type StateTransferSessionHandler struct {
	hub *NetworkingHub
}

func (h *StateTransferSessionHandler) OnOpen(session getty.Session) error               { return nil }
func (h *StateTransferSessionHandler) OnError(session getty.Session, err error)         {}
func (h *StateTransferSessionHandler) OnClose(session getty.Session)                    {}
func (h *StateTransferSessionHandler) OnMessage(session getty.Session, pkg interface{}) {}
func (h *StateTransferSessionHandler) OnCron(session getty.Session)                     {}
