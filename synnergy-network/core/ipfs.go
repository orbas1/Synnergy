package core

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"sync"

	logrus "github.com/sirupsen/logrus"
)

// IPFSService provides high level helpers for interacting with an IPFS gateway.
// It wraps the Storage module so ledger charging and local caching work
// transparently.
type IPFSService struct {
	storage *Storage
	client  *http.Client
	gateway string
	logger  *logrus.Logger
}

var (
	ipfsOnce sync.Once
	ipfsSvc  *IPFSService
)

// InitIPFS initialises the global IPFS service using the provided configuration.
// The ledger from helpers.CurrentLedger() is used for rent charging.
func InitIPFS(cfg *StorageConfig, lg *logrus.Logger) error {
	var err error
	ipfsOnce.Do(func() {
		if lg == nil {
			lg = logrus.New()
		}
		var storage *Storage
		storage, err = NewStorage(cfg, lg, nil)
		if err != nil {
			return
		}
		ipfsSvc = &IPFSService{
			storage: storage,
			client:  &http.Client{Timeout: cfg.GatewayTimeout},
			gateway: cfg.IPFSGateway,
			logger:  lg,
		}
	})
	return err
}

// IPFS returns the global service if initialised.
func IPFS() *IPFSService { return ipfsSvc }

// AddFile pins the given data via the configured gateway and broadcasts the CID.
func AddFile(ctx context.Context, data []byte, payer Address) (string, error) {
	if ipfsSvc == nil {
		return "", errors.New("ipfs service not initialised")
	}
	cid, _, err := ipfsSvc.storage.Pin(ctx, data, payer)
	if err != nil {
		return "", err
	}
	_ = Broadcast("ipfs:add", []byte(cid))
	ipfsSvc.logger.Infof("ipfs added %s", cid)
	return cid, nil
}

// GetFile retrieves data for the given CID using the storage cache and gateway.
func GetFile(ctx context.Context, cid string) ([]byte, error) {
	if ipfsSvc == nil {
		return nil, errors.New("ipfs service not initialised")
	}
	return ipfsSvc.storage.Retrieve(ctx, cid)
}

// UnpinFile removes a CID from the gateway pinset and broadcasts the event.
func UnpinFile(ctx context.Context, cid string) error {
	if ipfsSvc == nil {
		return errors.New("ipfs service not initialised")
	}
	url := fmt.Sprintf("%s/api/v0/pin/rm?arg=%s", ipfsSvc.gateway, cid)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, nil)
	if err != nil {
		return err
	}
	resp, err := ipfsSvc.client.Do(req)
	if err != nil {
		if resp != nil {
			resp.Body.Close()
		}
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(io.LimitReader(resp.Body, 128))
		return fmt.Errorf("gateway unpin %d: %s", resp.StatusCode, string(b))
	}
	_ = Broadcast("ipfs:unpin", []byte(cid))
	ipfsSvc.logger.Infof("ipfs unpinned %s", cid)
	return nil
}
