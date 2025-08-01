package core_test

import (
	"fmt"
	core "synnergy-network/core"
	"testing"
)

func TestDAOLifecycle(t *testing.T) {
	core.SetStore(core.NewInMemoryStore())
	creator := core.Address{1}
	dao, err := core.CreateDAO("TestDAO", creator)
	if err != nil {
		t.Fatalf("create err %v", err)
	}

	member := core.Address{2}
	if err := core.JoinDAO(dao.ID, member); err != nil {
		t.Fatalf("join err %v", err)
	}

	info, err := core.DAOInfo(dao.ID)
	if err != nil {
		t.Fatalf("info err %v", err)
	}
	key := fmt.Sprintf("%x", member[:])
	if !info.Members[key] {
		t.Fatalf("member not recorded")
	}

	if err := core.LeaveDAO(dao.ID, member); err != nil {
		t.Fatalf("leave err %v", err)
	}
	list, err := core.ListDAOs()
	if err != nil || len(list) == 0 {
		t.Fatalf("list err %v", err)
	}
}
