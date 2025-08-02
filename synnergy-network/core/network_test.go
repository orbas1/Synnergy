package core

import "testing"

func TestHandleNetworkMessageReplication(t *testing.T) {
	ClearReplicatedMessages()
	msg := NetworkMessage{Topic: "test", Content: []byte("payload")}
	HandleNetworkMessage(msg)
	msgs := GetReplicatedMessages("test")
	if len(msgs) != 1 || string(msgs[0]) != "payload" {
		t.Fatalf("expected replicated payload, got %v", msgs)
	}
}
