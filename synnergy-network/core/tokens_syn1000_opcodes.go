package core

func init() {
	Register(0xB7, wrap("Tokens_CreateSYN1000"))
	Register(0xB8, wrap("Tokens_AddStableReserve"))
	Register(0xB9, wrap("Tokens_SetStablePrice"))
	Register(0xBA, wrap("Tokens_StableReserveValue"))
}
