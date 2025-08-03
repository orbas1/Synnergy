package core

import (
	"errors"
	"fmt"
	"sync"

	"github.com/ethereum/go-ethereum/crypto"
)

// TxDistributor splits transaction fees between network stakeholders.
// 50%% of the fee goes to the miner, 30%% to the LoanPool treasury and
// the remaining 20%% funds the CharityPool.
//
// It uses the ledger to move funds directly from the transaction sender
// to the target accounts. The miner address is derived from the public key
// stored in the block header.
type TxDistributor struct {
	ledger *Ledger
	mu     sync.Mutex
}

// NewTxDistributor returns a fee distributor bound to a ledger instance.
func NewTxDistributor(ledger *Ledger) *TxDistributor {
	return &TxDistributor{ledger: ledger}
}

// AddressFromPubKey converts a compressed secp256k1 public key to an Address.
func AddressFromPubKey(pub []byte) (Address, error) {
	key, err := crypto.UnmarshalPubkey(pub)
	if err != nil {
		return AddressZero, err
	}
	return FromCommon(crypto.PubkeyToAddress(*key)), nil
}

// DistributeFees moves the transaction fee from the sender to the
// miner, LoanPoolAccount and CharityPoolAccount according to the
// distribution percentages described above.
func (d *TxDistributor) DistributeFees(from Address, minerPk []byte, fee uint64) error {
	if d.ledger == nil {
		return errors.New("distributor: nil ledger")
	}
	if fee == 0 {
		return nil
	}

	minerAddr, err := AddressFromPubKey(minerPk)
	if err != nil {
		return fmt.Errorf("decode miner: %w", err)
	}

	minerShare := fee / 2
	loanShare := fee * 30 / 100
	charityShare := fee - minerShare - loanShare

	d.mu.Lock()
	defer d.mu.Unlock()

	if err := d.ledger.Transfer(from, minerAddr, minerShare); err != nil {
		return err
	}
	if err := d.ledger.Transfer(from, LoanPoolAccount, loanShare); err != nil {
		return err
	}
	if err := d.ledger.Transfer(from, CharityPoolAccount, charityShare); err != nil {
		return err
	}
	return nil
}
