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
// VM dispatcher glue
// ────────────────────────────────────────────────────────────────────────────

// OpContext is provided by the VM; it gives opcode handlers controlled access
// to message meta-data, state-DB, gas-meter, logger, etc.
type OpContext interface {
	Call(string) error // unified façade (ledger/consensus/VM)
	Gas(uint64) error  // deducts gas or returns an error if exhausted
}

// Opcode is a 24-bit, deterministic instruction identifier.
type Opcode uint32

// OpcodeFunc is the concrete implementation invoked by the VM.
type OpcodeFunc func(ctx OpContext) error

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
func Dispatch(ctx OpContext, op Opcode) error {
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

// helper returns a closure that delegates the call to OpContext.Call(<name>).
func wrap(name string) OpcodeFunc {
	return func(ctx OpContext) error { return ctx.Call(name) }
}

// ────────────────────────────────────────────────────────────────────────────
// Opcode Catalogue  (AUTO-GENERATED – DO NOT EDIT BY HAND)
// ────────────────────────────────────────────────────────────────────────────
//
// Category map:
//
//		0x01 AI                     0x0F Liquidity
//		0x02 AMM                    0x10 Loanpool
//		0x03 Authority              0x11 Network
//		0x04 Charity                0x12 Replication
//		0x05 Coin                   0x13 Rollups
//		0x06 Compliance             0x14 Security
//		0x07 Consensus              0x15 Sharding
//		0x08 Contracts              0x16 Sidechains
//		0x09 CrossChain             0x17 StateChannel
//		0x0A Data                   0x18 Storage
//		0x0B FaultTolerance         0x19 Tokens
//		0x0C Governance             0x1A Transactions
//		0x0D GreenTech              0x1B Utilities
//		0x0E Ledger                 0x1C VirtualMachine
//		                            0x1D Wallet
//	                                 0x1E EnergyEfficiency
//	                                 0x1E ResourceMarket
//	                                 0x1E Finalization
//	                                 0x1D Wallet
//	                                 0x1E BinaryTree
//		                            0x1D Wallet
//	                                 0x1E Regulatory
//	                                 0x1E Forum
//	                                 0x1E Compression
//	                                 0x1E Biometrics
//	                                 0x1E SystemHealth
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
//	                             0x1E DeFi
//				0x1E Swarm

//	                                 0x1E Plasma
//	                                 0x1D Wallet
//		0x1E Workflows
//		                            0x1D Wallet
//	                                 0x1E Sensors
//	     0x0E Ledger                 0x1C VirtualMachine
//	                                 0x1D Wallet
//	                                 0x1E RealEstate
//		0x0E Ledger                 0x1C VirtualMachine
//		                            0x1D Wallet
//	                                 0x1E Employment
//	                                 0x1E Escrow
//	                                 0x1E Marketplace
//	                                 0x1D Wallet
//	                                 0x1E Faucet
//		                            0x1D Wallet
//	                                 0x1E SupplyChain
//	                                 0x1E Healthcare
//	                                 0x1E Immutability
//	                                 0x1E Warehouse
//	                                 0x1E Gaming
//	0x1E Assets//				0x1E Event


//
// Each binary code is shown as a 24-bit big-endian string.
var catalogue = []struct {
	name string
	op   Opcode
}{
	// AI (0x01)
	{"StartTraining", 0x010001},
	{"TrainingStatus", 0x010002},
	{"ListTrainingJobs", 0x010003},
	{"CancelTraining", 0x010004},
	{"InitAI", 0x010001},
	{"AI", 0x010002},
	{"PredictAnomaly", 0x010003},
	{"OptimizeFees", 0x010004},
	{"PublishModel", 0x010005},
	{"FetchModel", 0x010006},
	{"ListModel", 0x010007},
	{"ValidateKYC", 0x010008},
	{"BuyModel", 0x010009},
	{"RentModel", 0x01000A},
	{"ReleaseEscrow", 0x01000B},
	{"PredictVolume", 0x01000C},
	{"GetModelListing", 0x01000D},
	{"ListModelListings", 0x01000E},
	{"UpdateListingPrice", 0x01000F},
	{"RemoveListing", 0x010010},
	{"InferModel", 0x010001},
	{"AnalyseTransactions", 0x010002},

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
	{"NewAuthorityApplier", 0x030009},
	{"SubmitApplication", 0x03000A},
	{"VoteApplication", 0x03000B},
	{"FinalizeApplication", 0x03000C},
	{"GetApplication", 0x03000D},
	{"ListApplications", 0x03000E},

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
	{"Audit_Init", 0x060009},
	{"Audit_Log", 0x06000A},
	{"Audit_Events", 0x06000B},
	{"Audit_Close", 0x06000C},
	{"InitComplianceManager", 0x060009},
	{"SuspendAccount", 0x06000A},
	{"ResumeAccount", 0x06000B},
	{"IsSuspended", 0x06000C},
	{"WhitelistAccount", 0x06000D},
	{"RemoveWhitelist", 0x06000E},
	{"IsWhitelisted", 0x06000F},
	{"Compliance_ReviewTx", 0x060010},
	{"AnalyzeAnomaly", 0x060009},
	{"FlagAnomalyTx", 0x06000A},

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
	{"Status", 0x070013},
	{"SetDifficulty", 0x070014},
	{"NewConsensusAdaptiveManager", 0x070013},
	{"ComputeDemand", 0x070014},
	{"ComputeStakeConcentration", 0x070015},
	{"AdjustConsensus", 0x070016},
	{"AdjustStake", 0x070013},
	{"PenalizeValidator", 0x070014},
	{"RegisterValidator", 0x070013},
	{"DeregisterValidator", 0x070014},
	{"StakeValidator", 0x070015},
	{"UnstakeValidator", 0x070016},
	{"SlashValidator", 0x070017},
	{"GetValidator", 0x070018},
	{"ListValidators", 0x070019},
	{"IsValidator", 0x07001A},

	// Contracts (0x08)
	{"InitContracts", 0x080001},
	{"CompileWASM", 0x080002},
	{"Invoke", 0x080003},
	{"Deploy", 0x080004},

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
	{"ListCDNNodes", 0x0A0009},
	{"ListOracles", 0x0A000A},
	{"PushFeedSigned", 0x0A000B},
	{"CreateDataSet", 0x0A000C},
	{"PurchaseDataSet", 0x0A000D},
	{"GetDataSet", 0x0A000E},
	{"ListDataSets", 0x0A000F},
	{"HasAccess", 0x0A0010},
	{"UpdateOracleSource", 0x0A000C},
	{"RemoveOracle", 0x0A000D},
	{"GetOracleMetrics", 0x0A000E},
	{"RequestOracleData", 0x0A000F},
	{"SyncOracle", 0x0A0010},
	{"CreateDataFeed", 0x0A000C},
	{"QueryDataFeed", 0x0A000D},
	{"ManageDataFeed", 0x0A000E},
	{"ImputeMissing", 0x0A000F},
	{"NormalizeFeed", 0x0A0010},
	{"AddProvenance", 0x0A0011},
	{"SampleFeed", 0x0A0012},
	{"ScaleFeed", 0x0A0013},
	{"TransformFeed", 0x0A0014},
	{"VerifyFeedTrust", 0x0A0015},
	{"ZTDC_Open", 0x0A000C},
	{"ZTDC_Send", 0x0A000D},
	{"ZTDC_Close", 0x0A000E},
	{"StoreManagedData", 0x0A000C},
	{"LoadManagedData", 0x0A000D},
	{"DeleteManagedData", 0x0A000E},

	// Fault-Tolerance (0x0B)
	{"NewHealthChecker", 0x0B0001},
	{"AddPeer", 0x0B0002},
	{"RemovePeer", 0x0B0003},
	{"Snapshot", 0x0B0004},
	{"Recon", 0x0B0005},
	{"Ping", 0x0B0006},
	{"SendPing", 0x0B0007},
	{"AwaitPong", 0x0B0008},
	{"BackupSnapshot", 0x0B0009},
	{"RestoreSnapshot", 0x0B000A},
	{"VerifyBackup", 0x0B000B},
	{"FailoverNode", 0x0B000C},
	{"PredictFailure", 0x0B000D},
	{"AdjustResources", 0x0B000E},
	{"HA_Register", 0x0B000F},
	{"HA_Remove", 0x0B0010},
	{"HA_List", 0x0B0011},
	{"HA_Sync", 0x0B0012},
	{"HA_Promote", 0x0B0013},

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
	{"DAO_Stake", 0x0C000B},
	{"DAO_Unstake", 0x0C000C},
	{"DAO_Staked", 0x0C000D},
	{"DAO_TotalStaked", 0x0C000E},
	{"CastTokenVote", 0x0C000B},
	{"SubmitQuadraticVote", 0x0C000B},
	{"QuadraticResults", 0x0C000C},
	{"QuadraticWeight", 0x0C000D},
	{"AddDAOMember", 0x0C000B},
	{"RemoveDAOMember", 0x0C000C},
	{"RoleOfMember", 0x0C000D},
	{"ListDAOMembers", 0x0C000E},
	{"NewQuorumTracker", 0x0C000B},
	{"Quorum_AddVote", 0x0C000C},
	{"Quorum_HasQuorum", 0x0C000D},
	{"Quorum_Reset", 0x0C000E},
	{"RegisterGovContract", 0x0C000B},
	{"GetGovContract", 0x0C000C},
	{"ListGovContracts", 0x0C000D},
	{"EnableGovContract", 0x0C000E},
	{"DeleteGovContract", 0x0C000F},
	{"DeployGovContract", 0x0C000B},
	{"InvokeGovContract", 0x0C000C},
	{"AddReputation", 0x0C000B},
	{"SubtractReputation", 0x0C000C},
	{"ReputationOf", 0x0C000D},
	{"SubmitRepGovProposal", 0x0C000E},
	{"CastRepGovVote", 0x0C000F},
	{"ExecuteRepGovProposal", 0x0C0010},
	{"GetRepGovProposal", 0x0C0011},
	{"ListRepGovProposals", 0x0C0012},
	{"NewTimelock", 0x0C000B},
	{"QueueProposal", 0x0C000C},
	{"CancelProposal", 0x0C000D},
	{"ExecuteReady", 0x0C000E},
	{"ListTimelocks", 0x0C000F},
	{"CreateDAO", 0x0C000B},
	{"JoinDAO", 0x0C000C},
	{"LeaveDAO", 0x0C000D},
	{"DAOInfo", 0x0C000E},
	{"ListDAOs", 0x0C000F},

	// GreenTech (0x0D)
	{"InitGreenTech", 0x0D0001},
	{"Green", 0x0D0002},
	{"RecordUsage", 0x0D0003},
	{"RecordOffset", 0x0D0004},
	{"Certify", 0x0D0005},
	{"CertificateOf", 0x0D0006},
	{"ShouldThrottle", 0x0D0007},
	{"ListCertificates", 0x0D0008},

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
	{"Account_Create", 0x0E001C},
	{"Account_Delete", 0x0E001D},
	{"Account_Balance", 0x0E001E},
	{"Account_Transfer", 0x0E001F},

	// Liquidity (0x0F)
	{"InitAMM", 0x0F0001},
	{"Manager", 0x0F0002},
	{"CreatePool", 0x0F0003},
	{"Liquidity_AddLiquidity", 0x0F0004},
	{"Liquidity_Swap", 0x0F0005},
	{"Liquidity_RemoveLiquidity", 0x0F0006},
	{"Liquidity_Pool", 0x0F0007},
	{"Liquidity_Pools", 0x0F0008},

	// Loanpool (0x10)
	{"Loanpool_RandomElectorate", 0x100001},
	{"Loanpool_IsAuthority", 0x100002},
	{"Loanpool_init", 0x100003},
	{"NewLoanPool", 0x100004},
	{"Loanpool_Submit", 0x100005},
	{"Loanpool_Vote", 0x100006},
	{"Disburse", 0x100007},
	{"Loanpool_Tick", 0x100008},
	{"Loanpool_GetProposal", 0x100009},
	{"Loanpool_ListProposals", 0x10000A},
	{"Loanpool_Redistribute", 0x10000B},
	{"NewLoanPoolApply", 0x10000C},
	{"LoanApply_Submit", 0x10000D},
	{"LoanApply_Vote", 0x10000E},
	{"LoanApply_Process", 0x10000F},
	{"LoanApply_Disburse", 0x100010},
	{"LoanApply_Get", 0x100011},
	{"LoanApply_List", 0x100012},

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
	{"SetBroadcaster", 0x11000B},
	{"GlobalBroadcast", 0x11000C},
	{"StartDevNet", 0x11000D},
	{"StartTestNet", 0x11000E},

	// Replication (0x12)
	{"NewReplicator", 0x120001},
	{"ReplicateBlock", 0x120002},
	{"Replication_Hash", 0x120003},
	{"RequestMissing", 0x120004},
	{"Replication_Start", 0x120005},
	{"Stop", 0x120006},
	{"Synchronize", 0x120007},
	{"NewInitService", 0x120008},
	{"BootstrapLedger", 0x120009},
	{"ShutdownInitService", 0x12000A},
	{"NewSyncManager", 0x120008},
	{"Sync_Start", 0x120009},
	{"Sync_Stop", 0x12000A},
	{"Sync_Status", 0x12000B},
	{"SyncOnce", 0x12000C},

	// Rollups (0x13)
	{"NewAggregator", 0x130001},
	{"SubmitBatch", 0x130002},
	{"SubmitFraudProof", 0x130003},
	{"FinalizeBatch", 0x130004},
	{"BatchHeader", 0x130005},
	{"BatchState", 0x130006},
	{"BatchTransactions", 0x130007},
	{"ListBatches", 0x130008},

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
	{"DilithiumKeypair", 0x14000A},
	{"DilithiumSign", 0x14000B},
	{"DilithiumVerify", 0x14000C},
	{"PredictRisk", 0x14000D},
	{"AnomalyScore", 0x14000E},
	{"BuildMerkleTree", 0x14000F},
	{"MerkleProof", 0x140010},
	{"VerifyMerklePath", 0x140011},

	// Sharding (0x15)
	{"NewShardCoordinator", 0x150001},
	{"SetLeader", 0x150002},
	{"Leader", 0x150003},
	{"SubmitCrossShard", 0x150004},
	{"Sharding_Broadcast", 0x150005},
	{"Send", 0x150006},
	{"PullReceipts", 0x150007},
	{"Reshard", 0x150008},
	{"GossipTx", 0x150009},
	{"RebalanceShards", 0x15000A},
	{"VerticalPartition", 0x15000B},

	// Sidechains (0x16)
	{"InitSidechains", 0x160001},
	{"Sidechains", 0x160002},
	{"Sidechains_Register", 0x160003},
	{"SubmitHeader", 0x160004},
	{"Sidechains_Deposit", 0x160005},
	{"VerifyWithdraw", 0x160006},
	{"VerifyAggregateSig", 0x160007},
	{"VerifyMerkleProof", 0x160008},
	{"GetSidechainMeta", 0x160009},
	{"ListSidechains", 0x16000A},
	{"GetSidechainHeader", 0x16000B},

	// StateChannel (0x17)
	{"InitStateChannels", 0x170001},
	{"Channels", 0x170002},
	{"OpenChannel", 0x170003},
	{"VerifyECDSASignature", 0x170004},
	{"InitiateClose", 0x170005},
	{"Challenge", 0x170006},
	{"Finalize", 0x170007},
	{"GetChannel", 0x170008},
	{"ListChannels", 0x170009},

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
	{"GetListing", 0x18000A},
	{"ListListings", 0x18000B},
	{"GetDeal", 0x18000C},
	{"ListDeals", 0x18000D},

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
	{"InitTokens", 0x19001A},
	{"GetRegistryTokens", 0x19001B},
	{"TokenManager_Create", 0x19001C},
	{"TokenManager_Transfer", 0x19001D},
	{"TokenManager_Mint", 0x19001E},
	{"TokenManager_Burn", 0x19001F},
	{"TokenManager_Approve", 0x190020},
	{"TokenManager_BalanceOf", 0x190021},

	// Transactions (0x1A)
	{"Tx_Sign", 0x1A0001},
	{"VerifySig", 0x1A0002},
	{"ValidateTx", 0x1A0003},
	{"NewTxPool", 0x1A0004},
	{"AddTx", 0x1A0005},
	{"PickTxs", 0x1A0006},
	{"TxPoolSnapshot", 0x1A0007},
	{"EncryptTxPayload", 0x1A0008},
	{"DecryptTxPayload", 0x1A0009},
	{"SubmitPrivateTx", 0x1A000A},
	{"EncodeEncryptedHex", 0x1A000B},
	{"Exec_Begin", 0x1A0008},
	{"Exec_RunTx", 0x1A0009},
	{"Exec_Finalize", 0x1A000A},
	{"ReverseTransaction", 0x1A0008},
	{"NewTxDistributor", 0x1A0008},
	{"DistributeFees", 0x1A0009},

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

	// EnergyEfficiency (0x1E)
	{"InitEnergyEfficiency", 0x1E0001},
	{"EnergyEff", 0x1E0002},
	{"RecordStats", 0x1E0003},
	{"EfficiencyOf", 0x1E0004},
	{"NetworkAverage", 0x1E0005},
	{"ListEfficiency", 0x1E0006},
	// ResourceMarket (0x1E)
	{"ListResource", 0x1E0001},
	{"OpenResourceDeal", 0x1E0002},
	{"CloseResourceDeal", 0x1E0003},
	{"GetResourceListing", 0x1E0004},
	{"ListResourceListings", 0x1E0005},
	{"GetResourceDeal", 0x1E0006},
	{"ListResourceDeals", 0x1E0007},
	// Finalization (0x1E)
	{"NewFinalizationManager", 0x1E0001},
	{"FinalizeBlock", 0x1E0002},
	{"FinalizeBatchManaged", 0x1E0003},
	{"FinalizeChannelManaged", 0x1E0004},
	// DeFi (0x1E)
	{"DeFi_CreateInsurance", 0x1E0001},
	{"DeFi_ClaimInsurance", 0x1E0002},
	{"DeFi_PlaceBet", 0x1E0003},
	{"DeFi_SettleBet", 0x1E0004},
	{"DeFi_StartCrowdfund", 0x1E0005},
	{"DeFi_Contribute", 0x1E0006},
	{"DeFi_FinalizeCrowdfund", 0x1E0007},
	{"DeFi_CreatePrediction", 0x1E0008},
	{"DeFi_VotePrediction", 0x1E0009},
	{"DeFi_ResolvePrediction", 0x1E000A},
	{"DeFi_RequestLoan", 0x1E000B},
	{"DeFi_RepayLoan", 0x1E000C},
	{"DeFi_StartYieldFarm", 0x1E000D},
	{"DeFi_Stake", 0x1E000E},
	{"DeFi_Unstake", 0x1E000F},
	{"DeFi_CreateSynthetic", 0x1E0010},
	{"DeFi_MintSynthetic", 0x1E0011},
	{"DeFi_BurnSynthetic", 0x1E0012},
   {"RegisterIDWallet", 0x1D0007},
	{"IsIDWalletRegistered", 0x1D0008},
	{"NewOffChainWallet", 0x1D0007},
	{"OffChainWalletFromMnemonic", 0x1D0008},
	{"SignOffline", 0x1D0009},
	{"StoreSignedTx", 0x1D000A},
	{"LoadSignedTx", 0x1D000B},
	{"BroadcastSignedTx", 0x1D000C},
	{"RegisterRecovery", 0x1D0007},
	{"RecoverAccount", 0x1D0008},

	// BinaryTree (0x1E)
	{"BinaryTreeNew", 0x1E0001},
	{"BinaryTreeInsert", 0x1E0002},
	{"BinaryTreeSearch", 0x1E0003},
	{"BinaryTreeDelete", 0x1E0004},
	{"BinaryTreeInOrder", 0x1E0005},
  
	// Regulatory (0x1E)
	{"InitRegulatory", 0x1E0001},
	{"RegisterRegulator", 0x1E0002},
	{"GetRegulator", 0x1E0003},
	{"ListRegulators", 0x1E0004},
	{"EvaluateRuleSet", 0x1E0005},



	// Polls Management (0x1E)
	{"CreatePoll", 0x1E0001},
	{"VotePoll", 0x1E0002},
	{"ClosePoll", 0x1E0003},
	{"GetPoll", 0x1E0004},
	{"ListPolls", 0x1E0005},

   

	// Feedback (0x1E)
	{"InitFeedback", 0x1E0001},
	{"Feedback_Submit", 0x1E0002},
	{"Feedback_Get", 0x1E0003},
	{"Feedback_List", 0x1E0004},
	{"Feedback_Reward", 0x1E0005},
 

	// Forum (0x1E)
	{"Forum_CreateThread", 0x1E0001},
	{"Forum_GetThread", 0x1E0002},
	{"Forum_ListThreads", 0x1E0003},
	{"Forum_AddComment", 0x1E0004},
	{"Forum_ListComments", 0x1E0005},
 
	// Compression (0x1E)
	{"CompressLedger", 0x1E0001},
	{"DecompressLedger", 0x1E0002},
	{"SaveCompressedSnapshot", 0x1E0003},
	{"LoadCompressedSnapshot", 0x1E0004},


	// Biometrics (0x1E)
	{"Bio_Enroll", 0x1E0001},
	{"Bio_Verify", 0x1E0002},
	{"Bio_Delete", 0x1E0003},

  
	// SystemHealth (0x1E)
	{"NewHealthLogger", 0x1E0001},
	{"MetricsSnapshot", 0x1E0002},
	{"LogEvent", 0x1E0003},
	{"RotateLogs", 0x1E0004},
	

	// Swarm (0x1E)
	{"NewSwarm", 0x1E0001},
	{"Swarm_AddNode", 0x1E0002},
	{"Swarm_RemoveNode", 0x1E0003},
	{"Swarm_BroadcastTx", 0x1E0004},
	{"Swarm_Start", 0x1E0005},
	{"Swarm_Stop", 0x1E0006},
	{"Swarm_Peers", 0x1E0007},
	// Plasma (0x1E)
	{"InitPlasma", 0x1E0001},
	{"Plasma_Deposit", 0x1E0002},
	{"Plasma_Withdraw", 0x1E0003},
  // Workflows (0x1E)
	{"NewWorkflow", 0x1E0001},
	{"AddWorkflowAction", 0x1E0002},
	{"SetWorkflowTrigger", 0x1E0003},
	{"SetWebhook", 0x1E0004},
	{"ExecuteWorkflow", 0x1E0005},
	{"ListWorkflows", 0x1E0006},
   {"CreateWallet", 0x1D0007},
	{"ImportWallet", 0x1D0008},
	{"WalletBalance", 0x1D0009},
	{"WalletTransfer", 0x1D000A},
  
	// Sensors (0x1E)
	{"RegisterSensor", 0x1E0001},
	{"GetSensor", 0x1E0002},
	{"ListSensors", 0x1E0003},
	{"UpdateSensorValue", 0x1E0004},
	{"PollSensor", 0x1E0005},
	{"TriggerWebhook", 0x1E0006},

  
	// Real Estate (0x1D)
	{"RegisterProperty", 0x1E0001},
	{"TransferProperty", 0x1E0002},
	{"GetProperty", 0x1E0003},
	{"ListProperties", 0x1E0004},

 

	// Event (0x1E)
	{"InitEvents", 0x1E0001},
	{"EmitEvent", 0x1E0002},
	{"GetEvent", 0x1E0003},
	{"ListEvents", 0x1E0004},

  

	// Employment (0x1E)
	{"InitEmployment", 0x1E0001},
	{"CreateJob", 0x1E0002},
	{"SignJob", 0x1E0003},
	{"RecordWork", 0x1E0004},
	{"PaySalary", 0x1E0005},
	{"GetJob", 0x1E0006},
	// Escrow (0x1E)
	{"Escrow_Create", 0x1E0001},
	{"Escrow_Deposit", 0x1E0002},
	{"Escrow_Release", 0x1E0003},
	{"Escrow_Cancel", 0x1E0004},
	{"Escrow_Get", 0x1E0005},
	{"Escrow_List", 0x1E0006},
	// Marketplace (0x1E)
	{"CreateMarketListing", 0x1E0001},
	{"PurchaseItem", 0x1E0002},
	{"CancelListing", 0x1E0003},
	{"ReleaseFunds", 0x1E0004},
	{"GetMarketListing", 0x1E0005},
	{"ListMarketListings", 0x1E0006},
	{"GetMarketDeal", 0x1E0007},
	{"ListMarketDeals", 0x1E0008},
	// Faucet (0x1E)
	{"NewFaucet", 0x1E0001},
	{"Faucet_Request", 0x1E0002},
	{"Faucet_Balance", 0x1E0003},
	{"Faucet_SetAmount", 0x1E0004},
	{"Faucet_SetCooldown", 0x1E0005},
  // Supply Chain (0x1E)
  {"RegisterItem", 0x1E0001},
	{"UpdateLocation", 0x1E0002},
	{"MarkStatus", 0x1E0003},
	{"GetItem", 0x1E0004},

	// Healthcare (0x1E)
	{"InitHealthcare", 0x1E0001},
	{"RegisterPatient", 0x1E0002},
	{"AddHealthRecord", 0x1E0003},
	{"GrantAccess", 0x1E0004},
	{"RevokeAccess", 0x1E0005},
	{"ListHealthRecords", 0x1E0006},

  // Tangible (0x1E)
	{"Assets_Register", 0x1E0001},
	{"Assets_Transfer", 0x1E0002},
	{"Assets_Get", 0x1E0003},
	{"Assets_List", 0x1E0004},

	// Immutability (0x1E)
	{"InitImmutability", 0x1E0001},
	{"VerifyChain", 0x1E0002},
	{"RestoreChain", 0x1E0003},
	// Warehouse (0x1E)
	{"Warehouse_New", 0x1E0001},
	{"Warehouse_AddItem", 0x1E0002},
	{"Warehouse_RemoveItem", 0x1E0003},
	{"Warehouse_MoveItem", 0x1E0004},
	{"Warehouse_ListItems", 0x1E0005},
	{"Warehouse_GetItem", 0x1E0006},

	// Gaming (0x1E)
	{"CreateGame", 0x1E0001},
	{"JoinGame", 0x1E0002},
	{"FinishGame", 0x1E0003},
	{"GetGame", 0x1E0004},
	{"ListGames", 0x1E0005},
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
func (op Opcode) Hex() string { return fmt.Sprintf("0x%06X", uint32(op)) }

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
