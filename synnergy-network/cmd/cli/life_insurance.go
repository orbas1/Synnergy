package cli

import (
	"encoding/json"
	"github.com/spf13/cobra"
	tokens "synnergy-network/core/Tokens"
	"time"
)

// lifeCmd exposes basic utilities for the SYN2800 token.
var lifeCmd = &cobra.Command{
	Use:   "life",
	Short: "Utilities for SYN2800 life insurance tokens",
}

var lifeExampleCmd = &cobra.Command{
	Use:   "example",
	Short: "Show example life insurance policy metadata",
	RunE: func(cmd *cobra.Command, _ []string) error {
		pol := tokens.LifePolicy{
			PolicyID:       "POLICY123",
			Insured:        "John Doe",
			Beneficiary:    "Jane Doe",
			Premium:        1000,
			CoverageAmount: 50000,
			StartDate:      time.Now().UTC(),
			EndDate:        time.Now().AddDate(1, 0, 0).UTC(),
			Active:         true,
		}
		enc := json.NewEncoder(cmd.OutOrStdout())
		enc.SetIndent("", "  ")
		return enc.Encode(pol)
	},
}

func init() {
	lifeCmd.AddCommand(lifeExampleCmd)
}

// LifeCmd is the exported command used by the CLI index.
var LifeCmd = lifeCmd
