package core

import (
	"errors"
	"fmt"
	"sync"

	"github.com/ethereum/go-ethereum/crypto"
)

var (
	// PoSValidatorsAccount receives the PoS validator share of fees.
	PoSValidatorsAccount Address
	// PoHValidatorsAccount receives the PoH validator share of fees.
	PoHValidatorsAccount Address
	// Syn900RewardsAccount accrues rewards for authorised syn900 holders.
	Syn900RewardsAccount Address
	// AuthorityNodesAccount distributes rewards to authority nodes.
	AuthorityNodesAccount Address
)

func init() {
	var err error
	PoSValidatorsAccount, err = StringToAddress("0x506f5356616c696461746f727300000000000000")
	if err != nil {
		panic("invalid PoSValidatorsAccount: " + err.Error())
	}
	PoHValidatorsAccount, err = StringToAddress("0x506f4856616c696461746f727300000000000000")
	if err != nil {
		panic("invalid PoHValidatorsAccount: " + err.Error())
	}
	Syn900RewardsAccount, err = StringToAddress("0x53796e3930305265776172640000000000000000")
	if err != nil {
		panic("invalid Syn900RewardsAccount: " + err.Error())
	}
	AuthorityNodesAccount, err = StringToAddress("0x417574686f726974794e6f646573000000000000")
	if err != nil {
		panic("invalid AuthorityNodesAccount: " + err.Error())
	}
}

// TxDistributor splits transaction fees between network stakeholders.
// The fee is distributed as follows:
//
//   - 70%% → miners and validators
//   - 50%% of that 70%% → miners
//   - 39%% of that 70%% → PoS validators
//   - 20%% of that 70%% → PoH validators
//   - 5%%  → authorised syn900 token holders
//   - 10%% → loan pool
//   - 10%% → CharityPool
//   - 5%%  → authority nodes
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

// DistributeFees moves the transaction fee from the sender to all parties
// according to the distribution percentages described above.
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

	// Compute fixed percentage shares first to minimise rounding loss and
	// avoid overflow by dividing before multiplying.
	onePercent := fee / 100
	syn900Share := onePercent * 5
	loanPoolShare := onePercent * 10
	charityShare := onePercent * 10
	authorityShare := onePercent * 5

	// Remaining amount allocated to miners and validators (70%).
	remaining := fee - syn900Share - loanPoolShare - charityShare - authorityShare
	if remaining > fee {
		return fmt.Errorf("overflow in fee calculation")
	}

	// Split the 70% portion between miners, PoS and PoH validators using
	// weights 50:39:20 (sum 109) to honour the requested proportions.
	totalWeight := uint64(50 + 39 + 20)
	minerShare := remaining * 50 / totalWeight
	posShare := remaining * 39 / totalWeight
	pohShare := remaining - minerShare - posShare

	d.mu.Lock()
	defer d.mu.Unlock()

	if err := d.ledger.Transfer(from, minerAddr, minerShare); err != nil {
		return err
	}
	if err := d.ledger.Transfer(from, PoSValidatorsAccount, posShare); err != nil {
		return err
	}
	if err := d.ledger.Transfer(from, PoHValidatorsAccount, pohShare); err != nil {
		return err
	}
	if err := d.ledger.Transfer(from, Syn900RewardsAccount, syn900Share); err != nil {
		return err
	}
	if err := d.ledger.Transfer(from, LoanPoolAccount, loanPoolShare); err != nil {
		return err
	}
	if err := d.ledger.Transfer(from, CharityPoolAccount, charityShare); err != nil {
		return err
	}
	if err := d.ledger.Transfer(from, AuthorityNodesAccount, authorityShare); err != nil {
		return err
	}
	return nil
}
