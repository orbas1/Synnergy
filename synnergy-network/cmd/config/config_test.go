package config

import (
	"os"
	"testing"

	"github.com/spf13/viper"

	"synnergy-network/internal/testutil"
)

func TestLoadConfigDefault(t *testing.T) {
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd failed: %v", err)
	}
	defer os.Chdir(wd)
	viper.Reset()

	if err := os.Chdir(".."); err != nil {
		t.Fatalf("chdir failed: %v", err)
	}
	LoadConfig("")
	if AppConfig.Network.ID != "synnergy-mainnet" {
		t.Fatalf("unexpected network id: %s", AppConfig.Network.ID)
	}
}

func TestLoadConfigOverride(t *testing.T) {
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd failed: %v", err)
	}
	defer os.Chdir(wd)
	viper.Reset()

	if err := os.Chdir(".."); err != nil {
		t.Fatalf("chdir failed: %v", err)
	}
	LoadConfig("bootstrap")
	if AppConfig.Network.MaxPeers != 100 {
		t.Fatalf("expected MaxPeers 100, got %d", AppConfig.Network.MaxPeers)
	}
	if AppConfig.Network.DiscoveryTag != "synnergy-bootstrap" {
		t.Fatalf("expected discovery tag override")
	}
}

func TestLoadConfigSandbox(t *testing.T) {
	sb, err := testutil.NewSandbox()
	if err != nil {
		t.Fatalf("NewSandbox failed: %v", err)
	}
	defer sb.Cleanup()

	if err := os.Mkdir(sb.Path("config"), 0700); err != nil {
		t.Fatalf("Mkdir failed: %v", err)
	}

	data := []byte("network:\n  id: sandbox\n  max_peers: 42\n")
	if err := sb.WriteFile("config/default.yaml", data, 0600); err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}

	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd failed: %v", err)
	}
	defer os.Chdir(wd)
	viper.Reset()

	if err := os.Chdir(sb.Root); err != nil {
		t.Fatalf("chdir failed: %v", err)
	}
	LoadConfig("")

	if AppConfig.Network.ID != "sandbox" {
		t.Fatalf("expected network id sandbox, got %s", AppConfig.Network.ID)
	}
	if AppConfig.Network.MaxPeers != 42 {
		t.Fatalf("expected MaxPeers 42, got %d", AppConfig.Network.MaxPeers)
	}
}
