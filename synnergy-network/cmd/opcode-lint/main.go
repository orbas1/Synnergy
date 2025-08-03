package main

import (
	"fmt"
	"log"

	core "synnergy-network/core"
)

func main() {
	ops := core.Catalogue()
	seenOps := make(map[core.Opcode]struct{})
	seenNames := make(map[string]struct{})
	for _, info := range ops {
		if _, ok := seenOps[info.Op]; ok {
			log.Fatalf("duplicate opcode 0x%06X", info.Op)
		}
		seenOps[info.Op] = struct{}{}
		if _, ok := seenNames[info.Name]; ok {
			log.Fatalf("duplicate opcode name %s", info.Name)
		}
		seenNames[info.Name] = struct{}{}
	}
	fmt.Printf("checked %d opcodes, no collisions detected\n", len(ops))
}
