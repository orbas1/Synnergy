package core

import (
	"fmt"
	"testing"
)

func TestProposalLifecycle(t *testing.T) {
	// init in-memory store and ledger
	appStore = &InMemoryStore{data: make(map[string][]byte)}
	ledger = &Ledger{TokenBalances: map[string]uint64{fmt.Sprintf("%x", []byte{1}): 1}}

	creator := Address{1}
	prop := &GovProposal{Creator: creator, Description: "test"}
	if err := SubmitProposal(prop); err != nil {
		t.Fatalf("SubmitProposal err: %v", err)
	}

	// fetch by ID
	got, err := GetProposal(prop.ID)
	if err != nil {
		t.Fatalf("GetProposal err: %v", err)
	}
	if got.Description != "test" {
		t.Fatalf("unexpected description %q", got.Description)
	}

	list, err := ListProposals()
	if err != nil {
		t.Fatalf("ListProposals err: %v", err)
	}
	if len(list) != 1 {
		t.Fatalf("expected 1 proposal, got %d", len(list))
	}
}
