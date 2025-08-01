package cli

import (
	"io/ioutil"

	"github.com/spf13/cobra"
	core "synnergy-network/core"
)

// -----------------------------------------------------------------------------
// Controller
// -----------------------------------------------------------------------------
var biometrics = core.NewBiometricsAuth()

func bioEnroll(cmd *cobra.Command, args []string) error {
	addr, _ := cmd.Flags().GetString("address")
	file, _ := cmd.Flags().GetString("file")
	data, err := ioutil.ReadFile(file)
	if err != nil {
		return err
	}
	return biometrics.Enroll(addr, data)
}

func bioVerify(cmd *cobra.Command, args []string) error {
	addr, _ := cmd.Flags().GetString("address")
	file, _ := cmd.Flags().GetString("file")
	data, err := ioutil.ReadFile(file)
	if err != nil {
		return err
	}
	if biometrics.Verify(addr, data) {
		cmd.Println("match")
	} else {
		cmd.Println("mismatch")
	}
	return nil
}

func bioDelete(cmd *cobra.Command, args []string) error {
	addr, _ := cmd.Flags().GetString("address")
	biometrics.Delete(addr)
	return nil
}

// -----------------------------------------------------------------------------
// Cobra wiring
// -----------------------------------------------------------------------------
var bioCmd = &cobra.Command{
	Use:   "biometrics",
	Short: "Manage biometric authentication templates",
}

var bioEnrollCmd = &cobra.Command{
	Use:   "enroll",
	Short: "enroll biometric data from file",
	RunE:  bioEnroll,
}

var bioVerifyCmd = &cobra.Command{
	Use:   "verify",
	Short: "verify biometric data from file",
	RunE:  bioVerify,
}

var bioDeleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "delete biometric data for address",
	RunE:  bioDelete,
}

func init() {
	bioEnrollCmd.Flags().String("address", "", "address")
	bioEnrollCmd.Flags().String("file", "", "data file")
	bioEnrollCmd.MarkFlagRequired("address")
	bioEnrollCmd.MarkFlagRequired("file")

	bioVerifyCmd.Flags().String("address", "", "address")
	bioVerifyCmd.Flags().String("file", "", "data file")
	bioVerifyCmd.MarkFlagRequired("address")
	bioVerifyCmd.MarkFlagRequired("file")

	bioDeleteCmd.Flags().String("address", "", "address")
	bioDeleteCmd.MarkFlagRequired("address")

	bioCmd.AddCommand(bioEnrollCmd, bioVerifyCmd, bioDeleteCmd)
}

// Exported command variable
var BioCmd = bioCmd

func RegisterBiometrics(root *cobra.Command) { root.AddCommand(BioCmd) }
