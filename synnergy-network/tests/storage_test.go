package core

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	logrus "github.com/sirupsen/logrus"
)

// stubLedger implements MeteredState for testing storage rent charging
type stubLedger struct {
	calls []int64
	err   error
}

func (s *stubLedger) ChargeStorageRent(payer Address, amount int64) error {
	s.calls = append(s.calls, amount)
	return s.err
}

// dummy VM for syscall registration
type dummyVM struct {
	opcode  byte
	handler interface{}
}

func (d *dummyVM) RegisterSyscall(op byte, fn interface{}) {
	d.opcode = op
	d.handler = fn
}

// Test newDiskLRU put/get and eviction behavior
func TestDiskLRU_PutGetEvict(t *testing.T) {
	dir := t.TempDir()
	lru, err := newDiskLRU(dir, 2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// put two entries
	if err := lru.put("a", []byte("dataA")); err != nil {
		t.Fatalf("put a failed: %v", err)
	}
	if err := lru.put("b", []byte("dataB")); err != nil {
		t.Fatalf("put b failed: %v", err)
	}

	// get existing
	if d, ok := lru.get("a"); !ok || string(d) != "dataA" {
		t.Fatalf("get a returned %s, %v", d, ok)
	}

	// insert third => evict oldest (a)
	if err := lru.put("c", []byte("dataC")); err != nil {
		t.Fatalf("put c failed: %v", err)
	}

	if _, ok := lru.get("a"); ok {
		t.Fatalf("expected a evicted")
	}
	if d, ok := lru.get("b"); !ok || string(d) != "dataB" {
		t.Fatalf("get b after eviction failed: %v %v", d, ok)
	}
}

// Test NewStorage validation and endpoints
func TestNewStorage_InvalidConfig(t *testing.T) {
	_, err := NewStorage(nil, logrus.New(), nil)
	if err == nil {
		t.Fatalf("expected error for nil config")
	}
}

// Test Pin success, gateway error, cid mismatch, cache hit
func TestStorage_Pin_Retrieve(t *testing.T) {
	// sample data
	data := []byte("hello world")
	// start test server
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v0/add", func(w http.ResponseWriter, r *http.Request) {
		// respond with matching JSON
		var meta struct {
			Hash string
			Size string
		}
		cidStr, _ := func() (string, string) {
			encodedMH, _ := mh.Sum(data, mh.SHA2_256, -1)
			c := cid.NewCidV1(cid.Raw, encodedMH)
			return c.String(), fmt.Sprint(len(data))
		}()
		json.NewEncoder(w).Encode(meta)
	})
	mux.HandleFunc("/ipfs/", func(w http.ResponseWriter, r *http.Request) {
		w.Write(data)
	})
	tls := httptest.NewServer(mux)
	defer tls.Close()

	cfg := &StorageConfig{
		CacheDir:         t.TempDir(),
		CacheSizeEntries: 10,
		IPFSGateway:      tls.URL,
		GatewayTimeout:   time.Second,
	}
	stub := &stubLedger{}
	s, err := NewStorage(cfg, logrus.New(), stub)
	if err != nil {
		t.Fatalf("NewStorage failed: %v", err)
	}

	ctx := context.Background()

	// first Pin => cache miss + gateway call
	cidStr, size, err := s.Pin(ctx, data, addrWithByte(0x0))
	if err != nil {
		t.Fatalf("Pin failed: %v", err)
	}
	if size != int64(len(data)) {
		t.Errorf("unexpected size %d", size)
	}
	// ledger charged
	if len(stub.calls) != 1 || stub.calls[0] != size {
		t.Errorf("expected rent charged %d, got %v", size, stub.calls)
	}

	// second Pin => cache hit, no new ledger call
	stub.calls = nil
	cid2, _, err := s.Pin(ctx, data, addrWithByte(0x0))
	if err != nil {
		t.Fatalf("Pin cache failed: %v", err)
	}
	if cid2 != cidStr {
		t.Errorf("unexpected cid %s != %s", cid2, cidStr)
	}
	if len(stub.calls) != 0 {
		t.Errorf("expected no rent call, got %v", stub.calls)
	}

	// Retrieve from cache
	out, err := s.Retrieve(ctx, cidStr)
	if err != nil || !bytes.Equal(out, data) {
		t.Fatalf("Retrieve cache failed: %v %v", err, out)
	}

	// Retrieve via gateway
	// remove cache entry
	os.Remove(filepath.Join(cfg.CacheDir, cidStr))
	// call Retrieve
	out2, err := s.Retrieve(ctx, cidStr)
	if err != nil || !bytes.Equal(out2, data) {
		t.Fatalf("Retrieve gateway failed: %v %v", err, out2)
	}
}

// Test RegisterVMOpcode wiring
func TestRegisterVMOpcode(t *testing.T) {
	cfg := &StorageConfig{CacheDir: t.TempDir(), CacheSizeEntries: 1, IPFSGateway: "", GatewayTimeout: time.Second}
	s, _ := NewStorage(cfg, logrus.New(), nil)
	dvm := &dummyVM{}
	s.RegisterVMOpcode(dvm)
	if dvm.opcode != PinOpcode {
		t.Errorf("expected opcode %x, got %x", PinOpcode, dvm.opcode)
	}
	if dvm.handler == nil {
		t.Errorf("handler not set")
	}
}

// Test listing create and retrieval helpers
func TestStorage_Listings(t *testing.T) {
	appStore = &InMemoryStore{data: make(map[string][]byte)}
	prov := addrWithByte(0x01)
	_ = appStore.Set([]byte(fmt.Sprintf("identity:provider:%x", prov)), []byte{1})

	listing := &StorageListing{Provider: prov, PricePerGB: 10, CapacityGB: 5}
	if err := CreateListing(listing); err != nil {
		t.Fatalf("create listing: %v", err)
	}

	got, err := GetListing(listing.ID)
	if err != nil || got.ID != listing.ID {
		t.Fatalf("get listing failed: %v", err)
	}

	list, err := ListListings(&prov)
	if err != nil || len(list) != 1 {
		t.Fatalf("list listings failed: %v %v", err, list)
	}
}
