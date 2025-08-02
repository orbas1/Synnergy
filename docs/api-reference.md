# Synnergy API Reference

This reference outlines key Go packages and their primary entry points within the Synnergy Network. The APIs below are stable for development purposes and may change as the project evolves.

## Package Overview

| Package | Description |
|---------|-------------|
| `core` | Ledger, consensus, networking and token logic |
| `cmd` | Command-line interfaces and helpers |
| `GUI` | Web and desktop clients |
| `tests` | Unit and integration tests |

## Node Lifecycle API

The `core` package exposes constructors and methods for running a node.

```go
import (
    "synnergy-network/core"
)

func runNode() error {
    cfg := core.DefaultConfig()
    n, err := core.NewNode(cfg)
    if err != nil {
        return err
    }
    return n.Start()
}
```

## Ledger Operations

Ledger utilities allow creation of databases and submission of transactions.

```go
import (
    "synnergy-network/core/ledger"
)

func submit(tx ledger.Transaction) error {
    l, err := ledger.Open("./ledger.db")
    if err != nil {
        return err
    }
    defer l.Close()
    return l.Add(tx)
}
```

## Token Management

Token helpers provide minting and transfer operations for SYN assets.

```go
import (
    "synnergy-network/core/tokens"
)

func issue(addr string, amount uint64) error {
    return tokens.Mint(addr, amount)
}
```

## Smart Contracts

Contract deployment requires compiled Wasm artifacts.

```go
import (
    "synnergy-network/core/contract"
)

func deploy(path string) error {
    c, err := contract.LoadWasm(path)
    if err != nil {
        return err
    }
    return contract.Deploy(c)
}
```

## Wallet Management

Wallet helpers for key generation and signing live in the `core` package.

```go
import (
    "fmt"
    "synnergy-network/core"
)

func createWallet() (*core.HDWallet, error) {
    w, mnemonic, err := core.NewRandomWallet(128)
    if err != nil {
        return nil, err
    }
    fmt.Println("mnemonic:", mnemonic)
    return w, nil
}
```

## Networking

Connection pooling utilities manage peer communication.

```go
import (
    "context"
    "synnergy-network/core"
    "time"
)

func dial(addr string) error {
    d := core.NewDialer(5*time.Second, 30*time.Second)
    pool := core.NewConnPool(d, 5, time.Minute)
    conn, err := pool.Acquire(context.Background(), addr)
    if err != nil {
        return err
    }
    defer pool.Release(conn)
    return nil
}
```

Each package contains additional types and methods; run `go doc <package>` for complete documentation.
