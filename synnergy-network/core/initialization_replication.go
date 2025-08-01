package core

import (
	"context"

	logrus "github.com/sirupsen/logrus"
)

// InitService orchestrates ledger bootstrap via the replication subsystem
// and optionally starts consensus once the ledger is up to date.
type InitService struct {
	rep  *Replicator
	led  *Ledger
	cons ConsensusStarter
	log  *logrus.Logger
}

// ConsensusStarter defines the minimal consensus interface used by InitService.
type ConsensusStarter interface{ Start(context.Context) }

// NewInitService wires a Replicator to an existing ledger and peer manager.
// If cons is nil only replication is started.
func NewInitService(cfg *ReplicationConfig, logger *logrus.Logger, led *Ledger, pm PeerManager, cons ConsensusStarter) *InitService {
	if logger == nil {
		logger = logrus.StandardLogger()
	}
	rep := NewReplicator(cfg, logger, led, pm)
	return &InitService{rep: rep, led: led, cons: cons, log: logger}
}

// BootstrapLedger syncs missing blocks from peers if the ledger is empty.
func (s *InitService) BootstrapLedger(ctx context.Context) error {
	if s.led.LastHeight() > 0 {
		return nil
	}
	s.log.Info("ledger empty – synchronizing from peers")
	if err := s.rep.Synchronize(ctx); err != nil {
		return err
	}
	return nil
}

// Start bootstraps the ledger then launches replication and consensus.
func (s *InitService) Start(ctx context.Context) error {
	if err := s.BootstrapLedger(ctx); err != nil {
		return err
	}
	s.rep.Start()
	if s.cons != nil {
		s.cons.Start(ctx)
	}
	s.log.Info("initialization service started")
	return nil
}

// Shutdown stops background replication.
func (s *InitService) Shutdown() {
	s.rep.Stop()
	s.log.Info("initialization service stopped")
}
