package core

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"

	"github.com/pion/webrtc/v4"
)

// RPCWebRTC bridges HTTP RPC calls with WebRTC data channels.
// It exposes minimal endpoints for transaction submission and
// forwards messages to connected peers via WebRTC.
type RPCWebRTC struct {
	ledger    *Ledger
	consensus *SynnergyConsensus

	srv   *http.Server
	peers map[string]*webRTCPeer
	mu    sync.Mutex
}

// webRTCPeer groups a peer connection with its reusable data channels.
type webRTCPeer struct {
	conn     *webrtc.PeerConnection
	channels map[string]*webrtc.DataChannel
}

// NewRPCWebRTC creates a new bridge instance.
func NewRPCWebRTC(ledger *Ledger, cons *SynnergyConsensus) *RPCWebRTC {
	return &RPCWebRTC{
		ledger:    ledger,
		consensus: cons,
		peers:     make(map[string]*webRTCPeer),
	}
}

// RPC_Serve starts an HTTP server exposing basic RPC endpoints.
func (r *RPCWebRTC) RPC_Serve(addr string) error {
	mux := http.NewServeMux()
	mux.HandleFunc("/tx", r.handleTx)
	r.srv = &http.Server{Addr: addr, Handler: mux}
	return r.srv.ListenAndServe()
}

// RPC_Close gracefully shuts down the RPC server and all peer connections.
func (r *RPCWebRTC) RPC_Close() error {
	if r.srv != nil {
		_ = r.srv.Shutdown(context.Background())
	}
	r.mu.Lock()
	for id, p := range r.peers {
		for _, dc := range p.channels {
			_ = dc.Close()
		}
		_ = p.conn.Close()
		delete(r.peers, id)
	}
	r.mu.Unlock()
	return nil
}

// RPC_ConnectPeer accepts a WebRTC offer and returns the answer SDP.
func (r *RPCWebRTC) RPC_ConnectPeer(offerSDP string) (string, error) {
	pc, err := webrtc.NewPeerConnection(webrtc.Configuration{})
	if err != nil {
		return "", err
	}
	offer := webrtc.SessionDescription{Type: webrtc.SDPTypeOffer, SDP: offerSDP}
	if err := pc.SetRemoteDescription(offer); err != nil {
		return "", err
	}
	answer, err := pc.CreateAnswer(nil)
	if err != nil {
		return "", err
	}
	if err := pc.SetLocalDescription(answer); err != nil {
		return "", err
	}
	r.mu.Lock()
	id := fmt.Sprintf("%p", pc)
	r.peers[id] = &webRTCPeer{conn: pc, channels: make(map[string]*webrtc.DataChannel)}
	r.mu.Unlock()
	return answer.SDP, nil
}

// RPC_Broadcast sends data to all connected peers via a data channel named topic.
func (r *RPCWebRTC) RPC_Broadcast(topic string, data []byte) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	for id, peer := range r.peers {
		dc, ok := peer.channels[topic]
		if !ok {
			var err error
			dc, err = peer.conn.CreateDataChannel(topic, nil)
			if err != nil {
				return fmt.Errorf("peer %s: %w", id, err)
			}
			peer.channels[topic] = dc
		}
		if err := dc.Send(data); err != nil {
			return fmt.Errorf("peer %s send: %w", id, err)
		}
	}
	return nil
}

func (r *RPCWebRTC) handleTx(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	var tx Transaction
	if err := json.NewDecoder(req.Body).Decode(&tx); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if r.ledger == nil {
		http.Error(w, "ledger not initialised", http.StatusInternalServerError)
		return
	}
	r.ledger.AddToPool(&tx)
	w.WriteHeader(http.StatusOK)
}
