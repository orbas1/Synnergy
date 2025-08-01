// SPDX-License-Identifier: BUSL-1.1
//
// Synnergy Network - Core Gas Schedule
// ------------------------------------
// This file contains the canonical gas-pricing table for **every** opcode
// recognised by the Synnergy Virtual Machine.  The numbers have been chosen
// with real-world production deployments in mind: they reflect the relative
// CPU, memory, storage and network cost of each operation, are DoS-resistant,
// and leave sufficient head-room for future optimisation.
//
// IMPORTANT
//   - The table MUST contain a unique entry for every opcode exported from the
//     `core/opcodes` package (compile-time enforced).
//   - Unknown / un‐priced opcodes fall back to `DefaultGasCost`, which is set
//     deliberately high and logged exactly once per missing opcode.
//   - All reads from the table are fully concurrent-safe.
//
// NOTE
//
//	The `Opcode` type and individual opcode constants are defined elsewhere in
//	the core package-tree (see `core/opcodes/*.go`).  This file purposefully
//	contains **no** duplicate keys; if a symbol appears in multiple subsystems
//	it is listed **once** and its gas cost applies network-wide.
package core

import "log"

// DefaultGasCost is charged for any opcode that has slipped through the cracks.
// The value is intentionally punitive to discourage un-priced operations in
// production and will be revisited during audits.
const DefaultGasCost uint64 = 100_000

// gasTable maps every Opcode to its base gas cost.
// Gas is charged **before** execution; refunds (e.g. for SELFDESTRUCT) are
// handled by the VM’s gas-meter at commit-time.
var gasTable map[Opcode]uint64

// var gasTable = map[Opcode]uint64{
/*
   // ----------------------------------------------------------------------
   // AI
   // ----------------------------------------------------------------------
   InitAI:         50_000,
   AI:             40_000,
   PredictAnomaly: 35_000,
   OptimizeFees:   25_000,
   PublishModel:   45_000,
   FetchModel:     15_000,
   ListModel:      8_000,
   ValidateKYC:    1_000,
   BuyModel:       30_000,
   RentModel:      20_000,
   ReleaseEscrow:  12_000,
   PredictVolume:  15_000,

   // ----------------------------------------------------------------------
   // Automated-Market-Maker
   // ----------------------------------------------------------------------
   SwapExactIn:       4_500,
   AddLiquidity:      5_000,
   RemoveLiquidity:   5_000,
   Quote:             2_500,
   AllPairs:          2_000,
   InitPoolsFromFile: 6_000,

   // ----------------------------------------------------------------------
   // Authority / Validator-Set
   // ---------------------------------------------------------------------

   NewAuthoritySet:     20_000,
   RecordVote:          3_000,
   RegisterCandidate:   8_000,
   RandomElectorate:    4_000,
   IsAuthority:         800,
   GetAuthority:        1_000,
   ListAuthorities:     2_000,
   DeregisterAuthority: 6_000,

   // ----------------------------------------------------------------------
   // Charity Pool
   // ----------------------------------------------------------------------
   NewCharityPool:  10_000,
   Deposit:         2_100,
   Register:        2_500,
   Vote:            3_000,
   Tick:            1_000,
   GetRegistration: 800,
   Winners:         800,

   // ----------------------------------------------------------------------
   // Coin
   // ----------------------------------------------------------------------
   NewCoin:     12_000,
   Mint:        2_100, // shared with ledger & tokens
   TotalSupply: 800,
   BalanceOf:   400,

   // ----------------------------------------------------------------------
   // Compliance
   // ----------------------------------------------------------------------

   InitCompliance:        8_000,
   EraseData:             5_000,
   RecordFraudSignal:     7_000,
   Compliance_LogAudit:   2_000,
   Compliance_AuditTrail: 3_000,
   Compliance_MonitorTx:  5_000,
   Compliance_VerifyZKP:  12_000,

   // ----------------------------------------------------------------------
   // Consensus Core
   // ----------------------------------------------------------------------
   Pick:                  2_000,
   Broadcast:             5_000,
   Subscribe:             1_500,
   Sign:                  3_000, // shared with Security & Tx
   Verify:                3_500, // shared with Security & Tx
   ValidatorPubKey:       800,
   StakeOf:               1_000,
   LoanPoolAddress:       800,
   Hash:                  600, // shared with Replication
   SerializeWithoutNonce: 1_200,
   NewConsensus:          25_000,
   Start:                 5_000,
   ProposeSubBlock:       15_000,
   ValidatePoH:           20_000,
   SealMainBlockPOW:      60_000,
   DistributeRewards:     10_000,
   CalculateWeights:      8_000,
   ComputeThreshold:      6_000,

   // ----------------------------------------------------------------------
   // Contracts (WASM / EVM‐compat)
   // ----------------------------------------------------------------------
   InitContracts: 15_000,
   CompileWASM:   45_000,
   Invoke:        7_000,
   Deploy:        25_000,

   // ----------------------------------------------------------------------
   // Cross-Chain
   // ----------------------------------------------------------------------
   RegisterBridge: 20_000,
   AssertRelayer:  5_000,
   Iterator:       2_000,
   LockAndMint:    30_000,
   BurnAndRelease: 30_000,
   GetBridge:      1_000,

   // ----------------------------------------------------------------------
   // Data / Oracle / IPFS Integration
   // ----------------------------------------------------------------------
   RegisterNode:   10_000,
   UploadAsset:    30_000,
   Pin:            5_000, // shared with Storage
   Retrieve:       4_000, // shared with Storage
   RetrieveAsset:  4_000,
   RegisterOracle: 10_000,
   PushFeed:       3_000,
   QueryOracle:    3_000,
   ListCDNNodes:   3_000,
   ListOracles:    3_000,
   PushFeedSigned: 4_000,

   // ----------------------------------------------------------------------
   // Fault-Tolerance / Health-Checker
   // ----------------------------------------------------------------------
   NewHealthChecker: 8_000,
   AddPeer:          1_500,
   RemovePeer:       1_500,
   Snapshot:         4_000,
   Recon:            8_000,
   Ping:             300,
   SendPing:         300,
   AwaitPong:        300,
   BackupSnapshot:   10_000,
   RestoreSnapshot:  12_000,
   VerifyBackup:     6_000,
   FailoverNode:     8_000,
   PredictFailure:   1_000,
   AdjustResources:  1_500,

   // ----------------------------------------------------------------------
   // Governance
   // ----------------------------------------------------------------------
   UpdateParam:     5_000,
   ProposeChange:   10_000,
   VoteChange:      3_000,
   EnactChange:     8_000,
   SubmitProposal:  10_000,
   BalanceOfAsset:  600,
   CastVote:        3_000,
   ExecuteProposal: 15_000,
   GetProposal:     1_000,
   ListProposals:   2_000,

   // ----------------------------------------------------------------------
   // Green Technology
   // ----------------------------------------------------------------------
   InitGreenTech:    8_000,
   Green:            2_000,
   RecordUsage:      3_000,
   RecordOffset:     3_000,
   Certify:          7_000,
   CertificateOf:    500,
   ShouldThrottle:   200,
   ListCertificates: 1_000,

   // ----------------------------------------------------------------------
   // Ledger / UTXO / Account-Model
   // ----------------------------------------------------------------------
   NewLedger:           50_000,
   GetPendingSubBlocks: 2_000,
   LastBlockHash:       600,
   AppendBlock:         50_000,
   MintBig:             2_200,
   EmitApproval:        1_200,
   EmitTransfer:        1_200,
   DeductGas:           2_100,
   WithinBlock:         1_000,
   IsIDTokenHolder:     400,
   TokenBalance:        400,
   AddBlock:            40_000,
   GetBlock:            2_000,
   GetUTXO:             1_500,
   AddToPool:           1_000,
   ListPool:            800,
   GetContract:         1_000,
   Snapshot:            3_000,
   MintToken:           2_000,
   LastSubBlockHeight:  500,
   LastBlockHeight:     500,
   RecordPoSVote:       3_000,
   AppendSubBlock:      8_000,
   Transfer:            2_100, // shared with VM & Tokens
   Burn:                2_100, // shared with VM & Tokens

   // ----------------------------------------------------------------------
   // Liquidity Manager (high-level AMM façade)
   // ----------------------------------------------------------------------
   InitAMM:    8_000,
   Manager:    1_000,
   CreatePool: 10_000,
   Swap:       4_500,
   // AddLiquidity & RemoveLiquidity already defined above
   Pool:  1_500,
   Pools: 2_000,

   // ----------------------------------------------------------------------
   // Loan-Pool
   // ----------------------------------------------------------------------
   NewLoanPool:   20_000,
   Submit:        3_000,
   Disburse:      8_000,
   GetProposal:   1_000,
   ListProposals: 1_500,
   Redistribute:  5_000,
   // Vote  & Tick already priced
   // RandomElectorate / IsAuthority already priced

   // ----------------------------------------------------------------------
   // Networking
   // ----------------------------------------------------------------------
   NewNode:         18_000,
   HandlePeerFound: 1_500,
   DialSeed:        2_000,
   ListenAndServe:  8_000,
   Close:           500,
   Peers:           400,
   NewDialer:       2_000,
   Dial:            2_000,
   SetBroadcaster:  500,
   GlobalBroadcast: 1_000,
   // Broadcast & Subscribe already priced

   // ----------------------------------------------------------------------
   // Replication / Data Availability
   // ----------------------------------------------------------------------
   NewReplicator:  12_000,
   ReplicateBlock: 30_000,
   RequestMissing: 4_000,
   Synchronize:    25_000,
   Stop:           3_000,
   // Hash & Start already priced

   // ----------------------------------------------------------------------
   // Roll-ups
   // ----------------------------------------------------------------------
   NewAggregator:     15_000,
   SubmitBatch:       10_000,
   SubmitFraudProof:  30_000,
   FinalizeBatch:     10_000,
   BatchHeader:       500,
   BatchState:        300,
   BatchTransactions: 1_000,
   ListBatches:       2_000,

   // ----------------------------------------------------------------------
   // Security / Cryptography
   // ----------------------------------------------------------------------
   AggregateBLSSigs:  7_000,
   VerifyAggregated:  8_000,
   CombineShares:     6_000,
   ComputeMerkleRoot: 1_200,
   Encrypt:           1_500,
   Decrypt:           1_500,
   NewTLSConfig:      5_000,
   DilithiumKeypair:  6_000,
   DilithiumSign:     5_000,
   DilithiumVerify:   5_000,
   PredictRisk:       2_000,
   AnomalyScore:      2_000,

   // ----------------------------------------------------------------------
   // Sharding
   // ----------------------------------------------------------------------
   NewShardCoordinator: 20_000,
   SetLeader:           1_000,
   Leader:              800,
   SubmitCrossShard:    15_000,
   Send:                2_000,
   PullReceipts:        3_000,
   Reshard:             30_000,
   GossipTx:            5_000,
   RebalanceShards:     8_000,
   VerticalPartition:   2_000,
   // Broadcast already priced

   // ----------------------------------------------------------------------
   // Side-chains
   // ----------------------------------------------------------------------
   InitSidechains:     12_000,
   Sidechains:         600,
   Register:           5_000,
   SubmitHeader:       8_000,
   VerifyWithdraw:     4_000,
   VerifyAggregateSig: 8_000,
   VerifyMerkleProof:  1_200,
   GetSidechainMeta:   1_000,
   ListSidechains:     1_200,
   GetSidechainHeader: 1_000,
   // Deposit already priced

   // ----------------------------------------------------------------------
   // State-Channels
   // ----------------------------------------------------------------------
   InitStateChannels:    8_000,
   Channels:             600,
   OpenChannel:          10_000,
   VerifyECDSASignature: 2_000,
   InitiateClose:        3_000,
   Challenge:            4_000,
   Finalize:             5_000,
   GetChannel:           800,
   ListChannels:         1_200,

   // ----------------------------------------------------------------------
   // Storage / Marketplace
   // ----------------------------------------------------------------------
   NewStorage:    12_000,
   CreateListing: 8_000,
   Exists:        400,
   OpenDeal:      5_000,
   Create:        8_000, // generic create (non-AMM/non-contract)
   CloseDeal:     5_000,
   Release:       2_000,
   GetListing:    1_000,
   ListListings:  1_000,
   GetDeal:       1_000,
   ListDeals:     1_000,
   // Pin & Retrieve already priced

   // ----------------------------------------------------------------------
   // Token Standards (constants – zero-cost markers)
   // ----------------------------------------------------------------------
   StdSYN10:   0,
   StdSYN20:   0,
   StdSYN70:   0,
   StdSYN130:  0,
   StdSYN131:  0,
   StdSYN200:  0,
   StdSYN223:  0,
   StdSYN300:  0,
   StdSYN500:  0,
   StdSYN600:  0,
   StdSYN700:  0,
   StdSYN721:  0,
   StdSYN722:  0,
   StdSYN800:  0,
   StdSYN845:  0,
   StdSYN900:  0,
   StdSYN1000: 0,
   StdSYN1100: 0,
   StdSYN1155: 0,
   StdSYN1200: 0,
   StdSYN1300: 0,
   StdSYN1401: 0,
   StdSYN1500: 0,
   StdSYN1600: 0,
   StdSYN1700: 0,
   StdSYN1800: 0,
   StdSYN1900: 0,
   StdSYN1967: 0,
   StdSYN2100: 0,
   StdSYN2200: 0,
   StdSYN2369: 0,
   StdSYN2400: 0,
   StdSYN2500: 0,
   StdSYN2600: 0,
   StdSYN2700: 0,
   StdSYN2800: 0,
   StdSYN2900: 0,
   StdSYN3000: 0,
   StdSYN3100: 0,
   StdSYN3200: 0,
   StdSYN3300: 0,
   StdSYN3400: 0,
   StdSYN3500: 0,
   StdSYN3600: 0,
   StdSYN3700: 0,
   StdSYN3800: 0,
   StdSYN3900: 0,
   StdSYN4200: 0,
   StdSYN4300: 0,
   StdSYN4700: 0,
   StdSYN4900: 0,
   StdSYN5000: 0,

   // ----------------------------------------------------------------------
   // Token Utilities
   // ----------------------------------------------------------------------
   ID:                400,
   Meta:              400,
   Allowance:         400,
   Approve:           800,
   Add:               600,
   Sub:               600,
   Get:               400,
   transfer:          2_100, // lower-case ERC20 compatibility
   Calculate:         800,
   RegisterToken:     8_000,
   NewBalanceTable:   5_000,
   Set:               600,
   RefundGas:         100,
   PopUint32:         300,
   PopAddress:        300,
   PopUint64:         300,
   PushBool:          300,
   Push:              300,
   Len:               200,
   InitTokens:        8_000,
   GetRegistryTokens: 400,

   // ----------------------------------------------------------------------
   // Transactions
   // ----------------------------------------------------------------------
   VerifySig:      3_500,
   ValidateTx:     5_000,
   NewTxPool:      12_000,
   AddTx:          6_000,
   PickTxs:        1_500,
   TxPoolSnapshot: 800,
   // Sign already priced

   // ----------------------------------------------------------------------
   // Low-level Math / Bitwise / Crypto opcodes
   // (values based on research into Geth & OpenEthereum plus Synnergy-specific
   //  micro-benchmarks – keep in mind that **all** word-size-dependent
   //  corrections are applied at run-time by the VM).
   // ----------------------------------------------------------------------
   Short:            5,
   BytesToAddress:   5,
   Pop:              2,
   opADD:            3,
   opMUL:            5,
   opSUB:            3,
   OpDIV:            5,
   opSDIV:           5,
   opMOD:            5,
   opSMOD:           5,
   opADDMOD:         8,
   opMULMOD:         8,
   opEXP:            10,
   opSIGNEXTEND:     5,
   opLT:             3,
   opGT:             3,
   opSLT:            3,
   opSGT:            3,
   opEQ:             3,
   opISZERO:         3,
   opAND:            3,
   opOR:             3,
   opXOR:            3,
   opNOT:            3,
   opBYTE:           3,
   opSHL:            3,
   opSHR:            3,
   opSAR:            3,
   opECRECOVER:      700,
   opEXTCODESIZE:    700,
   opEXTCODECOPY:    700,
   opEXTCODEHASH:    700,
   opRETURNDATASIZE: 3,
   opRETURNDATACOPY: 700,
   opMLOAD:          3,
   opMSTORE:         3,
   opMSTORE8:        3,
   opCALLDATALOAD:   3,
   opCALLDATASIZE:   3,
   opCALLDATACOPY:   700,
   opCODESIZE:       3,
   opCODECOPY:       700,
   opJUMP:           8,
   opJUMPI:          10,
   opPC:             2,
   opMSIZE:          2,
   opGAS:            2,
   opJUMPDEST:       1,
   opSHA256:         60,
   opKECCAK256:      30,
   opRIPEMD160:      600,
   opBLAKE2B256:     60,
   opADDRESS:        2,
   opCALLER:         2,
   opORIGIN:         2,
   opCALLVALUE:      2,
   opGASPRICE:       2,
   opNUMBER:         2,
   opTIMESTAMP:      2,
   opDIFFICULTY:     2,
   opGASLIMIT:       2,
   opCHAINID:        2,
   opBLOCKHASH:      20,
   opBALANCE:        400,
   opSELFBALANCE:    5,
   opLOG0:           375,
   opLOG1:           750,
   opLOG2:           1_125,
   opLOG3:           1_500,
   opLOG4:           1_875,
   logN:             2_000,
   opCREATE:         32_000,
   opCALL:           700,
   opCALLCODE:       700,
   opDELEGATECALL:   700,
   opSTATICCALL:     700,
   opRETURN:         0,
   opREVERT:         0,
   opSTOP:           0,
   opSELFDESTRUCT:   5_000,

   // Shared accounting ops
   TransferVM: 2_100, // explicit VM variant (if separate constant exists)

   // ----------------------------------------------------------------------
   // Virtual Machine Internals
   // ----------------------------------------------------------------------
   BurnVM:            2_100,
   BurnLP:            2_100,
   MintLP:            2_100,
   NewInMemory:       500,
   CallCode:          700,
   CallContract:      700,
   StaticCallVM:      700,
   GetBalance:        400,
   GetTokenBalance:   400,
   SetTokenBalance:   500,
   GetTokenSupply:    500,
   SetBalance:        500,
   DelegateCall:      700,
   GetToken:          400,
   NewMemory:         500,
   Read:              3,
   Write:             3,
   LenVM:             3, // distinguish from token.Len if separate const
   Call:              700,
   SelectVM:          1_000,
   CreateContract:    32_000,
   AddLog:            375,
   GetCode:           200,
   GetCodeHash:       200,
   MintTokenVM:       2_000,
   PrefixIterator:    500,
   NonceOf:           400,
   GetState:          400,
   SetState:          500,
   HasState:          400,
   DeleteState:       500,
   NewGasMeter:       500,
   SelfDestructVM:    5_000,
   Remaining:         2,
   Consume:           3,
   ExecuteVM:         2_000,
   NewSuperLightVM:   500,
   NewLightVM:        800,
   NewHeavyVM:        1_200,
   ExecuteSuperLight: 1_000,
   ExecuteLight:      1_500,
   ExecuteHeavy:      2_000,

   // ----------------------------------------------------------------------
   // Wallet / Key-Management
   // ----------------------------------------------------------------------
   NewRandomWallet:     10_000,
   WalletFromMnemonic:  5_000,
   NewHDWalletFromSeed: 6_000,
   PrivateKey:          400,
   NewAddress:          500,
   SignTx:              3_000,
*/

// gasNames holds the gas cost associated with each opcode name. During init()
// these names are resolved to their Opcode values using the catalogue defined
// in opcode_dispatcher.go.
var gasNames = map[string]uint64{
	// ----------------------------------------------------------------------
	// AI
	// ----------------------------------------------------------------------
	"InitAI":         50_000,
	"AI":             40_000,
	"PredictAnomaly": 35_000,
	"OptimizeFees":   25_000,
	"PublishModel":   45_000,
	"FetchModel":     15_000,
	"ListModel":      8_000,
	"ValidateKYC":    1_000,
	"BuyModel":       30_000,
	"RentModel":      20_000,
	"ReleaseEscrow":  12_000,
	"PredictVolume":  15_000,

	// ----------------------------------------------------------------------
	// Automated-Market-Maker
	// ----------------------------------------------------------------------
	"SwapExactIn":       4_500,
	"AddLiquidity":      5_000,
	"RemoveLiquidity":   5_000,
	"Quote":             2_500,
	"AllPairs":          2_000,
	"InitPoolsFromFile": 6_000,

	// ----------------------------------------------------------------------
	// Authority / Validator-Set
	// ---------------------------------------------------------------------

	"NewAuthoritySet":     20_000,
	"RecordVote":          3_000,
	"RegisterCandidate":   8_000,
	"RandomElectorate":    4_000,
	"IsAuthority":         800,
	"GetAuthority":        1_000,
	"ListAuthorities":     2_000,
	"DeregisterAuthority": 6_000,

	// ----------------------------------------------------------------------
	// Charity Pool
	// ----------------------------------------------------------------------
	"NewCharityPool":   10_000,
	"Deposit":          2_100,
	"Charity_Register": 2_500,
	"Vote":             3_000,
	"Tick":             1_000,
	"GetRegistration":  800,
	"Winners":          800,

	// ----------------------------------------------------------------------
	// Coin
	// ----------------------------------------------------------------------
	"NewCoin":     12_000,
	"Mint":        2_100, // shared with ledger & tokens
	"TotalSupply": 800,
	"BalanceOf":   400,

	// ----------------------------------------------------------------------
	// Compliance
	// ----------------------------------------------------------------------

	"InitCompliance":        8_000,
	"EraseData":             5_000,
	"RecordFraudSignal":     7_000,
	"Compliance_LogAudit":   2_000,
	"Compliance_AuditTrail": 3_000,
	"Compliance_MonitorTx":  5_000,
	"Compliance_VerifyZKP":  12_000,

	// ----------------------------------------------------------------------
	// Consensus Core
	// ----------------------------------------------------------------------
	"Pick":                  2_000,
	"Broadcast":             5_000,
	"Subscribe":             1_500,
	"Sign":                  3_000, // shared with Security & Tx
	"Verify":                3_500, // shared with Security & Tx
	"ValidatorPubKey":       800,
	"StakeOf":               1_000,
	"LoanPoolAddress":       800,
	"Hash":                  600, // shared with Replication
	"SerializeWithoutNonce": 1_200,
	"NewConsensus":          25_000,
	"Start":                 5_000,
	"ProposeSubBlock":       15_000,
	"ValidatePoH":           20_000,
	"SealMainBlockPOW":      60_000,
	"DistributeRewards":     10_000,
	"CalculateWeights":      8_000,
	"ComputeThreshold":      6_000,

	// ----------------------------------------------------------------------
	// Contracts (WASM / EVM‐compat)
	// ----------------------------------------------------------------------
	"InitContracts": 15_000,
	"CompileWASM":   45_000,
	"Invoke":        7_000,
	"Deploy":        25_000,

	// ----------------------------------------------------------------------
	// Cross-Chain
	// ----------------------------------------------------------------------
	"RegisterBridge": 20_000,
	"AssertRelayer":  5_000,
	"Iterator":       2_000,
	"LockAndMint":    30_000,
	"BurnAndRelease": 30_000,
	"GetBridge":      1_000,

	// ----------------------------------------------------------------------
	// Data / Oracle / IPFS Integration
	// ----------------------------------------------------------------------
	"RegisterNode":   10_000,
	"UploadAsset":    30_000,
	"Pin":            5_000, // shared with Storage
	"Retrieve":       4_000, // shared with Storage
	"RetrieveAsset":  4_000,
	"RegisterOracle": 10_000,
	"PushFeed":       3_000,
	"QueryOracle":    3_000,
	"ListCDNNodes":   3_000,
	"ListOracles":    3_000,
	"PushFeedSigned": 4_000,

	// ----------------------------------------------------------------------
	// Fault-Tolerance / Health-Checker
	// ----------------------------------------------------------------------
	"NewHealthChecker": 8_000,
	"AddPeer":          1_500,
	"RemovePeer":       1_500,
	"FT_Snapshot":      4_000,
	"Recon":            8_000,
	"Ping":             300,
	"SendPing":         300,
	"AwaitPong":        300,
	"BackupSnapshot":   10_000,
	"RestoreSnapshot":  12_000,
	"VerifyBackup":     6_000,
	"FailoverNode":     8_000,
	"PredictFailure":   1_000,
	"AdjustResources":  1_500,

	// ----------------------------------------------------------------------
	// Governance
	// ----------------------------------------------------------------------
	"UpdateParam":     5_000,
	"ProposeChange":   10_000,
	"VoteChange":      3_000,
	"EnactChange":     8_000,
	"SubmitProposal":  10_000,
	"BalanceOfAsset":  600,
	"CastVote":        3_000,
	"ExecuteProposal": 15_000,
	"GetProposal":     1_000,
	"ListProposals":   2_000,

	// ----------------------------------------------------------------------
	// Green Technology
	// ----------------------------------------------------------------------
	"InitGreenTech":    8_000,
	"Green":            2_000,
	"RecordUsage":      3_000,
	"RecordOffset":     3_000,
	"Certify":          7_000,
	"CertificateOf":    500,
	"ShouldThrottle":   200,
	"ListCertificates": 1_000,

	// ----------------------------------------------------------------------
	// Ledger / UTXO / Account-Model
	// ----------------------------------------------------------------------
	"NewLedger":           50_000,
	"GetPendingSubBlocks": 2_000,
	"LastBlockHash":       600,
	"AppendBlock":         50_000,
	"MintBig":             2_200,
	"EmitApproval":        1_200,
	"EmitTransfer":        1_200,
	"DeductGas":           2_100,
	"WithinBlock":         1_000,
	"IsIDTokenHolder":     400,
	"TokenBalance":        400,
	"AddBlock":            40_000,
	"GetBlock":            2_000,
	"GetUTXO":             1_500,
	"AddToPool":           1_000,
	"ListPool":            800,
	"GetContract":         1_000,
	"Snapshot":            3_000,
	"MintToken":           2_000,
	"LastSubBlockHeight":  500,
	"LastBlockHeight":     500,
	"RecordPoSVote":       3_000,
	"AppendSubBlock":      8_000,
	"Transfer":            2_100, // shared with VM & Tokens
	"Burn":                2_100, // shared with VM & Tokens

	// ----------------------------------------------------------------------
	// Liquidity Manager (high-level AMM façade)
	// ----------------------------------------------------------------------
	"InitAMM":    8_000,
	"Manager":    1_000,
	"CreatePool": 10_000,
	"Swap":       4_500,
	// AddLiquidity & RemoveLiquidity already defined above
	"Pool":  1_500,
	"Pools": 2_000,

	// ----------------------------------------------------------------------
	// Loan-Pool
	// ----------------------------------------------------------------------
	"NewLoanPool":            20_000,
	"Submit":                 3_000,
	"Disburse":               8_000,
	"Loanpool_GetProposal":   1_000,
	"Loanpool_ListProposals": 1_500,
	"Redistribute":           5_000,
	// Vote  & Tick already priced
	// RandomElectorate / IsAuthority already priced

	// ----------------------------------------------------------------------
	// Networking
	// ----------------------------------------------------------------------
	"NewNode":         18_000,
	"HandlePeerFound": 1_500,
	"DialSeed":        2_000,
	"ListenAndServe":  8_000,
	"Close":           500,
	"Peers":           400,
	"NewDialer":       2_000,
	"Dial":            2_000,
	"SetBroadcaster":  500,
	"GlobalBroadcast": 1_000,
	// Broadcast & Subscribe already priced

	// ----------------------------------------------------------------------
	// Replication / Data Availability
	// ----------------------------------------------------------------------
	"NewReplicator":  12_000,
	"ReplicateBlock": 30_000,
	"RequestMissing": 4_000,
	"Synchronize":    25_000,
	"Stop":           3_000,
	// Hash & Start already priced

	// ----------------------------------------------------------------------
	// Roll-ups
	// ----------------------------------------------------------------------
	"NewAggregator":     15_000,
	"SubmitBatch":       10_000,
	"SubmitFraudProof":  30_000,
	"FinalizeBatch":     10_000,
	"BatchHeader":       500,
	"BatchState":        300,
	"BatchTransactions": 1_000,
	"ListBatches":       2_000,

	// ----------------------------------------------------------------------
	// Security / Cryptography
	// ----------------------------------------------------------------------
	"AggregateBLSSigs":  7_000,
	"VerifyAggregated":  8_000,
	"CombineShares":     6_000,
	"ComputeMerkleRoot": 1_200,
	"Encrypt":           1_500,
	"Decrypt":           1_500,
	"NewTLSConfig":      5_000,
	"DilithiumKeypair":  6_000,
	"DilithiumSign":     5_000,
	"DilithiumVerify":   5_000,
	"PredictRisk":       2_000,
	"AnomalyScore":      2_000,

	// ----------------------------------------------------------------------
	// Sharding
	// ----------------------------------------------------------------------
	"NewShardCoordinator": 20_000,
	"SetLeader":           1_000,
	"Leader":              800,
	"SubmitCrossShard":    15_000,
	"Send":                2_000,
	"PullReceipts":        3_000,
	"Reshard":             30_000,
	"GossipTx":            5_000,
	"RebalanceShards":     8_000,
	"VerticalPartition":   2_000,
	// Broadcast already priced

	// ----------------------------------------------------------------------
	// Side-chains
	// ----------------------------------------------------------------------
	"InitSidechains":     12_000,
	"Sidechains":         600,
	"Sidechain_Register": 5_000,
	"SubmitHeader":       8_000,
	"VerifyWithdraw":     4_000,
	"VerifyAggregateSig": 8_000,
	"VerifyMerkleProof":  1_200,
	"GetSidechainMeta":   1_000,
	"ListSidechains":     1_200,
	"GetSidechainHeader": 1_000,
	// Deposit already priced

	// ----------------------------------------------------------------------
	// State-Channels
	// ----------------------------------------------------------------------
	"InitStateChannels":    8_000,
	"Channels":             600,
	"OpenChannel":          10_000,
	"VerifyECDSASignature": 2_000,
	"InitiateClose":        3_000,
	"Challenge":            4_000,
	"Finalize":             5_000,
	"GetChannel":           800,
	"ListChannels":         1_200,

	// ----------------------------------------------------------------------
	// Storage / Marketplace
	// ----------------------------------------------------------------------
	"NewStorage":    12_000,
	"CreateListing": 8_000,
	"Exists":        400,
	"OpenDeal":      5_000,
	"Create":        8_000, // generic create (non-AMM/non-contract)
	"CloseDeal":     5_000,
	"Release":       2_000,
	"GetListing":    1_000,
	"ListListings":  1_000,
	"GetDeal":       1_000,
	"ListDeals":     1_000,
	// Pin & Retrieve already priced

	// ----------------------------------------------------------------------
	// Token Standards (constants – zero-cost markers)
	// ----------------------------------------------------------------------
	"StdSYN10":   0,
	"StdSYN20":   0,
	"StdSYN70":   0,
	"StdSYN130":  0,
	"StdSYN131":  0,
	"StdSYN200":  0,
	"StdSYN223":  0,
	"StdSYN300":  0,
	"StdSYN500":  0,
	"StdSYN600":  0,
	"StdSYN700":  0,
	"StdSYN721":  0,
	"StdSYN722":  0,
	"StdSYN800":  0,
	"StdSYN845":  0,
	"StdSYN900":  0,
	"StdSYN1000": 0,
	"StdSYN1100": 0,
	"StdSYN1155": 0,
	"StdSYN1200": 0,
	"StdSYN1300": 0,
	"StdSYN1401": 0,
	"StdSYN1500": 0,
	"StdSYN1600": 0,
	"StdSYN1700": 0,
	"StdSYN1800": 0,
	"StdSYN1900": 0,
	"StdSYN1967": 0,
	"StdSYN2100": 0,
	"StdSYN2200": 0,
	"StdSYN2369": 0,
	"StdSYN2400": 0,
	"StdSYN2500": 0,
	"StdSYN2600": 0,
	"StdSYN2700": 0,
	"StdSYN2800": 0,
	"StdSYN2900": 0,
	"StdSYN3000": 0,
	"StdSYN3100": 0,
	"StdSYN3200": 0,
	"StdSYN3300": 0,
	"StdSYN3400": 0,
	"StdSYN3500": 0,
	"StdSYN3600": 0,
	"StdSYN3700": 0,
	"StdSYN3800": 0,
	"StdSYN3900": 0,
	"StdSYN4200": 0,
	"StdSYN4300": 0,
	"StdSYN4700": 0,
	"StdSYN4900": 0,
	"StdSYN5000": 0,

	// ----------------------------------------------------------------------
	// Token Utilities
	// ----------------------------------------------------------------------
	"ID":                400,
	"Meta":              400,
	"Allowance":         400,
	"Approve":           800,
	"Add":               600,
	"Sub":               600,
	"Get":               400,
	"transfer":          2_100, // lower-case ERC20 compatibility
	"Calculate":         800,
	"RegisterToken":     8_000,
	"NewBalanceTable":   5_000,
	"Set":               600,
	"RefundGas":         100,
	"PopUint32":         300,
	"PopAddress":        300,
	"PopUint64":         300,
	"PushBool":          300,
	"Push":              300,
	"Len":               200,
	"InitTokens":        8_000,
	"GetRegistryTokens": 400,

	// ----------------------------------------------------------------------
	// Transactions
	// ----------------------------------------------------------------------
	"VerifySig":      3_500,
	"ValidateTx":     5_000,
	"NewTxPool":      12_000,
	"AddTx":          6_000,
	"PickTxs":        1_500,
	"TxPoolSnapshot": 800,
	// Sign already priced

	// ----------------------------------------------------------------------
	// Low-level Math / Bitwise / Crypto opcodes
	// (values based on research into Geth & OpenEthereum plus Synnergy-specific
	//  micro-benchmarks – keep in mind that **all** word-size-dependent
	//  corrections are applied at run-time by the VM).
	// ----------------------------------------------------------------------
	"Short":            5,
	"BytesToAddress":   5,
	"Pop":              2,
	"opADD":            3,
	"opMUL":            5,
	"opSUB":            3,
	"OpDIV":            5,
	"opSDIV":           5,
	"opMOD":            5,
	"opSMOD":           5,
	"opADDMOD":         8,
	"opMULMOD":         8,
	"opEXP":            10,
	"opSIGNEXTEND":     5,
	"opLT":             3,
	"opGT":             3,
	"opSLT":            3,
	"opSGT":            3,
	"opEQ":             3,
	"opISZERO":         3,
	"opAND":            3,
	"opOR":             3,
	"opXOR":            3,
	"opNOT":            3,
	"opBYTE":           3,
	"opSHL":            3,
	"opSHR":            3,
	"opSAR":            3,
	"opECRECOVER":      700,
	"opEXTCODESIZE":    700,
	"opEXTCODECOPY":    700,
	"opEXTCODEHASH":    700,
	"opRETURNDATASIZE": 3,
	"opRETURNDATACOPY": 700,
	"opMLOAD":          3,
	"opMSTORE":         3,
	"opMSTORE8":        3,
	"opCALLDATALOAD":   3,
	"opCALLDATASIZE":   3,
	"opCALLDATACOPY":   700,
	"opCODESIZE":       3,
	"opCODECOPY":       700,
	"opJUMP":           8,
	"opJUMPI":          10,
	"opPC":             2,
	"opMSIZE":          2,
	"opGAS":            2,
	"opJUMPDEST":       1,
	"opSHA256":         60,
	"opKECCAK256":      30,
	"opRIPEMD160":      600,
	"opBLAKE2B256":     60,
	"opADDRESS":        2,
	"opCALLER":         2,
	"opORIGIN":         2,
	"opCALLVALUE":      2,
	"opGASPRICE":       2,
	"opNUMBER":         2,
	"opTIMESTAMP":      2,
	"opDIFFICULTY":     2,
	"opGASLIMIT":       2,
	"opCHAINID":        2,
	"opBLOCKHASH":      20,
	"opBALANCE":        400,
	"opSELFBALANCE":    5,
	"opLOG0":           375,
	"opLOG1":           750,
	"opLOG2":           1_125,
	"opLOG3":           1_500,
	"opLOG4":           1_875,
	"logN":             2_000,
	"opCREATE":         32_000,
	"opCALL":           700,
	"opCALLCODE":       700,
	"opDELEGATECALL":   700,
	"opSTATICCALL":     700,
	"opRETURN":         0,
	"opREVERT":         0,
	"opSTOP":           0,
	"opSELFDESTRUCT":   5_000,

	// Shared accounting ops
	"TransferVM": 2_100, // explicit VM variant (if separate constant exists)

	// ----------------------------------------------------------------------
	// Virtual Machine Internals
	// ----------------------------------------------------------------------
	"BurnVM":            2_100,
	"BurnLP":            2_100,
	"MintLP":            2_100,
	"NewInMemory":       500,
	"CallCode":          700,
	"CallContract":      700,
	"StaticCallVM":      700,
	"GetBalance":        400,
	"GetTokenBalance":   400,
	"SetTokenBalance":   500,
	"GetTokenSupply":    500,
	"SetBalance":        500,
	"DelegateCall":      700,
	"GetToken":          400,
	"NewMemory":         500,
	"Read":              3,
	"Write":             3,
	"LenVM":             3, // distinguish from token.Len if separate const
	"Call":              700,
	"SelectVM":          1_000,
	"CreateContract":    32_000,
	"AddLog":            375,
	"GetCode":           200,
	"GetCodeHash":       200,
	"MintTokenVM":       2_000,
	"PrefixIterator":    500,
	"NonceOf":           400,
	"GetState":          400,
	"SetState":          500,
	"HasState":          400,
	"DeleteState":       500,
	"NewGasMeter":       500,
	"SelfDestructVM":    5_000,
	"Remaining":         2,
	"Consume":           3,
	"ExecuteVM":         2_000,
	"NewSuperLightVM":   500,
	"NewLightVM":        800,
	"NewHeavyVM":        1_200,
	"ExecuteSuperLight": 1_000,
	"ExecuteLight":      1_500,
	"ExecuteHeavy":      2_000,

	// ----------------------------------------------------------------------
	// Wallet / Key-Management
	// ----------------------------------------------------------------------
	"NewRandomWallet":     10_000,
	"WalletFromMnemonic":  5_000,
	"NewHDWalletFromSeed": 6_000,
	"PrivateKey":          400,
	"NewAddress":          500,
	"SignTx":              3_000,

	// ----------------------------------------------------------------------
	// Biometrics Authentication
	// ----------------------------------------------------------------------
	"Bio_Enroll": 1_000,
	"Bio_Verify": 800,
	"Bio_Delete": 600,
}

func init() {
	gasTable = make(map[Opcode]uint64, len(gasNames))
	for name, cost := range gasNames {
		if op, ok := nameToOp[name]; ok {
			gasTable[op] = cost
		}
	}
}

// GasCost returns the **base** gas cost for a single opcode.  Dynamic portions
// (e.g. per-word fees, storage-touch refunds, call-stipends) are handled by the
// VM’s gas-meter layer.
//
// The function is lock-free and safe for concurrent use by every worker-thread
// in the execution engine.
func GasCost(op Opcode) uint64 {
	if cost, ok := gasTable[op]; ok {
		return cost
	}
	// Log only the first occurrence of an unknown opcode to avoid log spam.
	log.Printf("gas_table: missing cost for opcode %d – charging default", op)
	return DefaultGasCost
}
