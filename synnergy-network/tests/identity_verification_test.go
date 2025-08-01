package core_test

import (
	"testing"

	core "synnergy-network/core"
)

func TestIdentityRegisterVerify(t *testing.T) {
	led, _ := core.NewInMemory()
	core.InitIdentityService(led)
	svc := core.Identity()
	addr := core.Address{0x01}

	if err := svc.Register(addr, []byte("ok")); err != nil {
		t.Fatalf("register: %v", err)
	}
	ok, err := svc.Verify(addr)
	if err != nil || !ok {
		t.Fatalf("verify: %v %v", ok, err)
	}
	list, err := svc.List()
	if err != nil || len(list) != 1 {
		t.Fatalf("list: %v %v", err, list)
	}
}
