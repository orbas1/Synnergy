package core

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"sort"
	"time"
)

// Opcode identifiers for CDN module
const (
	OpRegisterNode  uint16 = 0xC100
	OpUploadAsset   uint16 = 0xC101
	OpRetrieveAsset uint16 = 0xC102
)

// CDNNode represents a CDN provider node
type CDNNode struct {
	ID         Address   `json:"id"`
	Addr       string    `json:"addr"`
	CapacityMB int       `json:"capacity_mb"`
	Registered time.Time `json:"registered"`
}

// Asset metadata stored on-chain
type Asset struct {
	CID        string    `json:"cid"`
	Size       uint64    `json:"size"`
	UploadedAt time.Time `json:"uploaded_at"`
}

// chooseNodes selects top-capacity nodes for replication
func chooseNodes(asset Asset, replication int) ([]CDNNode, error) {
	store := CurrentStore()
	prefix := []byte("cdn:node:")
	it := store.Iterator(prefix, nil)
	defer it.Close()

	var nodes []CDNNode
	for it.Next() {
		var node CDNNode
		if err := json.Unmarshal(it.Value(), &node); err != nil {
			zap.L().Error("failed to unmarshal CDN node", zap.Error(err))
			continue
		}
		nodes = append(nodes, node)
	}

	if len(nodes) == 0 {
		return nil, fmt.Errorf("no CDN nodes registered")
	}

	sort.Slice(nodes, func(i, j int) bool {
		return nodes[i].CapacityMB > nodes[j].CapacityMB
	})

	if replication > len(nodes) {
		replication = len(nodes)
	}
	return nodes[:replication], nil
}

// RegisterNode allows a node to join the CDN network
func RegisterNode(n CDNNode) error {
	logger := zap.L().Sugar()
	logger.Infof("Registering CDN node %s at %s", n.ID, n.Addr)

	n.Registered = time.Now().UTC()
	key := fmt.Sprintf("cdn:node:%s", n.ID)

	store := CurrentStore()
	raw, err := json.Marshal(n)
	if err != nil {
		logger.Errorf("marshal node failed: %v", err)
		return err
	}
	if err := store.Set([]byte(key), raw); err != nil {
		logger.Errorf("ledger write failed: %v", err)
		return err
	}

	// Broadcast new node
	Broadcast(TopicCDNNodeRegistry, raw)
	logger.Infof("CDN node %s registered", n.ID)
	return nil
}

// UploadAsset pins data to storage, replicates to selected nodes, and records metadata
func UploadAsset(data []byte) (string, error) {
	logger := zap.L().Sugar()
	// Pin to underlying storage
	cid, err := Pin(data)
	if err != nil {
		logger.Errorf("storage pin failed: %v", err)
		return "", err
	}
	size := uint64(len(data))
	asset := Asset{CID: cid, Size: size, UploadedAt: time.Now().UTC()}

	// Choose replication targets
	replication := CDNReplicationFactor
	nodes, err := chooseNodes(asset, replication)
	if err != nil {
		logger.Errorf("node selection failed: %v", err)
		return cid, err
	}

	// Send replication instructions
	payload, _ := json.Marshal(struct {
		CID   string   `json:"cid"`
		Nodes []string `json:"nodes"`
	}{CID: cid, Nodes: func() []string {
		addrs := make([]string, len(nodes))
		for i, node := range nodes {
			addrs[i] = node.Addr
		}
		return addrs
	}()})
	Broadcast(TopicCDNReplication, payload)

	// Persist asset metadata on-chain
	key := fmt.Sprintf("cdn:asset:%s", cid)
	raw, err := json.Marshal(asset)
	if err != nil {
		logger.Errorf("marshal asset failed: %v", err)
		return cid, err
	}
	if err := CurrentStore().Set([]byte(key), raw); err != nil {
		logger.Errorf("ledger write failed: %v", err)
		return cid, err
	}

	logger.Infof("Asset %s uploaded and replication triggered", cid)
	return cid, nil
}

const CDNReplicationFactor = 3
const TopicCDNReplication = "cdn:replication"
const TopicCDNNodeRegistry = "cdn:node:registry"

func Pin(data []byte) (string, error) {
	// In real logic, this could hash the data and/or send it to IPFS
	hash := sha256.Sum256(data)
	return fmt.Sprintf("%x", hash[:]), nil // return CID
}

func Retrieve(cid string) ([]byte, error) {
	// Simulate lookup in local CDN storage by CID
	key := []byte(fmt.Sprintf("cdn:asset:data:%s", cid))
	return CurrentStore().Get(key)
}

// RetrieveAsset fetches data from local or remote storage
func RetrieveAsset(cid string) ([]byte, error) {
	logger := zap.L().Sugar()
	data, err := Retrieve(cid)
	if err == nil {
		return data, nil
	}
	logger.Warnf("local retrieve failed for %s: %v, attempting network fetch", cid, err)

	// Fallback: broadcast request and wait for response (simplified)
	req := []byte(cid)
	Broadcast(TopicCDNFetchRequest, req)
	// In production, implement response listener or direct P2P fetch
	// For now, return error
	return nil, fmt.Errorf("asset %s not found locally", cid)
}

const TopicCDNFetchRequest = "cdn:fetch:request"

// Oracle represents an on-chain data feed
// used by crosschain bridges, logistics, finance, governance

type Oracle struct {
	ID        string    `json:"id"`
	Source    string    `json:"source"` // e.g., "price:BTC-USD" or "weather:NYC"
	LastValue []byte    `json:"last_value"`
	Timestamp time.Time `json:"timestamp"`
}

// RegisterOracle registers a new data feed oracle
func RegisterOracle(o Oracle) error {
	logger := zap.L().Sugar()
	if o.ID == "" {
		o.ID = uuid.New().String()
	}
	key := fmt.Sprintf("oracle:config:%s", o.ID)
	o.Timestamp = time.Now().UTC()

	raw, err := json.Marshal(o)
	if err != nil {
		logger.Errorf("marshal oracle failed: %v", err)
		return err
	}
	if err := CurrentStore().Set([]byte(key), raw); err != nil {
		logger.Errorf("persist oracle failed: %v", err)
		return err
	}
	// broadcast new oracle config
	Broadcast(TopicOracleRegistry, raw)
	logger.Infof("Oracle registered: %s (source=%s)", o.ID, o.Source)
	return nil
}

// PushFeed submits a new data point for an oracle
func PushFeed(oracleID string, value []byte) error {
	logger := zap.L().Sugar()
	cfgKey := fmt.Sprintf("oracle:config:%s", oracleID)
	rawCfg, err := CurrentStore().Get([]byte(cfgKey))
	if err != nil {
		logger.Errorf("oracle config not found: %s", oracleID)
		return ErrNotFound
	}
	var o Oracle
	if err := json.Unmarshal(rawCfg, &o); err != nil {
		logger.Errorf("unmarshal oracle config failed: %v", err)
		return err
	}
	// update value
	o.LastValue = value
	o.Timestamp = time.Now().UTC()
	raw, err := json.Marshal(o)
	if err != nil {
		logger.Errorf("marshal oracle update failed: %v", err)
		return err
	}
	dataKey := fmt.Sprintf("oracle:data:%s", oracleID)
	if err := CurrentStore().Set([]byte(dataKey), raw); err != nil {
		logger.Errorf("persist oracle feed failed: %v", err)
		return err
	}
	// broadcast new feed
	Broadcast(TopicOracleFeed, raw)
	logger.Infof("Oracle feed pushed: %s at %s", oracleID, o.Timestamp)
	return nil
}

const (
	TopicOracleRegistry = "oracle:registry"
	TopicOracleFeed     = "oracle:feed"
)

// QueryOracle retrieves the latest value for an oracle
func QueryOracle(oracleID string) ([]byte, error) {
	logger := zap.L().Sugar()
	dataKey := fmt.Sprintf("oracle:data:%s", oracleID)
	raw, err := CurrentStore().Get([]byte(dataKey))
	if err != nil {
		logger.Errorf("oracle data not found: %s", oracleID)
		return nil, ErrNotFound
	}
	var o Oracle
	if err := json.Unmarshal(raw, &o); err != nil {
		logger.Errorf("unmarshal oracle data failed: %v", err)
		return nil, err
	}
	// return raw.LastValue
	return o.LastValue, nil
}
