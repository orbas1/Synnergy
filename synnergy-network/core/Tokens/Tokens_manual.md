# Tokens Manual

This document outlines the token standards used throughout the Synnergy
network.  All token implementations must conform to these conventions to
ensure interoperability across nodes, wallets and smart contracts.

## Core Concepts

### Address

All tokens operate on 20‑byte account identifiers represented by the
`Address` type.  The type is defined once in `index.go` and reused across
the package to avoid duplication and keep the package free from direct
dependencies on the parent `core` module.

### TokenInterfaces

Every token exposes the `TokenInterfaces` interface.  At minimum a token
must provide a `Meta` method returning implementation specific metadata.
Specialised tokens embed this interface to signal compatibility with the
Synnergy tooling.

### BaseToken and Registry

Common balance and allowance logic is provided by the `BaseToken`
implementation in `base.go`.  Constructors derive a deterministic token
ID with `deriveID`, populate initial balances via `NewBalanceTable` and
register themselves using `RegisterToken` so that other packages can
discover them by `TokenID`.


#### Allowances

`BaseToken` mirrors ERC‑20 style spending approvals.  The `Approve`
method sets allowances for third‑party spenders while `TransferFrom`
consumes those allowances and moves funds on behalf of the owner.
Allowances are decremented after each successful transfer.

### Standards

The `TokenStandard` enumeration lists all SYN standards.  Each concrete
token specifies its standard via the `Standard` field in the metadata
structure.  Implementations must embed this identifier into their token
ID so that clients can route transactions to the correct handlers.

## Example Tokens

* **SYN10** – Basic currency token providing exchange rate information.
* **SYN200** – Carbon credit tokens with verification records for
  environmental projects.
* **SYN845** – Debt instruments tracking principal, interest and payment
  history.
* **SYN1100** – Healthcare record tokens offering fine‑grained access
  control to encrypted medical data.

Each specialised token should keep its implementation self‑contained and
only rely on the shared types and interfaces defined within this
package.  This consolidation simplifies maintenance and ensures that new
tokens automatically integrate with existing network tooling.

## CLI Integra<<<<<<< codex/focus-on-stage-8-in-agents.mdtion

The `tokens` command in the CLI provides a uniform interface to inspect
and administer any registered token.  It supports listing tokens,
querying balances, transferring funds, minting, burning and managing
allowances.  Addresses are supplied as hex strings and decoded into the
common 20‑byte `Address` format.


