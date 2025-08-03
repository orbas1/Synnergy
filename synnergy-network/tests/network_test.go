package core_test

import (
	"context"
	"net"
	core "synnergy-network/core"
	"testing"
	"time"
)

//------------------------------------------------------------
// Test HandleNetworkMessage replication logic
//------------------------------------------------------------

func TestHandleNetworkMessage(t *testing.T) {
	// reset global store for test isolation
	replicatedMu.Lock()
	replicatedMessages = make(map[string][][]byte)
	replicatedMu.Unlock()

	msgs := []NetworkMessage{
		{Topic: "tx", Content: []byte{0x01}},
		{Topic: "tx", Content: []byte{0x02}},
		{Topic: "block", Content: []byte{0xFF}},
	}
	for _, m := range msgs {
		HandleNetworkMessage(m)
	}

	replicatedMu.Lock()
	defer replicatedMu.Unlock()

	if len(replicatedMessages["tx"]) != 2 {
		t.Fatalf("topic tx count=%d want 2", len(replicatedMessages["tx"]))
	}
	if len(replicatedMessages["block"]) != 1 {
		t.Fatalf("topic block count=%d want 1", len(replicatedMessages["block"]))
	}
}

//------------------------------------------------------------
// Dialer tests – success & failure paths (table‑driven)
//------------------------------------------------------------

func TestDialerDial(t *testing.T) {
	// start local TCP server for the success case
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen err %v", err)
	}
	defer ln.Close()

	addr := ln.Addr().String()

	cases := []struct {
		name    string
		target  string
		wantErr bool
	}{
		{"Success", addr, false},
		{"NoListener", "127.0.0.1:65000", true},
	}

	d := NewDialer(200*time.Millisecond, 0)

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()

			conn, err := d.Dial(ctx, tc.target)
			if (err != nil) != tc.wantErr {
				t.Fatalf("err=%v wantErr=%v", err, tc.wantErr)
			}
			if err == nil {
				conn.Close()
			}
		})
	}
}
