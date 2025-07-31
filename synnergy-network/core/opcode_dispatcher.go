// SPDX-License-Identifier: BUSL-1.1
//
// Synnergy Network – Core ▸ Opcode Dispatcher
// -------------------------------------------
//
//   - Every high-level function in the protocol is assigned a UNIQUE 24-bit
//     opcode:  0xCCNNNN  →  CC = category (1 byte), NNNN = ordinal (2 bytes).
//
//   - The dispatcher maps opcodes ➝ concrete handlers and enforces gas-pricing
//     through core.GasCost() before execution.
//
//   - All collisions or missing handlers are FATAL at start-up; nothing slips
//     into production unnoticed.
//
//     ────────────────────────────────────────────────────────────────────────────
//     AUTOMATED SECTION
//     -----------------
//     The table below is **generated** by `go generate ./...` (see the generator
//     in `cmd/genopcodes`).  Edit ONLY if you know what you’re doing; otherwise
//     add new function names to `generator/input/functions.yml` and re-run
//     `go generate`.  The generator guarantees deterministic, collision-free
//     opcodes and keeps this file lint-clean.
//
//     Format per line:
//     <FunctionName>  =  <24-bit-binary>  =  <HexOpcode>
//
//     NB: Tabs are significant – tools rely on them when parsing for audits.
package core

import (
	"encoding/hex"
	"fmt"
	"log"
	"sync"
)

// ────────────────────────────────────────────────────────────────────────────
// Context & Dispatcher glue
// ────────────────────────────────────────────────────────────────────────────

// Context is provided by the VM; it gives opcode handlers controlled access
// to message meta-data, state-DB, gas-meter, logger, etc.
type Context interface {
	Call(string) error // unified façade (ledger/consensus/VM)
	Gas(uint64) error  // deducts gas or returns an error if exhausted
}

// Opcode is a 24-bit, deterministic instruction identifier.
type Opcode uint32

// OpcodeFunc is the concrete implementation invoked by the VM.
type OpcodeFunc func(ctx Context) error

// opcodeTable holds the runtime mapping (populated once in init()).
var (
	opcodeTable = make(map[Opcode]OpcodeFunc, 1024)
	nameToOp    = make(map[string]Opcode, 1024)
	mu          sync.RWMutex
)

// Register binds an opcode to its function handler.
// It panics on duplicates – this should never happen in CI-tested builds.
func Register(op Opcode, fn OpcodeFunc) {
	mu.Lock()
	defer mu.Unlock()
	if _, exists := opcodeTable[op]; exists {
		log.Panicf("[OPCODES] collision: 0x%06X already registered", op)
	}
	opcodeTable[op] = fn
}

// Dispatch is called by the VM executor for every instruction.
func Dispatch(ctx Context, op Opcode) error {
	mu.RLock()
	fn, ok := opcodeTable[op]
	mu.RUnlock()

	if !ok {
		return fmt.Errorf("unknown opcode 0x%06X", op)
	}
	// Pre-charge gas (base only – dynamic part inside fn)
	if err := ctx.Gas(GasCost(Opcode(op))); err != nil {
		return err
	}
	return fn(ctx)
}

// helper returns a closure that delegates the call to Context.Call(<name>).
func wrap(name string) OpcodeFunc {
	return func(ctx Context) error { return ctx.Call(name) }
}

// ────────────────────────────────────────────────────────────────────────────
// Opcode Catalogue  (AUTO-GENERATED – DO NOT EDIT BY HAND)
// ────────────────────────────────────────────────────────────────────────────
//
// Category map:
//
//	0x01 AI                     0x0F Liquidity
//	0x02 AMM                    0x10 Loanpool
//	0x03 Authority              0x11 Network
//	0x04 Charity                0x12 Replication
//	0x05 Coin                   0x13 Rollups
//	0x06 Compliance             0x14 Security
//	0x07 Consensus              0x15 Sharding
//	0x08 Contracts              0x16 Sidechains
//	0x09 CrossChain             0x17 StateChannel
//	0x0A Data                   0x18 Storage
//	0x0B FaultTolerance         0x19 Tokens
//	0x0C Governance             0x1A Transactions
//	0x0D GreenTech              0x1B Utilities
//	0x0E Ledger                 0x1C VirtualMachine
//	                            0x1D Wallet
//
// Each binary code is shown as a 24-bit big-endian string.
var catalogue = []struct {
	name string
	op   Opcode
}{
	// AI (0x01)
	{"InitAI", 0x010001},         // 00000001 00000000 00000001
	{"AI", 0x010002},             // 00000001 00000000 00000010
	{"PredictAnomaly", 0x010003}, // 00000001 00000000 00000011
	{"OptimizeFees", 0x010004},
	{"PublishModel", 0x010005},
	{"FetchModel", 0x010006},
	{"ListModel", 0x010007},
	{"ValidateKYC_AI", 0x010008},
	{"BuyModel", 0x010009},
	{"RentModel", 0x01000A},
	{"ReleaseEscrow", 0x01000B},
	{"PredictVolume", 0x01000C},

	// AMM (0x02)
	{"SwapExactIn", 0x020001},
	{"AMM_AddLiquidity", 0x020002},
	{"AMM_RemoveLiquidity", 0x020003},
	{"Quote", 0x020004},
	{"AllPairs", 0x020005},
	{"InitPoolsFromFile", 0x020006},

	// Authority (0x03)
	{"NewAuthoritySet", 0x030001},
	{"RecordVote", 0x030002},
	{"RegisterCandidate", 0x030003},
	{"RandomElectorate", 0x030004},
	{"IsAuthority", 0x030005},
	{"GetAuthority", 0x030006},
	{"ListAuthorities", 0x030007},
	{"DeregisterAuthority", 0x030008},

	// Charity (0x04)
	{"NewCharityPool", 0x040001},
	{"Charity_Deposit", 0x040002},
	{"Charity_Register", 0x040003},
	{"Charity_Vote", 0x040004},
	{"Charity_Tick", 0x040005},
	{"Charity_GetRegistration", 0x040006},
	{"Charity_Winners", 0x040007},

	// Coin (0x05)
	{"NewCoin", 0x050001},
	{"Coin_Mint", 0x050002},
	{"Coin_TotalSupply", 0x050003},
	{"Coin_BalanceOf", 0x050004},
	{"Coin_Transfer", 0x050005},
	{"Coin_Burn", 0x050006},

	// Compliance (0x06)
	{"InitCompliance", 0x060001},
	{"Compliance_ValidateKYC", 0x060002},
	{"EraseData", 0x060003},
	{"RecordFraudSignal", 0x060004},
	{"Compliance_LogAudit", 0x060005},
	{"Compliance_AuditTrail", 0x060006},
	{"Compliance_MonitorTx", 0x060007},
	{"Compliance_VerifyZKP", 0x060008},

	// Consensus (0x07)
	{"Pick", 0x070001},
	{"Consensus_Broadcast", 0x070002},
	{"Consensus_Subscribe", 0x070003},
	{"Consensus_Sign", 0x070004},
	{"Consensus_Verify", 0x070005},
	{"ValidatorPubKey", 0x070006},
	{"StakeOf", 0x070007},
	{"LoanPoolAddress", 0x070008},
	{"Consensus_Hash", 0x070009},
	{"SerializeWithoutNonce", 0x07000A},
	{"NewConsensus", 0x07000B},
	{"Consensus_Start", 0x07000C},
	{"ProposeSubBlock", 0x07000D},
	{"ValidatePoH", 0x07000E},
	{"SealMainBlockPOW", 0x07000F},
	{"DistributeRewards", 0x070010},
	{"CalculateWeights", 0x070011},
	{"ComputeThreshold", 0x070012},

	// Contracts (0x08)
	{"InitContracts", 0x080001},
	{"CompileWASM", 0x080002},
	{"Invoke", 0x080003},

	// Cross-Chain (0x09)
	{"RegisterBridge", 0x090001},
	{"AssertRelayer", 0x090002},
	{"Iterator", 0x090003},
	{"LockAndMint", 0x090004},
	{"BurnAndRelease", 0x090005},
	{"GetBridge", 0x090006},

	// Data (0x0A)
	{"RegisterNode", 0x0A0001},
	{"UploadAsset", 0x0A0002},
	{"Data_Pin", 0x0A0003},
	{"Data_Retrieve", 0x0A0004},
	{"RetrieveAsset", 0x0A0005},
	{"RegisterOracle", 0x0A0006},
	{"PushFeed", 0x0A0007},
	{"QueryOracle", 0x0A0008},

	// Fault-Tolerance (0x0B)
	{"NewHealthChecker", 0x0B0001},
	{"AddPeer", 0x0B0002},
	{"RemovePeer", 0x0B0003},
	{"Snapshot", 0x0B0004},
	{"Recon", 0x0B0005},
	{"Ping", 0x0B0006},
	{"SendPing", 0x0B0007},
	{"AwaitPong", 0x0B0008},

	// Governance (0x0C)
	{"UpdateParam", 0x0C0001},
	{"ProposeChange", 0x0C0002},
	{"VoteChange", 0x0C0003},
	{"EnactChange", 0x0C0004},
	{"SubmitProposal", 0x0C0005},
	{"BalanceOfAsset", 0x0C0006},
	{"CastVote", 0x0C0007},
	{"ExecuteProposal", 0x0C0008},
	{"GetProposal", 0x0C0009},
	{"ListProposals", 0x0C000A},

	// GreenTech (0x0D)
	{"InitGreenTech", 0x0D0001},
	{"Green", 0x0D0002},
	{"RecordUsage", 0x0D0003},
	{"RecordOffset", 0x0D0004},
	{"Certify", 0x0D0005},
	{"CertificateOf", 0x0D0006},
	{"ShouldThrottle", 0x0D0007},

	// Ledger (0x0E)
	{"NewLedger", 0x0E0001},
	{"GetPendingSubBlocks", 0x0E0002},
	{"LastBlockHash", 0x0E0003},
	{"AppendBlock", 0x0E0004},
	{"MintBig", 0x0E0005},
	{"EmitApproval", 0x0E0006},
	{"EmitTransfer", 0x0E0007},
	{"DeductGas", 0x0E0008},
	{"WithinBlock", 0x0E0009},
	{"IsIDTokenHolder", 0x0E000A},
	{"TokenBalance", 0x0E000B},
	{"AddBlock", 0x0E000C},
	{"GetBlock", 0x0E000D},
	{"GetUTXO", 0x0E000E},
	{"AddToPool", 0x0E000F},
	{"ListPool", 0x0E0010},
	{"GetContract", 0x0E0011},
	{"Ledger_BalanceOf", 0x0E0012},
	{"Ledger_Snapshot", 0x0E0013},
	{"MintToken", 0x0E0014},
	{"LastSubBlockHeight", 0x0E0015},
	{"LastBlockHeight", 0x0E0016},
	{"RecordPoSVote", 0x0E0017},
	{"AppendSubBlock", 0x0E0018},
	{"Ledger_Transfer", 0x0E0019},
	{"Ledger_Mint", 0x0E001A},
	{"Ledger_Burn", 0x0E001B},

	// Liquidity (0x0F)
	{"InitAMM", 0x0F0001},
	{"Manager", 0x0F0002},
	{"CreatePool", 0x0F0003},
	{"Liquidity_AddLiquidity", 0x0F0004},
	{"Liquidity_Swap", 0x0F0005},
	{"Liquidity_RemoveLiquidity", 0x0F0006},

	// Loanpool (0x10)
	{"Loanpool_RandomElectorate", 0x100001},
	{"Loanpool_IsAuthority", 0x100002},
	{"Loanpool_init", 0x100003},
	{"NewLoanPool", 0x100004},
	{"Loanpool_Submit", 0x100005},
	{"Loanpool_Vote", 0x100006},
	{"Disburse", 0x100007},
	{"Loanpool_Tick", 0x100008},

	// Network (0x11)
	{"NewNode", 0x110001},
	{"HandlePeerFound", 0x110002},
	{"DialSeed", 0x110003},
	{"Network_Broadcast", 0x110004},
	{"Network_Subscribe", 0x110005},
	{"ListenAndServe", 0x110006},
	{"Close", 0x110007},
	{"Peers", 0x110008},
	{"NewDialer", 0x110009},
	{"Dial", 0x11000A},

	// Replication (0x12)
	{"NewReplicator", 0x120001},
	{"ReplicateBlock", 0x120002},
	{"Replication_Hash", 0x120003},
	{"RequestMissing", 0x120004},
	{"Replication_Start", 0x120005},
	{"Stop", 0x120006},

	// Rollups (0x13)
	{"NewAggregator", 0x130001},
	{"SubmitBatch", 0x130002},
	{"SubmitFraudProof", 0x130003},
	{"FinalizeBatch", 0x130004},

	// Security (0x14)
	{"Security_Sign", 0x140001},
	{"Security_Verify", 0x140002},
	{"AggregateBLSSigs", 0x140003},
	{"VerifyAggregated", 0x140004},
	{"CombineShares", 0x140005},
	{"ComputeMerkleRoot", 0x140006},
	{"Encrypt", 0x140007},
	{"Decrypt", 0x140008},
	{"NewTLSConfig", 0x140009},

	// Sharding (0x15)
	{"NewShardCoordinator", 0x150001},
	{"SetLeader", 0x150002},
	{"Leader", 0x150003},
	{"SubmitCrossShard", 0x150004},
	{"Sharding_Broadcast", 0x150005},
	{"Send", 0x150006},
	{"PullReceipts", 0x150007},
	{"Reshard", 0x150008},

	// Sidechains (0x16)
	{"InitSidechains", 0x160001},
	{"Sidechains", 0x160002},
	{"Sidechains_Register", 0x160003},
	{"SubmitHeader", 0x160004},
	{"Sidechains_Deposit", 0x160005},
	{"VerifyWithdraw", 0x160006},
	{"VerifyAggregateSig", 0x160007},
	{"VerifyMerkleProof", 0x160008},

	// StateChannel (0x17)
	{"InitStateChannels", 0x170001},
	{"Channels", 0x170002},
	{"OpenChannel", 0x170003},
	{"VerifyECDSASignature", 0x170004},
	{"InitiateClose", 0x170005},
	{"Challenge", 0x170006},
	{"Finalize", 0x170007},

	// Storage (0x18)
	{"NewStorage", 0x180001},
	{"Storage_Pin", 0x180002},
	{"Storage_Retrieve", 0x180003},
	{"CreateListing", 0x180004},
	{"Exists", 0x180005},
	{"OpenDeal", 0x180006},
	{"Storage_Create", 0x180007},
	{"CloseDeal", 0x180008},
	{"Release", 0x180009},

	// Tokens (0x19)
	{"ID", 0x190001},
	{"Meta", 0x190002},
	{"Tokens_BalanceOf", 0x190003},
	{"Tokens_Transfer", 0x190004},
	{"Allowance", 0x190005},
	{"Tokens_Approve", 0x190006},
	{"Tokens_Mint", 0x190007},
	{"Tokens_Burn", 0x190008},
	{"Add", 0x190009},
	{"Sub", 0x19000A},
	{"Get", 0x19000B},
	{"approve_lower", 0x19000C},
	{"transfer_lower", 0x19000D},
	{"Calculate", 0x19000E},
	{"RegisterToken", 0x19000F},
	{"Tokens_Create", 0x190010},
	{"NewBalanceTable", 0x190011},
	{"Set", 0x190012},
	{"RefundGas", 0x190013},
	{"PopUint32", 0x190014},
	{"PopAddress", 0x190015},
	{"PopUint64", 0x190016},
	{"PushBool", 0x190017},
	{"Push", 0x190018},
	{"Len_Tokens", 0x190019},

	// Transactions (0x1A)
	{"Tx_Sign", 0x1A0001},
	{"VerifySig", 0x1A0002},
	{"ValidateTx", 0x1A0003},
	{"NewTxPool", 0x1A0004},

	// Utilities (0x1B) – EVM-compatible arithmetic & crypto
	{"Short", 0x1B0001},
	{"BytesToAddress", 0x1B0002},
	{"Pop", 0x1B0003},
	{"opADD", 0x1B0004},
	{"opMUL", 0x1B0005},
	{"opSUB", 0x1B0006},
	{"OpDIV", 0x1B0007},
	{"opSDIV", 0x1B0008},
	{"opMOD", 0x1B0009},
	{"opSMOD", 0x1B000A},
	{"opADDMOD", 0x1B000B},
	{"opMULMOD", 0x1B000C},
	{"opEXP", 0x1B000D},
	{"opSIGNEXTEND", 0x1B000E},
	{"opLT", 0x1B000F},
	{"opGT", 0x1B0010},
	{"opSLT", 0x1B0011},
	{"opSGT", 0x1B0012},
	{"opEQ", 0x1B0013},
	{"opISZERO", 0x1B0014},
	{"opAND", 0x1B0015},
	{"opOR", 0x1B0016},
	{"opXOR", 0x1B0017},
	{"opNOT", 0x1B0018},
	{"opBYTE", 0x1B0019},
	{"opSHL", 0x1B001A},
	{"opSHR", 0x1B001B},
	{"opSAR", 0x1B001C},
	{"opECRECOVER", 0x1B001D},
	{"opEXTCODESIZE", 0x1B001E},
	{"opEXTCODECOPY", 0x1B001F},
	{"opEXTCODEHASH", 0x1B0020},
	{"opRETURNDATASIZE", 0x1B0021},
	{"opRETURNDATACOPY", 0x1B0022},
	{"opMLOAD", 0x1B0023},
	{"opMSTORE", 0x1B0024},
	{"opMSTORE8", 0x1B0025},
	{"opCALLDATALOAD", 0x1B0026},
	{"opCALLDATASIZE", 0x1B0027},
	{"opCALLDATACOPY", 0x1B0028},
	{"opCODESIZE", 0x1B0029},
	{"opCODECOPY", 0x1B002A},
	{"opJUMP", 0x1B002B},
	{"opJUMPI", 0x1B002C},
	{"opPC", 0x1B002D},
	{"opMSIZE", 0x1B002E},
	{"opGAS", 0x1B002F},
	{"opJUMPDEST", 0x1B0030},
	{"opSHA256", 0x1B0031},
	{"opKECCAK256", 0x1B0032},
	{"opRIPEMD160", 0x1B0033},
	{"opBLAKE2B256", 0x1B0034},
	{"opADDRESS", 0x1B0035},
	{"opCALLER", 0x1B0036},
	{"opORIGIN", 0x1B0037},
	{"opCALLVALUE", 0x1B0038},
	{"opGASPRICE", 0x1B0039},
	{"opNUMBER", 0x1B003A},
	{"opTIMESTAMP", 0x1B003B},
	{"opDIFFICULTY", 0x1B003C},
	{"opGASLIMIT", 0x1B003D},
	{"opCHAINID", 0x1B003E},
	{"opBLOCKHASH", 0x1B003F},
	{"opBALANCE", 0x1B0040},
	{"opSELFBALANCE", 0x1B0041},
	{"opLOG0", 0x1B0042},
	{"opLOG1", 0x1B0043},
	{"opLOG2", 0x1B0044},
	{"opLOG3", 0x1B0045},
	{"opLOG4", 0x1B0046},
	{"logN", 0x1B0047},
	{"opCREATE", 0x1B0048},
	{"opCALL", 0x1B0049},
	{"opCALLCODE", 0x1B004A},
	{"opDELEGATECALL", 0x1B004B},
	{"opSTATICCALL", 0x1B004C},
	{"opRETURN", 0x1B004D},
	{"opREVERT", 0x1B004E},
	{"opSTOP", 0x1B004F},
	{"opSELFDESTRUCT", 0x1B0050},
	{"Utilities_Transfer", 0x1B0051},
	{"Utilities_Mint", 0x1B0052},
	{"Utilities_Burn", 0x1B0053},

	// Virtual Machine (0x1C)
	{"VM_Burn", 0x1C0001},
	{"BurnLP", 0x1C0002},
	{"MintLP", 0x1C0003},
	{"NewInMemory", 0x1C0004},
	{"CallCode", 0x1C0005},
	{"CallContract", 0x1C0006},
	{"StaticCall", 0x1C0007},
	{"GetBalance", 0x1C0008},
	{"GetTokenBalance", 0x1C0009},
	{"SetTokenBalance", 0x1C000A},
	{"GetTokenSupply", 0x1C000B},
	{"SetBalance", 0x1C000C},
	{"DelegateCall", 0x1C000D},
	{"GetToken", 0x1C000E},
	{"NewMemory", 0x1C000F},
	{"VM_Read", 0x1C0010},
	{"VM_Write", 0x1C0011},
	{"VM_Len", 0x1C0012},
	{"VM_Call", 0x1C0013},
	{"SelectVM", 0x1C0014},
	{"CreateContract", 0x1C0015},
	{"VM_GetContract", 0x1C0016},
	{"AddLog", 0x1C0017},
	{"GetCode", 0x1C0018},
	{"GetCodeHash", 0x1C0019},
	{"MintToken_VM", 0x1C001A},
	{"VM_Transfer", 0x1C001B},
	{"PrefixIterator", 0x1C001C},
	{"Snapshot_VM", 0x1C001D},
	{"NonceOf", 0x1C001E},
	{"IsIDTokenHolder_VM", 0x1C001F},
	{"GetState", 0x1C0020},
	{"SetState", 0x1C0021},
	{"HasState", 0x1C0022},
	{"DeleteState", 0x1C0023},
	{"BalanceOf_VM", 0x1C0024},
	{"NewGasMeter", 0x1C0025},
	{"SelfDestruct", 0x1C0026},
	{"Remaining", 0x1C0027},
	{"Consume", 0x1C0028},
	{"Execute", 0x1C0029},
	{"NewSuperLightVM", 0x1C002A},
	{"NewLightVM", 0x1C002B},
	{"NewHeavyVM", 0x1C002C},
	{"ExecuteSuperLight", 0x1C002D},
	{"ExecuteLight", 0x1C002E},
	{"ExecuteHeavy", 0x1C002F},

	// Wallet (0x1D)
	{"NewRandomWallet", 0x1D0001},
	{"WalletFromMnemonic", 0x1D0002},
	{"NewHDWalletFromSeed", 0x1D0003},
	{"PrivateKey", 0x1D0004},
	{"NewAddress", 0x1D0005},
	{"SignTx", 0x1D0006},
}

// init wires the catalogue into the live dispatcher.
func init() {
	for _, entry := range catalogue {
		nameToOp[entry.name] = entry.op
		Register(entry.op, wrap(entry.name))
		bin := make([]byte, 3)
		bin[0] = byte(entry.op >> 16)
		bin[1] = byte(entry.op >> 8)
		bin[2] = byte(entry.op)
		log.Printf("[OPCODES] %-32s = %08b = 0x%06X",
			entry.name, bin, entry.op)
	}
	log.Printf("[OPCODES] %d opcodes registered; %d gas-priced", len(opcodeTable), len(gasTable))
}

// Hex returns the canonical hexadecimal representation (upper-case, 6 digits).
func (op Opcode) Hex() string { return fmt.Sprintf("0x%06X", op) }

// Bytes gives the 3-byte big-endian encoding used in VM bytecode streams.
func (op Opcode) Bytes() []byte {
	b := make([]byte, 3)
	b[0] = byte(op >> 16)
	b[1] = byte(op >> 8)
	b[2] = byte(op)
	return b
}

// String implements fmt.Stringer.
func (op Opcode) String() string { return op.Hex() }

// ParseOpcode converts a 3-byte slice into an Opcode, validating length.
func ParseOpcode(b []byte) (Opcode, error) {
	if len(b) != 3 {
		return 0, fmt.Errorf("opcode length must be 3, got %d", len(b))
	}
	return Opcode(uint32(b[0])<<16 | uint32(b[1])<<8 | uint32(b[2])), nil
}

// MustParseOpcode is a helper that panics on error (used in tests/tools).
func MustParseOpcode(b []byte) Opcode {
	op, err := ParseOpcode(b)
	if err != nil {
		panic(err)
	}
	return op
}

// DebugDump returns the full mapping in <name>=<hex> form, sorted
// lexicographically.  Useful for CLI tools (`synner opcodes`).
func DebugDump() []string {
	mu.RLock()
	defer mu.RUnlock()
	out := make([]string, 0, len(nameToOp))
	for n, op := range nameToOp {
		out = append(out, fmt.Sprintf("%s=%s", n, op.Hex()))
	}
	// simple in-place sort (no import cycle with `sort`)
	for i := 0; i < len(out)-1; i++ {
		for j := i + 1; j < len(out); j++ {
			if out[j] < out[i] {
				out[i], out[j] = out[j], out[i]
			}
		}
	}
	return out
}

// ToBytecode handy helper: returns raw 3-byte opcode plus gas meter prelude.
func ToBytecode(fn string) ([]byte, error) {
	op, ok := nameToOp[fn]
	if !ok {
		return nil, fmt.Errorf("unknown function %q", fn)
	}
	return op.Bytes(), nil
}

// HexDump is syntactic sugar: hex-encodes the 3-byte opcode.
func HexDump(fn string) (string, error) {
	b, err := ToBytecode(fn)
	if err != nil {
		return "", err
	}
	return "0x" + hex.EncodeToString(b), nil
}
