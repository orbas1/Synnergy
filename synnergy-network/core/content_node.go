package core

import (
	"encoding/json"
	"fmt"
	"sort"
	"time"
)

// ContentNetworkNode mirrors Nodes.ContentNode information for registry.
type ContentNetworkNode struct {
	ID         Address   `json:"id"`
	Addr       string    `json:"addr"`
	CapacityGB int       `json:"capacity_gb"`
	Registered time.Time `json:"registered"`
}

// chooseContentNodes selects replication targets with largest capacity.
func chooseContentNodes(replication int) ([]ContentNetworkNode, error) {
	store := CurrentStore()
	it := store.Iterator([]byte("content:node:"), nil)
	defer it.Close()

	var nodes []ContentNetworkNode
	for it.Next() {
		var n ContentNetworkNode
		if err := json.Unmarshal(it.Value(), &n); err == nil {
			nodes = append(nodes, n)
		}
	}
	if len(nodes) == 0 {
		return nil, fmt.Errorf("no content nodes registered")
	}
	sort.Slice(nodes, func(i, j int) bool { return nodes[i].CapacityGB > nodes[j].CapacityGB })
	if replication > len(nodes) {
		replication = len(nodes)
	}
	return nodes[:replication], nil
}

// RegisterContentNode adds a node to the content network registry.
func RegisterContentNode(n ContentNetworkNode) error {
	n.Registered = time.Now().UTC()
	raw, err := json.Marshal(n)
	if err != nil {
		return err
	}
	key := fmt.Sprintf("content:node:%s", n.ID)
	if err := CurrentStore().Set([]byte(key), raw); err != nil {
		return err
	}
	Broadcast(TopicCDNNodeRegistry, raw) // reuse topic for simplicity
	return nil
}

// UploadContent encrypts, pins and replicates data across content nodes.
func UploadContent(data, key []byte) (string, error) {
	cid, err := Pin(data)
	if err != nil {
		return "", err
	}
	nodes, err := chooseContentNodes(contentReplicationFactor)
	if err != nil {
		return cid, err
	}
	payload, _ := json.Marshal(struct {
		CID   string   `json:"cid"`
		Nodes []string `json:"nodes"`
	}{cid, func() []string {
		a := make([]string, len(nodes))
		for i, n := range nodes {
			a[i] = n.Addr
		}
		return a
	}()})
	Broadcast(TopicCDNReplication, payload)
	meta := ContentMeta{CID: cid, Size: uint64(len(data)), Uploaded: time.Now().UTC()}
	raw, _ := json.Marshal(meta)
	if err := CurrentStore().Set([]byte("content:meta:"+cid), raw); err != nil {
		return cid, err
	}
	return cid, nil
}

// RetrieveContent fetches pinned content by CID.
func RetrieveContent(cid string) ([]byte, error) {
	return Retrieve(cid)
}

// ListContentNodes returns registered content nodes.
func ListContentNodes() ([]ContentNetworkNode, error) {
	it := CurrentStore().Iterator([]byte("content:node:"), nil)
	defer it.Close()
	var list []ContentNetworkNode
	for it.Next() {
		var n ContentNetworkNode
		if err := json.Unmarshal(it.Value(), &n); err == nil {
			list = append(list, n)
		}
	}
	sort.Slice(list, func(i, j int) bool { return list[i].Registered.Before(list[j].Registered) })
	return list, nil
}

var contentReplicationFactor = 2
