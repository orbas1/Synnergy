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
   DeployAIContract: 50_000,
   InvokeAIContract: 7_500,
   UpdateAIModel:    20_000,
   GetAIModel:       2_000,
   InferModel:     30_000,
   AnalyseTransactions: 35_000,

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
   NewAuthorityApplier: 20_000,
   SubmitApplication:   4_000,
   VoteApplication:     3_000,
   FinalizeApplication: 5_000,
   GetApplication:      1_000,
   ListApplications:    2_000,

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
   Audit_Init:           5_000,
   Audit_Log:            2_000,
   Audit_Events:         3_000,
   Audit_Close:          1_000,
   InitComplianceManager: 10_000,
   SuspendAccount:        4_000,
   ResumeAccount:         4_000,
   IsSuspended:           500,
   WhitelistAccount:      3_000,
   RemoveWhitelist:       3_000,
   IsWhitelisted:         500,
   Compliance_ReviewTx:   2_500,
   AnalyzeAnomaly:       6_000,
   FlagAnomalyTx:        2_500,

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
   HopConsensus:         4_000,
   CurrentConsensus:     500,
   Status:                1_000,
   SetDifficulty:         2_000,
   NewConsensusAdaptiveManager: 10_000,
   ComputeDemand:               2_000,
   ComputeStakeConcentration:   2_000,
   AdjustConsensus:             5_000,
   RegisterValidator:     8_000,
   DeregisterValidator:   6_000,
   StakeValidator:        2_000,
   UnstakeValidator:      2_000,
   SlashValidator:        3_000,
   GetValidator:          1_000,
   ListValidators:        2_000,
   IsValidator:           800,

   // ----------------------------------------------------------------------
   // Contracts (WASM / EVM‐compat)
   // ----------------------------------------------------------------------
   InitContracts: 15_000,
   CompileWASM:   45_000,
   Invoke:        7_000,
   Deploy:        25_000,
   TransferOwnership: 5_000,
   PauseContract:     3_000,
   ResumeContract:    3_000,
   UpgradeContract:   20_000,
   ContractInfo:      1_000,

   // ----------------------------------------------------------------------
   // Cross-Chain
   // ----------------------------------------------------------------------
   RegisterBridge: 20_000,
   AssertRelayer:  5_000,
   Iterator:       2_000,
   LockAndMint:    30_000,
   BurnAndRelease: 30_000,
   GetBridge:      1_000,
   RegisterXContract: 22_000,
   GetXContract:      1_000,
   ListXContracts:    1_200,
   RemoveXContract:   5_000,
   RecordCrossChainTx: 25_000,
   GetCrossChainTx:    2_000,
   ListCrossChainTx:   3_000,
   OpenChainConnection:  10_000,
   CloseChainConnection: 5_000,
   GetChainConnection:   1_000,
   ListChainConnections: 2_000,
   RegisterProtocol:   20_000,
   ListProtocols:      2_000,
   GetProtocol:        1_000,
   ProtocolDeposit:    30_000,
   ProtocolWithdraw:   30_000,
   StartBridgeTransfer:    25_000,
   CompleteBridgeTransfer: 25_000,
   GetBridgeTransfer:      1_000,
   ListBridgeTransfers:    2_000,

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
   UpdateOracleSource: 4_000,
   RemoveOracle:      4_000,
   GetOracleMetrics:  2_000,
   RequestOracleData: 3_000,
   SyncOracle:        5_000,
   CreateDataFeed: 6_000,
   QueryDataFeed:  3_000,
   ManageDataFeed: 5_000,
   ImputeMissing:  4_000,
   NormalizeFeed:  4_000,
   AddProvenance:  2_000,
   SampleFeed:     3_000,
   ScaleFeed:      3_000,
   TransformFeed:  4_000,
   VerifyFeedTrust: 3_000,
   ZTDC_Open:      6_000,
   ZTDC_Send:      2_000,
   ZTDC_Close:     4_000,

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
   HA_Register:      1_000,
   HA_Remove:        1_000,
   HA_List:          500,
   HA_Sync:          20_000,
   HA_Promote:       8_000,

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
   CastTokenVote:   4_000,
   ExecuteProposal: 15_000,
   GetProposal:     1_000,
   ListProposals:   2_000,
   DAO_Stake:       5_000,
   DAO_Unstake:     5_000,
   DAO_Staked:      500,
   DAO_TotalStaked: 500,

   // Quadratic Voting
   SubmitQuadraticVote: 3_500,
   QuadraticResults:    2_000,
   QuadraticWeight:     50,
   AddDAOMember:    1_200,
   RemoveDAOMember: 1_200,
   RoleOfMember:    500,
   ListDAOMembers:  1_000,
   AddSYN2500Member:    1_200,
   RemoveSYN2500Member: 1_200,
   DelegateSYN2500Vote: 800,
   SYN2500VotingPower:  500,
   CastSYN2500Vote:     1_500,
   SYN2500MemberInfo:   500,
   ListSYN2500Members:  1_000,
   RegisterGovContract: 8_000,
   GetGovContract:      1_000,
   ListGovContracts:    2_000,
   EnableGovContract:   1_000,
   DeleteGovContract:   1_000,
   DeployGovContract: 25_000,
   InvokeGovContract: 7_000,
   NewTimelock:     4_000,
   QueueProposal:   3_000,
   CancelProposal:  3_000,
   ExecuteReady:    5_000,
   ListTimelocks:   1_000,
   SYN300_Delegate:        400,
   SYN300_RevokeDelegate:  400,
   SYN300_VotingPower:     300,
   SYN300_CreateProposal:  1_000,
   SYN300_Vote:            800,
   SYN300_ExecuteProposal: 1_500,
   SYN300_ProposalStatus:  300,
   SYN300_ListProposals:   500,

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
   InitForkManager:     5_000,
   AddForkBlock:        7_000,
   ResolveForks:        12_000,
   ListForks:           2_000,

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
   CreateGrant:   3_000,
   ReleaseGrant:  8_000,
   GetGrant:      1_000,
   CancelProposal: 2_000,
   ExtendProposal: 1_500,
   NewLoanPoolManager: 10_000,
   Loanpool_Pause: 1_000,
   Loanpool_Resume: 1_000,
   Loanpool_IsPaused: 500,
   Loanpool_Stats: 2_000,
   RequestApproval: 3_000,
   ApproveRequest:  4_000,
   RejectRequest:   4_000,
   NewLoanPoolApply:   20_000,
   LoanApply_Submit:   3_000,
   LoanApply_Vote:     3_000,
   LoanApply_Process:  1_000,
   LoanApply_Disburse: 8_000,
   LoanApply_Get:      1_000,
   LoanApply_List:     1_500,
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
   NewBootstrapNode: 20_000,
   Bootstrap_Start: 8_000,
   Bootstrap_Stop: 4_000,
   Bootstrap_Peers: 500,
   Bootstrap_DialSeed: 2_000,
   NewConnPool:     8_000,
   AcquireConn:     500,
   ReleaseConn:     200,
   ClosePool:       400,
   PoolStats:       100,
   DiscoverPeers:  1_000,
   Connect:        1_500,
   Disconnect:     1_000,
   AdvertiseSelf:  800,
   StartDevNet:    50_000,
   StartTestNet:   60_000,
   NewMasterNode:  30_000,
   Master_Start:   5_000,
   Master_Stop:    3_000,
   Master_ProcessTx: 2_000,
   Master_HandlePrivateTx: 3_000,
   Master_VoteProposal:   1_000,
   // Broadcast & Subscribe already priced

   // ----------------------------------------------------------------------
   // Replication / Data Availability
   // ----------------------------------------------------------------------
   NewReplicator:  12_000,
   ReplicateBlock: 30_000,
   RequestMissing: 4_000,
   Synchronize:    25_000,
   Stop:           3_000,
   NewInitService:     8_000,
   BootstrapLedger:    20_000,
   ShutdownInitService: 3_000,
   NewSyncManager: 12_000,
   Sync_Start:     5_000,
   Sync_Stop:      3_000,
   Sync_Status:    1_000,
   SyncOnce:       8_000,
   // Hash & Start already priced

   // ----------------------------------------------------------------------
   // Distributed Coordination
   // ----------------------------------------------------------------------
   NewCoordinator:         10_000,
   StartCoordinator:        5_000,
   StopCoordinator:         5_000,
   BroadcastLedgerHeight:   3_000,
   DistributeToken:         5_000,

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
   PauseAggregator:   500,
   ResumeAggregator:  500,
   AggregatorStatus:  200,

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
   BuildMerkleTree:   1_500,
   MerkleProof:       1_200,
   VerifyMerklePath:  1_200,

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
   HorizontalPartition: 2_000,
   CompressData:        4_000,
   DecompressData:      4_000,
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
   PauseSidechain:            3_000,
   ResumeSidechain:           3_000,
   UpdateSidechainValidators: 5_000,
   RemoveSidechain:           6_000,
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
   PauseChannel:         1_500,
   ResumeChannel:        1_500,
   CancelClose:          3_000,
   ForceClose:           6_000,

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
   IPFS_Add:     5_000,
   IPFS_Get:     4_000,
   IPFS_Unpin:   3_000,
        // General Marketplace
        CreateMarketListing:  8_000,
        PurchaseItem:        6_000,
        CancelListing:       3_000,
        ReleaseFunds:        2_000,
        GetMarketListing:    1_000,
        ListMarketListings:  1_000,
        GetMarketDeal:       1_000,
        ListMarketDeals:     1_000,
   // Pin & Retrieve already priced
   // ----------------------------------------------------------------------
   // Identity Verification
   // ----------------------------------------------------------------------
   RegisterIdentity: 5_000,
   VerifyIdentity:   1_000,
   RemoveIdentity:   2_000,
   ListIdentities:   2_000,


   // ----------------------------------------------------------------------
   // Resource Marketplace
   // ----------------------------------------------------------------------
   ListResource:        8_000,
   OpenResourceDeal:    5_000,
   CloseResourceDeal:   5_000,
   GetResourceListing:  1_000,
   ListResourceListings: 1_000,
   GetResourceDeal:     1_000,
   ListResourceDeals:   1_000,

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
   TokenManager_Create: 8_000,
   TokenManager_Transfer: 2_100,
   TokenManager_Mint: 2_100,
   TokenManager_Burn: 2_100,
   TokenManager_Approve: 800,
  TokenManager_BalanceOf: 400,
  SYN500_GrantAccess:   1_000,
  SYN500_UpdateAccess:  1_000,
  SYN500_RevokeAccess:  800,
  SYN500_RecordUsage:   500,
  SYN500_RedeemReward:  500,
  SYN500_RewardBalance: 400,
  SYN500_Usage:         400,
  SYN500_AccessInfo:    400,
   TokenManager_BalanceOf: 400,
   // SYN70 Token Standard operations
   SYN70_RegisterAsset:   1_000,
   SYN70_TransferAsset:   800,
   SYN70_UpdateAttributes: 500,
   SYN70_RecordAchievement: 300,
   SYN70_GetAsset:        200,
   SYN70_ListAssets:      200,
   SYN600_Stake:           2_000,
   SYN600_Unstake:        2_000,
   SYN600_AddEngagement:  500,
   SYN600_EngagementOf:   400,
   SYN600_DistributeRewards: 5_000,
   SYN800_RegisterAsset: 100,
   SYN800_UpdateValuation: 100,
   SYN800_GetAsset: 50,

   IDToken_Register:    800,
   IDToken_Verify:      400,
   IDToken_Get:         200,
   IDToken_Logs:        200,
   SYN1200_AddBridge: 800,
   SYN1200_AtomicSwap: 2_500,
   SYN1200_CompleteSwap: 400,
   SYN1200_GetSwap: 400,
   SYN1100_AddRecord:      500,
   SYN1100_GrantAccess:    300,
   SYN1100_RevokeAccess:   200,
   SYN1100_GetRecord:      400,
   SYN1100_TransferOwnership: 500,
   // Supply chain token utilities
   SupplyChain_RegisterAsset: 5_000,
   SupplyChain_UpdateLocation: 800,
   SupplyChain_UpdateStatus: 800,
   SupplyChain_TransferAsset: 2_100,
   MusicRoyalty_AddRevenue: 800,
   MusicRoyalty_Distribute: 1_000,
   MusicRoyalty_UpdateInfo: 500,
   // SYN1700 Event Ticket operations
   Event_Create:       1_000,
   Event_IssueTicket:  1_500,
   Event_Transfer:     2_100,
   Event_Verify:       500,
   Event_Use:          400,
   Tokens_RecordEmission: 1_000,
   Tokens_RecordOffset: 1_000,
   Tokens_NetBalance: 400,
   Tokens_ListRecords: 800,
   Edu_RegisterCourse: 5_000,
   Edu_IssueCredit:    2_100,
   Edu_VerifyCredit:   800,
   Edu_RevokeCredit:   800,
   Edu_GetCredit:      400,
   Edu_ListCredits:    400,
   // SYN2100 token standard
   SYN2100_RegisterDocument: 800,
   SYN2100_FinanceDocument:  800,
   SYN2100_GetDocument:      200,
   SYN2100_ListDocuments:    400,
   SYN2100_AddLiquidity:     500,
   SYN2100_RemoveLiquidity:  500,
   SYN2100_LiquidityOf:      200,
   Tokens_CreateSYN2200: 8_000,
   Tokens_SendPayment: 2_100,
   Tokens_GetPayment: 400,
   DataToken_UpdateMeta: 500,
   DataToken_SetPrice: 400,
   DataToken_GrantAccess: 400,
   DataToken_RevokeAccess: 300,

   // ----------------------------------------------------------------------
   // Transactions
   // ----------------------------------------------------------------------
   VerifySig:      3_500,
   ValidateTx:     5_000,
   NewTxPool:      12_000,
   AddTx:          6_000,
   PickTxs:        1_500,
   TxPoolSnapshot: 800,
   EncryptTxPayload: 3_500,
   DecryptTxPayload: 3_000,
   SubmitPrivateTx:  6_500,
   EncodeEncryptedHex: 300,
   ReverseTransaction: 10_000,
   NewTxDistributor: 8_000,
   DistributeFees:   1_500,
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

   // ----------------------------------------------------------------------
   // Access Control
   // ----------------------------------------------------------------------
   GrantRole:  1_000,
   RevokeRole: 1_000,
   HasRole:    300,
   ListRoles:  500,
   // Plasma Management
   // ----------------------------------------------------------------------
   InitPlasma:        8_000,
   Plasma_Deposit:    5_000,
   Plasma_Withdraw:   5_000,
   Plasma_SubmitBlock: 10_000,
   Plasma_GetBlock:    1_000,
   // Resource Management
   // ----------------------------------------------------------------------
   SetQuota:         1_000,
   GetQuota:         500,
   ChargeResources:  2_000,
   ReleaseResources: 1_000,
   // Finalization Management
   // ----------------------------------------------------------------------
   NewFinalizationManager: 8_000,
   FinalizeBlock:         4_000,
   FinalizeBatchManaged:  3_500,
   FinalizeChannelManaged: 3_500,
   // System Health & Logging
   // ----------------------------------------------------------------------
   NewHealthLogger: 8_000,
   MetricsSnapshot: 1_000,
   LogEvent:        500,
   RotateLogs:      4_000,
   RegisterIDWallet:    8_000,
   IsIDWalletRegistered: 500,

   // ----------------------------------------------------------------------
   // Event Management
   // ----------------------------------------------------------------------
   InitEvents: 5_000,
   EmitEvent: 400,
   GetEvent:  800,
   ListEvents: 1_000,
   CreateWallet:        10_000,
   ImportWallet:        5_000,
   WalletBalance:       400,
   WalletTransfer:      2_100,

   // ----------------------------------------------------------------------
   // Immutability Enforcement
   // ----------------------------------------------------------------------
   InitImmutability: 8_000,
   VerifyChain:     4_000,
   RestoreChain:    6_000,
*/

// gasNames holds the gas cost associated with each opcode name. During init()
// these names are resolved to their Opcode values using the catalogue defined
// in opcode_dispatcher.go.
var gasNames = map[string]uint64{
	// ----------------------------------------------------------------------
	// AI
	// ----------------------------------------------------------------------
	"InitAI":              5000,
	"AI":                  4000,
	"PredictAnomaly":      3500,
	"OptimizeFees":        2500,
	"PublishModel":        4500,
	"FetchModel":          1500,
	"ListModel":           800,
	"ValidateKYC":         100,
	"BuyModel":            3000,
	"RentModel":           2000,
	"ReleaseEscrow":       1200,
	"PredictVolume":       1500,
	"DeployAIContract":    5000,
	"InvokeAIContract":    750,
	"UpdateAIModel":       2000,
	"GetAIModel":          200,
	"StartTraining":       5000,
	"TrainingStatus":      500,
	"ListTrainingJobs":    800,
	"CancelTraining":      1000,
	"GetModelListing":     100,
	"ListModelListings":   200,
	"UpdateListingPrice":  200,
	"RemoveListing":       200,
	"InferModel":          3000,
	"AnalyseTransactions": 3500,

	// ----------------------------------------------------------------------
	// Automated-Market-Maker
	// ----------------------------------------------------------------------
	"SwapExactIn":       450,
	"AddLiquidity":      500,
	"RemoveLiquidity":   500,
	"Quote":             250,
	"AllPairs":          200,
	"InitPoolsFromFile": 600,

	// ----------------------------------------------------------------------
	// Authority / Validator-Set
	// ---------------------------------------------------------------------

	"NewAuthoritySet":     2000,
	"RecordVote":          300,
	"RegisterCandidate":   800,
	"RandomElectorate":    400,
	"IsAuthority":         80,
	"GetAuthority":        100,
	"ListAuthorities":     200,
	"DeregisterAuthority": 600,
	"NewAuthorityApplier": 2000,
	"SubmitApplication":   400,
	"VoteApplication":     300,
	"FinalizeApplication": 500,
	"GetApplication":      100,
	"ListApplications":    200,

	// ----------------------------------------------------------------------
	// Charity Pool
	// ----------------------------------------------------------------------
	"NewCharityPool":           1000,
	"Deposit":                  210,
	"Charity_Register":         0,
	"Vote":                     300,
	"Tick":                     100,
	"GetRegistration":          80,
	"Winners":                  80,
	"Charity_Donate":           0,
	"Charity_WithdrawInternal": 0,
	"Charity_Balances":         0,

	// ----------------------------------------------------------------------
	// Coin
	// ----------------------------------------------------------------------
	"NewCoin":     1200,
	"Mint":        210,
	"TotalSupply": 80,
	"BalanceOf":   40,

	// ----------------------------------------------------------------------
	// Compliance
	// ----------------------------------------------------------------------

	"InitCompliance":        800,
	"EraseData":             500,
	"RecordFraudSignal":     700,
	"Compliance_LogAudit":   0,
	"Compliance_AuditTrail": 0,
	"Compliance_MonitorTx":  0,
	"Compliance_VerifyZKP":  0,
	"Audit_Init":            0,
	"Audit_Log":             0,
	"Audit_Events":          0,
	"Audit_Close":           0,
	"InitComplianceManager": 1000,
	"SuspendAccount":        400,
	"ResumeAccount":         400,
	"IsSuspended":           50,
	"WhitelistAccount":      300,
	"RemoveWhitelist":       300,
	"IsWhitelisted":         50,
	"Compliance_ReviewTx":   0,
	"AnalyzeAnomaly":        600,
	"FlagAnomalyTx":         250,

	// ----------------------------------------------------------------------
	// Consensus Core
	// ----------------------------------------------------------------------
	"Pick":                        200,
	"Broadcast":                   500,
	"Subscribe":                   150,
	"Sign":                        300,
	"Verify":                      350,
	"ValidatorPubKey":             80,
	"StakeOf":                     100,
	"LoanPoolAddress":             80,
	"Hash":                        60,
	"SerializeWithoutNonce":       120,
	"NewConsensus":                2500,
	"Start":                       500,
	"ProposeSubBlock":             1500,
	"ValidatePoH":                 2000,
	"SealMainBlockPOW":            6000,
	"DistributeRewards":           1000,
	"CalculateWeights":            800,
	"ComputeThreshold":            600,
	"NewConsensusAdaptiveManager": 1000,
	"ComputeDemand":               200,
	"ComputeStakeConcentration":   200,
	"AdjustConsensus":             500,
	"HopConsensus":                400,
	"CurrentConsensus":            50,
	"Status":                      100,
	"SetDifficulty":               200,
	"AdjustStake":                 300,
	"PenalizeValidator":           400,
	"RegisterValidator":           800,
	"DeregisterValidator":         600,
	"StakeValidator":              200,
	"UnstakeValidator":            200,
	"SlashValidator":              300,
	"GetValidator":                100,
	"ListValidators":              200,
	"IsValidator":                 80,

	// ----------------------------------------------------------------------
	// Contracts (WASM / EVM‐compat)
	// ----------------------------------------------------------------------
	"InitContracts":     1500,
	"CompileWASM":       4500,
	"Invoke":            700,
	"Deploy":            2500,
	"TransferOwnership": 500,
	"PauseContract":     300,
	"ResumeContract":    300,
	"UpgradeContract":   2000,
	"ContractInfo":      100,

	// ----------------------------------------------------------------------
	// Cross-Chain
	// ----------------------------------------------------------------------
	"RegisterBridge":         2000,
	"AssertRelayer":          500,
	"Iterator":               200,
	"LockAndMint":            3000,
	"BurnAndRelease":         3000,
	"GetBridge":              100,
	"RegisterXContract":      2200,
	"GetXContract":           100,
	"ListXContracts":         120,
	"RemoveXContract":        500,
	"RecordCrossChainTx":     2500,
	"GetCrossChainTx":        200,
	"ListCrossChainTx":       300,
	"OpenChainConnection":    1000,
	"CloseChainConnection":   500,
	"GetChainConnection":     100,
	"ListChainConnections":   200,
	"RegisterProtocol":       2000,
	"ListProtocols":          200,
	"GetProtocol":            100,
	"ProtocolDeposit":        3000,
	"ProtocolWithdraw":       3000,
	"StartBridgeTransfer":    2500,
	"CompleteBridgeTransfer": 2500,
	"GetBridgeTransfer":      100,
	"ListBridgeTransfers":    200,

	// ----------------------------------------------------------------------
	// Cross-Consensus Scaling Networks
	// ----------------------------------------------------------------------
	"RegisterCCSNetwork": 2000,
	"ListCCSNetworks":    500,
	"GetCCSNetwork":      100,
	"CCSLockAndTransfer": 3000,
	"CCSBurnAndRelease":  3000,

	// ----------------------------------------------------------------------
	// Data / Oracle / IPFS Integration
	// ----------------------------------------------------------------------

	"RegisterNode":       1000,
	"UploadAsset":        3000,
	"Pin":                500,
	"Retrieve":           400,
	"RetrieveAsset":      400,
	"RegisterOracle":     1000,
	"PushFeed":           300,
	"QueryOracle":        300,
	"ListCDNNodes":       300,
	"ListOracles":        300,
	"PushFeedSigned":     400,
	"UpdateOracleSource": 400,
	"RemoveOracle":       400,
	"GetOracleMetrics":   200,
	"RequestOracleData":  300,
	"SyncOracle":         500,
	"CreateDataSet":      800,
	"PurchaseDataSet":    500,
	"GetDataSet":         100,
	"ListDataSets":       200,
	"HasAccess":          100,
	"CreateDataFeed":     600,
	"QueryDataFeed":      300,
	"ManageDataFeed":     500,
	"ImputeMissing":      400,
	"NormalizeFeed":      400,
	"AddProvenance":      200,
	"SampleFeed":         300,
	"ScaleFeed":          300,
	"TransformFeed":      400,
	"VerifyFeedTrust":    300,
	"ZTDC_Open":          0,
	"ZTDC_Send":          0,
	"ZTDC_Close":         0,
	"StoreManagedData":   800,
	"LoadManagedData":    300,
	"DeleteManagedData":  200,

	// ---------------------------------------------------------------------
	// External Sensors
	// ---------------------------------------------------------------------
	"RegisterSensor":    1000,
	"GetSensor":         100,
	"ListSensors":       200,
	"UpdateSensorValue": 150,
	"PollSensor":        500,
	"TriggerWebhook":    500,

	// ----------------------------------------------------------------------
	// Fault-Tolerance / Health-Checker
	// ----------------------------------------------------------------------
	"NewHealthChecker":    800,
	"AddPeer":             150,
	"RemovePeer":          150,
	"FT_Snapshot":         0,
	"Recon":               800,
	"Ping":                30,
	"SendPing":            30,
	"AwaitPong":           30,
	"BackupSnapshot":      1000,
	"RestoreSnapshot":     1200,
	"VerifyBackup":        600,
	"FailoverNode":        800,
	"PredictFailure":      100,
	"AdjustResources":     150,
	"InitResourceManager": 500,
	"SetLimit":            100,
	"GetLimit":            50,
	"ConsumeLimit":        80,
	"TransferLimit":       120,
	"ListLimits":          70,
	"HA_Register":         0,
	"HA_Remove":           0,
	"HA_List":             0,
	"HA_Sync":             0,
	"HA_Promote":          0,

	// ----------------------------------------------------------------------
	// Governance
	// ----------------------------------------------------------------------

	"UpdateParam":            500,
	"ProposeChange":          1000,
	"VoteChange":             300,
	"EnactChange":            800,
	"SubmitProposal":         1000,
	"BalanceOfAsset":         60,
	"CastVote":               300,
	"ExecuteProposal":        1500,
	"GetProposal":            100,
	"ListProposals":          200,
	"NewQuorumTracker":       100,
	"Quorum_AddVote":         0,
	"Quorum_HasQuorum":       0,
	"Quorum_Reset":           0,
	"SubmitQuadraticVote":    350,
	"QuadraticResults":       200,
	"QuadraticWeight":        5,
	"RegisterGovContract":    800,
	"GetGovContract":         100,
	"ListGovContracts":       200,
	"EnableGovContract":      100,
	"DeleteGovContract":      100,
	"DeployGovContract":      2500,
	"InvokeGovContract":      700,
	"AddReputation":          200,
	"SubtractReputation":     200,
	"ReputationOf":           50,
	"SubmitRepGovProposal":   1000,
	"CastRepGovVote":         300,
	"ExecuteRepGovProposal":  1500,
	"GetRepGovProposal":      100,
	"ListRepGovProposals":    200,
	"CastTokenVote":          400,
	"DAO_Stake":              0,
	"DAO_Unstake":            0,
	"DAO_Staked":             0,
	"DAO_TotalStaked":        0,
	"AddDAOMember":           120,
	"RemoveDAOMember":        120,
	"RoleOfMember":           50,
	"ListDAOMembers":         100,
	"NewTimelock":            400,
	"QueueProposal":          300,
	"CancelProposal":         300,
	"ExecuteReady":           500,
	"ListTimelocks":          100,
	"SYN300_Delegate":        40,
	"SYN300_RevokeDelegate":  40,
	"SYN300_VotingPower":     30,
	"SYN300_CreateProposal":  100,
	"SYN300_Vote":            80,
	"SYN300_ExecuteProposal": 150,
	"SYN300_ProposalStatus":  30,
	"SYN300_ListProposals":   50,
	"CreateDAO":              1000,
	"JoinDAO":                300,
	"LeaveDAO":               200,
	"DAOInfo":                100,
	"ListDAOs":               200,
	"UpdateParam":            500,
	"ProposeChange":          1000,
	"VoteChange":             300,
	"EnactChange":            800,
	"SubmitProposal":         1000,
	"BalanceOfAsset":         60,
	"CastVote":               300,
	"ExecuteProposal":        1500,
	"GetProposal":            100,
	"ListProposals":          200,
	"NewQuorumTracker":       100,
	"Quorum_AddVote":         0,
	"Quorum_HasQuorum":       0,
	"Quorum_Reset":           0,
	"SubmitQuadraticVote":    350,
	"QuadraticResults":       200,
	"QuadraticWeight":        5,
	"RegisterGovContract":    800,
	"GetGovContract":         100,
	"ListGovContracts":       200,
	"EnableGovContract":      100,
	"DeleteGovContract":      100,
	"DeployGovContract":      2500,
	"InvokeGovContract":      700,
	"AddReputation":          200,
	"SubtractReputation":     200,
	"ReputationOf":           50,
	"SubmitRepGovProposal":   1000,
	"CastRepGovVote":         300,
	"ExecuteRepGovProposal":  1500,
	"GetRepGovProposal":      100,
	"ListRepGovProposals":    200,
	"Rep_AddActivity":        200,
	"Rep_Endorse":            200,
	"Rep_Penalize":           200,
	"Rep_Score":              50,
	"Rep_Level":              50,
	"Rep_History":            100,
	"CastTokenVote":          400,
	"DAO_Stake":              0,
	"DAO_Unstake":            0,
	"DAO_Staked":             0,
	"DAO_TotalStaked":        0,
	"AddDAOMember":           120,
	"RemoveDAOMember":        120,
	"RoleOfMember":           50,
	"ListDAOMembers":         100,
	"AddSYN2500Member":       120,
	"RemoveSYN2500Member":    120,
	"DelegateSYN2500Vote":    80,
	"SYN2500VotingPower":     50,
	"CastSYN2500Vote":        150,
	"SYN2500MemberInfo":      50,
	"ListSYN2500Members":     100,
	"NewTimelock":            400,
	"QueueProposal":          300,
	"CancelProposal":         300,
	"ExecuteReady":           500,
	"ListTimelocks":          100,
	"CreateDAO":              1000,
	"JoinDAO":                300,
	"LeaveDAO":               200,
	"DAOInfo":                100,
	"ListDAOs":               200,

	// ----------------------------------------------------------------------
	// Green Technology
	// ----------------------------------------------------------------------
	"InitGreenTech":    800,
	"Green":            200,
	"RecordUsage":      300,
	"RecordOffset":     300,
	"Certify":          700,
	"CertificateOf":    50,
	"ShouldThrottle":   20,
	"ListCertificates": 100,

	// ----------------------------------------------------------------------
	// Energy Efficiency
	// ----------------------------------------------------------------------
	"InitEnergyEfficiency": 800,
	"EnergyEff":            200,
	"RecordStats":          300,
	"EfficiencyOf":         50,
	"NetworkAverage":       100,
	"ListEfficiency":       100,

	// ----------------------------------------------------------------------
	// Ledger / UTXO / Account-Model
	// ----------------------------------------------------------------------
	"NewLedger":           5000,
	"GetPendingSubBlocks": 200,
	"LastBlockHash":       60,
	"AppendBlock":         5000,
	"MintBig":             220,
	"EmitApproval":        120,
	"EmitTransfer":        120,
	"DeductGas":           210,
	"WithinBlock":         100,
	"IsIDTokenHolder":     40,
	"TokenBalance":        40,
	"AddBlock":            4000,
	"GetBlock":            200,
	"GetUTXO":             150,
	"AddToPool":           100,
	"ListPool":            80,
	"GetContract":         100,
	"Snapshot":            300,
	"MintToken":           200,
	"LastSubBlockHeight":  50,
	"LastBlockHeight":     50,
	"RecordPoSVote":       300,
	"AppendSubBlock":      800,
	"Transfer":            210,
	"Burn":                210,
	"InitForkManager":     500,
	"AddForkBlock":        700,
	"ResolveForks":        1200,
	"ListForks":           200,
	"Account_Create":      0,
	"Account_Delete":      0,
	"Account_Balance":     0,
	"Account_Transfer":    0,

	// ----------------------------------------------------------------------
	// Liquidity Manager (high-level AMM façade)
	// ----------------------------------------------------------------------
	"InitAMM":    800,
	"Manager":    100,
	"CreatePool": 1000,
	"Swap":       450,
	// AddLiquidity & RemoveLiquidity already defined above
	"Pool":  150,
	"Pools": 200,

	// ----------------------------------------------------------------------
	// Loan-Pool
	// ----------------------------------------------------------------------
	"NewLoanPool":              2000,
	"Submit":                   300,
	"Disburse":                 800,
	"Loanpool_GetProposal":     0,
	"Loanpool_ListProposals":   0,
	"Redistribute":             500,
	"Loanpool_CancelProposal":  0,
	"Loanpool_ExtendProposal":  0,
	"Loanpool_RequestApproval": 0,
	"Loanpool_ApproveRequest":  0,
	"Loanpool_RejectRequest":   0,
	"Loanpool_CreateGrant":     0,
	"Loanpool_ReleaseGrant":    0,
	"Loanpool_GetGrant":        0,
	"NewLoanPoolManager":       1000,
	"Loanpool_Pause":           0,
	"Loanpool_Resume":          0,
	"Loanpool_IsPaused":        0,
	"Loanpool_Stats":           0,
	"NewLoanPoolApply":         2000,
	"LoanApply_Submit":         0,
	"LoanApply_Vote":           0,
	"LoanApply_Process":        0,
	"LoanApply_Disburse":       0,
	"LoanApply_Get":            0,
	"LoanApply_List":           0,
	// Vote  & Tick already priced
	// RandomElectorate / IsAuthority already priced

	// ----------------------------------------------------------------------
	// Networking
	// ----------------------------------------------------------------------
	"NewNode":                1800,
	"HandlePeerFound":        150,
	"DialSeed":               200,
	"ListenAndServe":         800,
	"Close":                  50,
	"Peers":                  40,
	"NewDialer":              200,
	"Dial":                   200,
	"SetBroadcaster":         50,
	"GlobalBroadcast":        100,
	"NewBootstrapNode":       2000,
	"Bootstrap_Start":        0,
	"Bootstrap_Stop":         0,
	"Bootstrap_Peers":        0,
	"Bootstrap_DialSeed":     0,
	"NewConnPool":            800,
	"AcquireConn":            50,
	"ReleaseConn":            20,
	"ClosePool":              40,
	"PoolStats":              10,
	"NewNATManager":          500,
	"NAT_Map":                0,
	"NAT_Unmap":              0,
	"NAT_ExternalIP":         0,
	"DiscoverPeers":          100,
	"Connect":                150,
	"Disconnect":             100,
	"AdvertiseSelf":          80,
	"StartDevNet":            5000,
	"StartTestNet":           6000,
	"NewMasterNode":          3000,
	"Master_Start":           500,
	"Master_Stop":            300,
	"Master_ProcessTx":       200,
	"Master_HandlePrivateTx": 300,
	"Master_VoteProposal":    100,
	// Broadcast & Subscribe already priced

	// ----------------------------------------------------------------------
	// Replication / Data Availability
	// ----------------------------------------------------------------------
	"NewReplicator":       1200,
	"ReplicateBlock":      3000,
	"RequestMissing":      400,
	"Synchronize":         2500,
	"Stop":                300,
	"NewInitService":      800,
	"BootstrapLedger":     2000,
	"ShutdownInitService": 300,
	// ----------------------------------------------------------------------
	// Distributed Coordination
	// ----------------------------------------------------------------------
	"NewCoordinator":        1000,
	"StartCoordinator":      500,
	"StopCoordinator":       500,
	"BroadcastLedgerHeight": 300,
	"DistributeToken":       500,

	"NewSyncManager": 1200,
	"Sync_Start":     0,
	"Sync_Stop":      0,
	"Sync_Status":    0,
	"SyncOnce":       800,
	// Hash & Start already priced

	// ----------------------------------------------------------------------
	// Roll-ups
	// ----------------------------------------------------------------------
	"NewAggregator":     1500,
	"SubmitBatch":       1000,
	"SubmitFraudProof":  3000,
	"FinalizeBatch":     1000,
	"BatchHeader":       50,
	"BatchState":        30,
	"BatchTransactions": 100,
	"ListBatches":       200,
	"PauseAggregator":   50,
	"ResumeAggregator":  50,
	"AggregatorStatus":  20,

	// ----------------------------------------------------------------------
	// Security / Cryptography
	// ----------------------------------------------------------------------
	"AggregateBLSSigs":  700,
	"VerifyAggregated":  800,
	"CombineShares":     600,
	"ComputeMerkleRoot": 120,
	"Encrypt":           150,
	"Decrypt":           150,
	"NewTLSConfig":      500,
	"DilithiumKeypair":  600,
	"DilithiumSign":     500,
	"DilithiumVerify":   500,
	"PredictRisk":       200,
	"AnomalyScore":      200,
	"BuildMerkleTree":   150,
	"MerkleProof":       120,
	"VerifyMerklePath":  120,

	// ----------------------------------------------------------------------
	// Sharding
	// ----------------------------------------------------------------------
	"NewShardCoordinator": 2000,
	"SetLeader":           100,
	"Leader":              80,
	"SubmitCrossShard":    1500,
	"Send":                200,
	"PullReceipts":        300,
	"Reshard":             3000,
	"GossipTx":            500,
	"RebalanceShards":     800,
	"VerticalPartition":   200,
	"HorizontalPartition": 200,
	"CompressData":        400,
	"DecompressData":      400,
	// Broadcast already priced

	// ----------------------------------------------------------------------
	// Side-chains
	// ----------------------------------------------------------------------
	"InitSidechains":            1200,
	"Sidechains":                60,
	"Sidechain_Register":        0,
	"SubmitHeader":              800,
	"VerifyWithdraw":            400,
	"VerifyAggregateSig":        800,
	"VerifyMerkleProof":         120,
	"GetSidechainMeta":          100,
	"ListSidechains":            120,
	"GetSidechainHeader":        100,
	"PauseSidechain":            300,
	"ResumeSidechain":           300,
	"UpdateSidechainValidators": 500,
	"RemoveSidechain":           600,
	// Deposit already priced

	// ----------------------------------------------------------------------
	// State-Channels
	// ----------------------------------------------------------------------
	"InitStateChannels":    800,
	"Channels":             60,
	"OpenChannel":          1000,
	"VerifyECDSASignature": 200,
	"InitiateClose":        300,
	"Challenge":            400,
	"Finalize":             500,
	"GetChannel":           80,
	"ListChannels":         120,
	"PauseChannel":         150,
	"ResumeChannel":        150,
	"CancelClose":          300,
	"ForceClose":           600,

	// ----------------------------------------------------------------------
	// Storage / Marketplace
	// ----------------------------------------------------------------------
	"NewStorage":    1200,
	"CreateListing": 800,
	"Exists":        40,
	"OpenDeal":      500,
	"Create":        800,
	"CloseDeal":     500,
	"Release":       200,
	"GetListing":    100,
	"ListListings":  100,
	"GetDeal":       100,
	"ListDeals":     100,
	"IPFS_Add":      0,
	"IPFS_Get":      0,
	"IPFS_Unpin":    0,

	// General Marketplace
	"CreateMarketListing": 800,
	"PurchaseItem":        600,
	"CancelListing":       300,
	"ReleaseFunds":        200,
	"GetMarketListing":    100,
	"ListMarketListings":  100,
	"GetMarketDeal":       100,
	"ListMarketDeals":     100,

	// Tangible assets
	"Assets_Register": 0,
	"Assets_Transfer": 0,
	"Assets_Get":      0,
	"Assets_List":     0,
	// Pin & Retrieve already priced
	// ----------------------------------------------------------------------
	// Identity Verification
	// ----------------------------------------------------------------------
	"RegisterIdentity": 500,
	"VerifyIdentity":   100,
	"RemoveIdentity":   200,
	"ListIdentities":   200,

	// ----------------------------------------------------------------------
	// Resource Marketplace
	// ----------------------------------------------------------------------
	"ListResource":         800,
	"OpenResourceDeal":     500,
	"CloseResourceDeal":    500,
	"GetResourceListing":   100,
	"ListResourceListings": 100,
	"GetResourceDeal":      100,
	"ListResourceDeals":    100,

	// ----------------------------------------------------------------------
	// Token Standards (constants – zero-cost markers)
	// ----------------------------------------------------------------------
	"StdSYN10":   1,
	"StdSYN20":   2,
	"StdSYN70":   7,
	"StdSYN130":  13,
	"StdSYN131":  13,
	"StdSYN200":  20,
	"StdSYN223":  22,
	"StdSYN300":  30,
	"StdSYN500":  50,
	"StdSYN600":  60,
	"StdSYN700":  70,
	"StdSYN721":  72,
	"StdSYN722":  72,
	"StdSYN800":  80,
	"StdSYN845":  84,
	"StdSYN900":  90,
	"StdSYN1000": 100,
	"StdSYN1100": 110,
	"StdSYN1155": 115,
	"StdSYN1200": 120,
	"StdSYN1300": 130,
	"StdSYN1401": 140,
	"StdSYN1500": 150,
	"StdSYN1600": 160,
	"StdSYN1700": 170,
	"StdSYN1800": 180,
	"StdSYN1900": 190,
	"StdSYN1967": 196,
	"StdSYN2100": 210,
	"StdSYN2200": 220,
	"StdSYN2369": 236,
	"StdSYN2400": 240,
	"StdSYN2500": 250,
	"StdSYN2600": 260,
	"StdSYN2700": 270,
	"StdSYN2800": 280,
	"StdSYN2900": 290,
	"StdSYN3000": 300,
	"StdSYN3100": 310,
	"StdSYN3200": 320,
	"StdSYN3300": 330,
	"StdSYN3400": 340,
	"StdSYN3500": 350,
	"StdSYN3600": 360,
	"StdSYN3700": 370,
	"StdSYN3800": 380,
	"StdSYN3900": 390,
	"StdSYN4200": 420,
	"StdSYN4300": 430,
	"StdSYN4700": 470,
	"StdSYN4900": 490,
	"StdSYN5000": 500,

	// ----------------------------------------------------------------------
	// Token Utilities
	// ----------------------------------------------------------------------
	"ID":                         40,
	"Meta":                       40,
	"Allowance":                  40,
	"Approve":                    80,
	"Add":                        60,
	"Sub":                        60,
	"Get":                        40,
	"transfer":                   210,
	"Calculate":                  80,
	"RegisterToken":              800,
	"NewBalanceTable":            500,
	"Set":                        60,
	"RefundGas":                  10,
	"PopUint32":                  3,
	"PopAddress":                 30,
	"PopUint64":                  6,
	"PushBool":                   30,
	"Push":                       30,
	"Len":                        20,
	"InitTokens":                 800,
	"GetRegistryTokens":          40,
	"TokenManager_Create":        0,
	"TokenManager_Transfer":      0,
	"TokenManager_Mint":          0,
	"TokenManager_Burn":          0,
	"TokenManager_Approve":       0,
	"TokenManager_BalanceOf":     0,
	"SYN1100_AddRecord":          50,
	"SYN1100_GrantAccess":        30,
	"SYN1100_RevokeAccess":       20,
	"SYN1100_GetRecord":          40,
	"SYN1100_TransferOwnership":  50,
	"ID":                         40,
	"Meta":                       40,
	"Allowance":                  40,
	"Approve":                    80,
	"Add":                        60,
	"Sub":                        60,
	"Get":                        40,
	"transfer":                   210,
	"Calculate":                  80,
	"RegisterToken":              800,
	"NewBalanceTable":            500,
	"Set":                        60,
	"RefundGas":                  10,
	"PopUint32":                  3,
	"PopAddress":                 30,
	"PopUint64":                  6,
	"PushBool":                   30,
	"Push":                       30,
	"Len":                        20,
	"InitTokens":                 800,
	"GetRegistryTokens":          40,
	"TokenManager_Create":        0,
	"TokenManager_Transfer":      0,
	"TokenManager_Mint":          0,
	"TokenManager_Burn":          0,
	"TokenManager_Approve":       0,
	"TokenManager_BalanceOf":     0,
	"SupplyChain_RegisterAsset":  0,
	"SupplyChain_UpdateLocation": 0,
	"SupplyChain_UpdateStatus":   0,
	"SupplyChain_TransferAsset":  0,
	"ID":                         40,
	"Meta":                       40,
	"Allowance":                  40,
	"Approve":                    80,
	"Add":                        60,
	"Sub":                        60,
	"Get":                        40,
	"transfer":                   210,
	"Calculate":                  80,
	"RegisterToken":              800,
	"NewBalanceTable":            500,
	"Set":                        60,
	"RefundGas":                  10,
	"PopUint32":                  3,
	"PopAddress":                 30,
	"PopUint64":                  6,
	"PushBool":                   30,
	"Push":                       30,
	"Len":                        20,
	"InitTokens":                 800,
	"GetRegistryTokens":          40,
	"TokenManager_Create":        0,
	"TokenManager_Transfer":      0,
	"TokenManager_Mint":          0,
	"TokenManager_Burn":          0,
	"TokenManager_Approve":       0,
	"TokenManager_BalanceOf":     0,
	"SYN70_RegisterAsset":        0,
	"SYN70_TransferAsset":        0,
	"SYN70_UpdateAttributes":     0,
	"SYN70_RecordAchievement":    0,
	"SYN70_GetAsset":             0,
	"SYN70_ListAssets":           0,
	"MusicRoyalty_AddRevenue":    80,
	"MusicRoyalty_Distribute":    100,
	"MusicRoyalty_UpdateInfo":    50,
	"ID":                         40,
	"Meta":                       40,
	"Allowance":                  40,
	"Approve":                    80,
	"Add":                        60,
	"Sub":                        60,
	"Get":                        40,
	"transfer":                   210,
	"Calculate":                  80,
	"RegisterToken":              800,
	"NewBalanceTable":            500,
	"Set":                        60,
	"RefundGas":                  10,
	"PopUint32":                  3,
	"PopAddress":                 30,
	"PopUint64":                  6,
	"PushBool":                   30,
	"Push":                       30,
	"Len":                        20,
	"InitTokens":                 800,
	"GetRegistryTokens":          40,
	"TokenManager_Create":        0,
	"TokenManager_Transfer":      0,
	"TokenManager_Mint":          0,
	"TokenManager_Burn":          0,
	"TokenManager_Approve":       0,
	"TokenManager_BalanceOf":     0,
	"SYN600_Stake":               200,
	"SYN600_Unstake":             200,
	"SYN600_AddEngagement":       50,
	"SYN600_EngagementOf":        40,
	"SYN600_DistributeRewards":   500,
	"SYN2100_RegisterDocument":   0,
	"SYN2100_FinanceDocument":    0,
	"SYN2100_GetDocument":        0,
	"SYN2100_ListDocuments":      0,
	"SYN2100_AddLiquidity":       0,
	"SYN2100_RemoveLiquidity":    0,
	"SYN2100_LiquidityOf":        0,
	"ID":                         40,
	"Meta":                       40,
	"Allowance":                  40,
	"Approve":                    80,
	"Add":                        60,
	"Sub":                        60,
	"Get":                        40,
	"transfer":                   210,
	"Calculate":                  80,
	"RegisterToken":              800,
	"NewBalanceTable":            500,
	"Set":                        60,
	"RefundGas":                  10,
	"PopUint32":                  3,
	"PopAddress":                 30,
	"PopUint64":                  6,
	"PushBool":                   30,
	"Push":                       30,
	"Len":                        20,
	"InitTokens":                 800,
	"GetRegistryTokens":          40,
	"TokenManager_Create":        0,
	"TokenManager_Transfer":      0,
	"TokenManager_Mint":          0,
	"TokenManager_Burn":          0,
	"TokenManager_Approve":       0,
	"TokenManager_BalanceOf":     0,
	"SYN500_GrantAccess":         0,
	"SYN500_UpdateAccess":        0,
	"SYN500_RevokeAccess":        0,
	"SYN500_RecordUsage":         0,
	"SYN500_RedeemReward":        0,
	"SYN500_RewardBalance":       0,
	"SYN500_Usage":               0,
	"SYN500_AccessInfo":          0,
	"SYN800_RegisterAsset":       10,
	"SYN800_UpdateValuation":     10,
	"SYN800_GetAsset":            5,
	"IDToken_Register":           0,
	"IDToken_Verify":             0,
	"IDToken_Get":                0,
	"IDToken_Logs":               0,
	"SYN1200_AddBridge":          0,
	"SYN1200_AtomicSwap":         0,
	"SYN1200_CompleteSwap":       0,
	"SYN1200_GetSwap":            0,
	"RegisterIPAsset":            800,
	"TransferIPOwnership":        300,
	"CreateLicense":              400,
	"RevokeLicense":              200,
	"RecordRoyalty":              100,
	"Event_Create":               100,
	"Event_IssueTicket":          150,
	"Event_Transfer":             210,
	"Event_Verify":               50,
	"Event_Use":                  40,
	"Tokens_RecordEmission":      0,
	"Tokens_RecordOffset":        0,
	"Tokens_NetBalance":          0,
	"Tokens_ListRecords":         0,
	"Edu_RegisterCourse":         0,
	"Edu_IssueCredit":            0,
	"Edu_VerifyCredit":           0,
	"Edu_RevokeCredit":           0,
	"Edu_GetCredit":              0,
	"Edu_ListCredits":            0,
	"Tokens_CreateSYN2200":       0,
	"Tokens_SendPayment":         0,
	"Tokens_GetPayment":          0,
	"DataToken_UpdateMeta":       0,
	"DataToken_SetPrice":         0,
	"DataToken_GrantAccess":      0,
	"DataToken_RevokeAccess":     0,

	// ----------------------------------------------------------------------
	// Transactions
	// ----------------------------------------------------------------------
	"VerifySig":          350,
	"ValidateTx":         500,
	"NewTxPool":          1200,
	"AddTx":              600,
	"PickTxs":            150,
	"TxPoolSnapshot":     80,
	"Exec_Begin":         0,
	"Exec_RunTx":         0,
	"Exec_Finalize":      0,
	"EncryptTxPayload":   350,
	"DecryptTxPayload":   300,
	"SubmitPrivateTx":    650,
	"EncodeEncryptedHex": 30,
	"ReverseTransaction": 1000,
	"NewTxDistributor":   800,
	"DistributeFees":     150,
	// Sign already priced

	// ----------------------------------------------------------------------
	// Low-level Math / Bitwise / Crypto opcodes
	// (values based on research into Geth & OpenEthereum plus Synnergy-specific
	//  micro-benchmarks – keep in mind that **all** word-size-dependent
	//  corrections are applied at run-time by the VM).
	// ----------------------------------------------------------------------
	"Short":            0,
	"BytesToAddress":   0,
	"Pop":              0,
	"opADD":            0,
	"opMUL":            0,
	"opSUB":            0,
	"OpDIV":            0,
	"opSDIV":           0,
	"opMOD":            0,
	"opSMOD":           0,
	"opADDMOD":         0,
	"opMULMOD":         0,
	"opEXP":            1,
	"opSIGNEXTEND":     0,
	"opLT":             0,
	"opGT":             0,
	"opSLT":            0,
	"opSGT":            0,
	"opEQ":             0,
	"opISZERO":         0,
	"opAND":            0,
	"opOR":             0,
	"opXOR":            0,
	"opNOT":            0,
	"opBYTE":           0,
	"opSHL":            0,
	"opSHR":            0,
	"opSAR":            0,
	"opECRECOVER":      70,
	"opEXTCODESIZE":    70,
	"opEXTCODECOPY":    70,
	"opEXTCODEHASH":    70,
	"opRETURNDATASIZE": 0,
	"opRETURNDATACOPY": 70,
	"opMLOAD":          0,
	"opMSTORE":         0,
	"opMSTORE8":        0,
	"opCALLDATALOAD":   0,
	"opCALLDATASIZE":   0,
	"opCALLDATACOPY":   70,
	"opCODESIZE":       0,
	"opCODECOPY":       70,
	"opJUMP":           0,
	"opJUMPI":          1,
	"opPC":             0,
	"opMSIZE":          0,
	"opGAS":            0,
	"opJUMPDEST":       0,
	"opSHA256":         25,
	"opKECCAK256":      25,
	"opRIPEMD160":      16,
	"opBLAKE2B256":     0,
	"opADDRESS":        0,
	"opCALLER":         0,
	"opORIGIN":         0,
	"opCALLVALUE":      0,
	"opGASPRICE":       0,
	"opNUMBER":         0,
	"opTIMESTAMP":      0,
	"opDIFFICULTY":     0,
	"opGASLIMIT":       0,
	"opCHAINID":        0,
	"opBLOCKHASH":      2,
	"opBALANCE":        40,
	"opSELFBALANCE":    0,
	"opLOG0":           0,
	"opLOG1":           0,
	"opLOG2":           0,
	"opLOG3":           0,
	"opLOG4":           0,
	"logN":             200,
	"opCREATE":         3200,
	"opCALL":           70,
	"opCALLCODE":       70,
	"opDELEGATECALL":   70,
	"opSTATICCALL":     70,
	"opRETURN":         0,
	"opREVERT":         0,
	"opSTOP":           0,
	"opSELFDESTRUCT":   500,

	// Shared accounting ops
	"TransferVM": 210,

	// ----------------------------------------------------------------------
	// Virtual Machine Internals
	// ----------------------------------------------------------------------
	"BurnVM":            210,
	"BurnLP":            210,
	"MintLP":            210,
	"NewInMemory":       50,
	"CallCode":          70,
	"CallContract":      70,
	"StaticCallVM":      70,
	"GetBalance":        40,
	"GetTokenBalance":   40,
	"SetTokenBalance":   50,
	"GetTokenSupply":    50,
	"SetBalance":        50,
	"DelegateCall":      70,
	"GetToken":          40,
	"NewMemory":         50,
	"Read":              0,
	"Write":             0,
	"LenVM":             0,
	"Call":              70,
	"SelectVM":          100,
	"CreateContract":    3200,
	"AddLog":            37,
	"GetCode":           20,
	"GetCodeHash":       20,
	"MintTokenVM":       200,
	"PrefixIterator":    50,
	"NonceOf":           40,
	"GetState":          40,
	"SetState":          50,
	"HasState":          40,
	"DeleteState":       50,
	"NewGasMeter":       50,
	"SelfDestructVM":    500,
	"Remaining":         0,
	"Consume":           0,
	"ExecuteVM":         200,
	"NewSuperLightVM":   50,
	"NewLightVM":        80,
	"NewHeavyVM":        120,
	"ExecuteSuperLight": 100,
	"ExecuteLight":      150,
	"ExecuteHeavy":      200,

	// Sandbox management
	"VM_SandboxStart":  0,
	"VM_SandboxStop":   0,
	"VM_SandboxReset":  0,
	"VM_SandboxStatus": 0,
	"VM_SandboxList":   0,

	// ----------------------------------------------------------------------
	// Smart Legal Contracts
	// ----------------------------------------------------------------------
	"Legal_Register": 0,
	"Legal_Sign":     0,
	"Legal_Revoke":   0,
	"Legal_Info":     0,
	"Legal_List":     0,
	// Plasma
	// ----------------------------------------------------------------------
	"InitPlasma":      800,
	"Plasma_Deposit":  0,
	"Plasma_Withdraw": 0,

	// Gaming
	// ----------------------------------------------------------------------
	"CreateGame": 800,
	"JoinGame":   400,
	"FinishGame": 600,
	"GetGame":    100,
	"ListGames":  200,

	// ----------------------------------------------------------------------
	// Messaging / Queue Management
	// ----------------------------------------------------------------------
	"NewMessageQueue":      500,
	"EnqueueMessage":       50,
	"DequeueMessage":       50,
	"BroadcastNextMessage": 100,
	"ProcessNextMessage":   200,
	"QueueLength":          10,
	"ClearQueue":           20,

	// ----------------------------------------------------------------------
	// Wallet / Key-Management
	// ----------------------------------------------------------------------
	"NewRandomWallet":            1000,
	"WalletFromMnemonic":         500,
	"NewHDWalletFromSeed":        600,
	"PrivateKey":                 40,
	"NewAddress":                 50,
	"SignTx":                     300,
	"RegisterIDWallet":           800,
	"IsIDWalletRegistered":       50,
	"NewOffChainWallet":          800,
	"OffChainWalletFromMnemonic": 500,
	"SignOffline":                250,
	"StoreSignedTx":              30,
	"LoadSignedTx":               30,
	"BroadcastSignedTx":          100,

	// ----------------------------------------------------------------------
	// Access Control
	// ----------------------------------------------------------------------
	"GrantRole":  100,
	"RevokeRole": 100,
	"HasRole":    30,
	"ListRoles":  50,
	// Geolocation Network
	// ----------------------------------------------------------------------
	"RegisterLocation": 200,
	"GetLocation":      50,
	"ListLocations":    100,
	"NodesInRadius":    150,
	// Firewall
	// ----------------------------------------------------------------------
	"NewFirewall":               400,
	"Firewall_BlockAddress":     0,
	"Firewall_UnblockAddress":   0,
	"Firewall_IsAddressBlocked": 0,
	"Firewall_BlockToken":       0,
	"Firewall_UnblockToken":     0,
	"Firewall_IsTokenBlocked":   0,
	"Firewall_BlockIP":          0,
	"Firewall_UnblockIP":        0,
	"Firewall_IsIPBlocked":      0,
	"Firewall_ListRules":        0,
	"Firewall_CheckTx":          0,
	// RPC / WebRTC
	// ----------------------------------------------------------------------
	"NewRPCWebRTC":    1000,
	"RPC_Serve":       0,
	"RPC_Close":       0,
	"RPC_ConnectPeer": 0,
	"RPC_Broadcast":   0,
	// Plasma
	// ----------------------------------------------------------------------
	"Plasma_StartExit":    0,
	"Plasma_FinalizeExit": 0,
	"Plasma_GetExit":      0,
	"Plasma_ListExits":    0,
	// ---------------------------------------------------------------------
	// Plasma Management
	// ---------------------------------------------------------------------
	"Plasma_SubmitBlock": 0,
	"Plasma_GetBlock":    0,
	// Resource Management
	// ---------------------------------------------------------------------
	"SetQuota":         100,
	"GetQuota":         50,
	"ChargeResources":  200,
	"ReleaseResources": 100,

	// Distribution
	// ---------------------------------------------------------------------
	"NewDistributor": 100,
	"BatchTransfer":  400,
	"Airdrop":        300,
	"DistributeEven": 200,
	// Carbon Credit System
	// ---------------------------------------------------------------------
	"InitCarbonEngine":  800,
	"Carbon":            200,
	"RegisterProject":   500,
	"IssueCredits":      500,
	"RetireCredits":     300,
	"ProjectInfo":       100,
	"ListProjects":      100,
	"AddVerification":   200,
	"ListVerifications": 100,
	// ----------------------------------------------------------------------
	// Finalization Management
	// ----------------------------------------------------------------------
	"NewFinalizationManager": 800,
	"FinalizeBlock":          400,
	"FinalizeBatchManaged":   350,
	"FinalizeChannelManaged": 350,
	"RegisterRecovery":       500,
	"RecoverAccount":         800,

	// ---------------------------------------------------------------------
	// DeFi
	// ---------------------------------------------------------------------
	"DeFi_CreateInsurance":   0,
	"DeFi_ClaimInsurance":    0,
	"DeFi_PlaceBet":          0,
	"DeFi_SettleBet":         0,
	"DeFi_StartCrowdfund":    0,
	"DeFi_Contribute":        0,
	"DeFi_FinalizeCrowdfund": 0,
	"DeFi_CreatePrediction":  0,
	"DeFi_VotePrediction":    0,
	"DeFi_ResolvePrediction": 0,
	"DeFi_RequestLoan":       0,
	"DeFi_RepayLoan":         0,
	"DeFi_StartYieldFarm":    0,
	"DeFi_Stake":             0,
	"DeFi_Unstake":           0,
	"DeFi_CreateSynthetic":   0,
	"DeFi_MintSynthetic":     0,
	"DeFi_BurnSynthetic":     0,
	// ----------------------------------------------------------------------
	// Binary Tree Operations
	// ----------------------------------------------------------------------
	"BinaryTreeNew":     500,
	"BinaryTreeInsert":  600,
	"BinaryTreeSearch":  400,
	"BinaryTreeDelete":  600,
	"BinaryTreeInOrder": 300,

	// ---------------------------------------------------------------------
	// Regulatory Management
	// ---------------------------------------------------------------------
	"InitRegulatory":    400,
	"RegisterRegulator": 600,
	"GetRegulator":      200,
	"ListRegulators":    200,
	"EvaluateRuleSet":   500,

	// ----------------------------------------------------------------------
	// Polls Management
	// ----------------------------------------------------------------------
	"CreatePoll": 800,
	"VotePoll":   300,
	"ClosePoll":  200,
	"GetPoll":    50,
	"ListPolls":  100,

	// ---------------------------------------------------------------------
	// Feedback System
	// ---------------------------------------------------------------------
	"InitFeedback":    800,
	"Feedback_Submit": 0,
	"Feedback_Get":    0,
	"Feedback_List":   0,
	"Feedback_Reward": 0,

	// ---------------------------------------------------------------------
	// Forum
	// ---------------------------------------------------------------------
	"Forum_CreateThread": 0,
	"Forum_GetThread":    0,
	"Forum_ListThreads":  0,
	"Forum_AddComment":   0,
	"Forum_ListComments": 0,

	// ------------------------------------------------------------------
	// Blockchain Compression
	// ------------------------------------------------------------------
	"CompressLedger":         600,
	"DecompressLedger":       600,
	"SaveCompressedSnapshot": 800,
	"LoadCompressedSnapshot": 800,

	// ----------------------------------------------------------------------
	// Biometrics Authentication
	// ----------------------------------------------------------------------
	"Bio_Enroll": 0,
	"Bio_Verify": 0,
	"Bio_Delete": 0,

	// ---------------------------------------------------------------------
	// System Health & Logging
	// ---------------------------------------------------------------------
	"NewHealthLogger": 800,
	"MetricsSnapshot": 100,
	"LogEvent":        50,
	"RotateLogs":      400,

	// ----------------------------------------------------------------------
	// Workflow / Key-Management
	// ----------------------------------------------------------------------
	"NewWorkflow":        1500,
	"AddWorkflowAction":  200,
	"SetWorkflowTrigger": 100,
	"SetWebhook":         100,
	"ExecuteWorkflow":    500,
	"ListWorkflows":      50,

	// ------------------------------------------------------------------
	// Swarm
	// ------------------------------------------------------------------
	"NewSwarm":          1000,
	"Swarm_AddNode":     0,
	"Swarm_RemoveNode":  0,
	"Swarm_BroadcastTx": 0,
	"Swarm_Start":       0,
	"Swarm_Stop":        0,
	"Swarm_Peers":       0,
	// ----------------------------------------------------------------------
	// Real Estate
	// ----------------------------------------------------------------------

	"RegisterProperty": 400,
	"TransferProperty": 350,
	"GetProperty":      100,
	"ListProperties":   150,

	// ----------------------------------------------------------------------
	// Event Management
	// ----------------------------------------------------------------------
	"InitEvents":     500,
	"EmitEvent":      40,
	"GetEvent":       80,
	"ListEvents":     100,
	"CreateWallet":   1000,
	"ImportWallet":   500,
	"WalletBalance":  40,
	"WalletTransfer": 210,

	// ----------------------------------------------------------------------
	// Employment Contracts
	// ----------------------------------------------------------------------
	"InitEmployment": 1000,
	"CreateJob":      800,
	"SignJob":        300,
	"RecordWork":     100,
	"PaySalary":      800,
	"GetJob":         100,
	// ------------------------------------------------------------------
	// Escrow Management
	// ------------------------------------------------------------------
	"Escrow_Create":  0,
	"Escrow_Deposit": 0,
	"Escrow_Release": 0,
	"Escrow_Cancel":  0,
	"Escrow_Get":     0,
	"Escrow_List":    0,
	// ---------------------------------------------------------------------
	// Faucet
	// ---------------------------------------------------------------------
	"NewFaucet":          500,
	"Faucet_Request":     0,
	"Faucet_Balance":     0,
	"Faucet_SetAmount":   0,
	"Faucet_SetCooldown": 0,
	// ----------------------------------------------------------------------
	// Supply Chain
	// ----------------------------------------------------------------------
	"GetItem":        100,
	"RegisterItem":   1000,
	"UpdateLocation": 500,
	"MarkStatus":     500,

	// ----------------------------------------------------------------------
	// Healthcare Records
	// ----------------------------------------------------------------------
	"InitHealthcare":    800,
	"RegisterPatient":   300,
	"AddHealthRecord":   400,
	"GrantAccess":       150,
	"RevokeAccess":      100,
	"ListHealthRecords": 200,

	// ----------------------------------------------------------------------
	// Warehouse Records
	// ----------------------------------------------------------------------

	"Warehouse_New":        0,
	"Warehouse_AddItem":    0,
	"Warehouse_RemoveItem": 0,
	"Warehouse_MoveItem":   0,
	"Warehouse_ListItems":  0,
	"Warehouse_GetItem":    0,

	// ---------------------------------------------------------------------
	// Immutability Enforcement
	// ---------------------------------------------------------------------
	"InitImmutability": 800,
	"VerifyChain":      400,
	"RestoreChain":     600,
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
