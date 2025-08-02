package core_test

import (
	"testing"
	"time"

	core "synnergy-network/core"
)

func TestDAOProposalQuadraticVoting(t *testing.T) {
	core.SetStore(core.NewInMemoryStore())
	dir := t.TempDir()
	if err := core.InitLedger(dir); err != nil {
		t.Fatalf("ledger init failed: %v", err)
	}
	creator := core.Address{1}
	dao, err := core.CreateDAO("TestDAO", creator)
	if err != nil {
		t.Fatalf("create dao: %v", err)
	}
	led := core.CurrentLedger()
	led.TokenBalances[creator.String()+":"+core.Code] = 100
	voter := core.Address{2}
	led.TokenBalances[voter.String()+":"+core.Code] = 25
	if err := core.JoinDAO(dao.ID, voter); err != nil {
		t.Fatalf("join dao: %v", err)
	}
	p, err := core.CreateDAOProposal(dao.ID, creator, "test", time.Millisecond)
	if err != nil {
		t.Fatalf("create proposal: %v", err)
	}
	if err := core.VoteDAOProposal(p.ID, voter, 25, true); err != nil {
		t.Fatalf("vote: %v", err)
	}
	// duplicate vote should fail
	if err := core.VoteDAOProposal(p.ID, voter, 25, true); err == nil {
		t.Fatalf("expected duplicate vote error")
	}
	forW, againstW, err := core.TallyDAOProposal(p.ID)
	if err != nil {
		t.Fatalf("tally: %v", err)
	}
	if forW <= againstW {
		t.Fatalf("expected for > against")
	}
	time.Sleep(2 * time.Millisecond)
	if err := core.ExecuteDAOProposal(p.ID); err != nil {
		t.Fatalf("execute: %v", err)
	}
}
