package core

func init() {
	Register(0xB7, wrap("TokensCreateSYN1000"))
	Register(0xB8, wrap("TokensAddStableReserve"))
	Register(0xB9, wrap("TokensSetStablePrice"))
	Register(0xBA, wrap("TokensStableReserveValue"))
}
