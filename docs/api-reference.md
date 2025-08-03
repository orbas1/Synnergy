# Synnergy API Reference

This reference outlines key Go packages and their primary entry points
within the Synnergy Network. The APIs below are stable for development
purposes and may change as the project evolves.

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
import "synnergy-network/core"

func runNode() (*core.Node, error) {
    cfg := core.Config{
        ListenAddr:   ":9000",
        DiscoveryTag: "synnergy",
    }
    n, err := core.NewNode(cfg)
    if err != nil {
        return nil, err
    }
    go n.ListenAndServe()
    return n, nil
}
```

## Ledger Operations

Ledger utilities allow creation of databases and submission of transactions.

```go
import "synnergy-network/core"

func submit(tx core.Transaction) error {
    l, err := core.OpenLedger("./ledger")
    if err != nil {
        return err
    }
    defer l.Close()
    l.AddToPool(&tx)
    return nil
}
```

## Token Management

Token helpers provide minting and transfer operations for SYN assets.

```go
import "synnergy-network/core"

func issue(id core.TokenID, addr core.Address, amount uint64) error {
    tm := core.NewTokenManager(nil, nil)
    return tm.Mint(id, addr, amount)
}
```

## Smart Contracts

Contract deployment requires compiled Wasm artifacts.

```go
import (
    "os"
    "synnergy-network/core"
)

func deploy(path string) error {
    code, err := os.ReadFile(path)
    if err != nil {
        return err
    }
    reg := core.GetContractRegistry()
    return reg.Deploy(core.Address{}, code, nil, 1_000_000)
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

Each package contains additional types and methods; run
`go doc <package>` for complete documentation.
