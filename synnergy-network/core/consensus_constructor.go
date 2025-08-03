package core

import (
	"math/big"

	"github.com/sirupsen/logrus"
)

// NewConsensus constructs a minimal SynnergyConsensus engine wiring the
// provided components. It initialises block heights from the ledger and sets a
// zero difficulty value so that other components can safely build upon it.
func NewConsensus(
	lg *logrus.Logger,
	led *Ledger,
	p2p interface{},
	crypt interface{},
	pool interface{},
	auth interface{},
) (*SynnergyConsensus, error) {
	if lg == nil {
		lg = logrus.StandardLogger()
	}

	return &SynnergyConsensus{
		logger:        lg,
		ledger:        led,
		p2p:           p2p,
		crypto:        crypt,
		pool:          pool,
		auth:          auth,
		nextSubHeight: led.LastSubBlockHeight() + 1,
		nextBlkHeight: led.LastBlockHeight() + 1,
		curDifficulty: big.NewInt(0),
		blkTimes:      make([]int64, 0, 100),
	}, nil
}
