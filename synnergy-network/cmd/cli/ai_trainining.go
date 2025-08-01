package cli

import (
	"encoding/json"
	"fmt"
	"github.com/spf13/cobra"
	core "synnergy-network/core"
)

// AITrainingController wraps core AI training helpers.
type AITrainingController struct{}

func (c *AITrainingController) Start(dataset, model string) (string, error) {
	id, err := core.AI().StartTraining(dataset, model, nil, core.ModuleAddress("cli"))
	if err != nil {
		return "", err
	}
	return id, nil
}

func (c *AITrainingController) Status(id string) (core.TrainingJob, error) {
	return core.AI().TrainingStatus(id)
}

func (c *AITrainingController) List() ([]core.TrainingJob, error) {
	return core.AI().ListTrainingJobs()
}

func (c *AITrainingController) Cancel(id string) error {
	return core.AI().CancelTraining(id)
}

// Root command
var aiTrainCmd = &cobra.Command{
	Use:   "ai-train",
	Short: "Manage on-chain AI model training jobs",
}

var aiTrainStartCmd = &cobra.Command{
	Use:  "start [datasetCID] [modelCID]",
	Args: cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctrl := &AITrainingController{}
		id, err := ctrl.Start(args[0], args[1])
		if err != nil {
			return err
		}
		fmt.Println(id)
		return nil
	},
}

var aiTrainStatusCmd = &cobra.Command{
	Use:  "status [jobID]",
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctrl := &AITrainingController{}
		job, err := ctrl.Status(args[0])
		if err != nil {
			return err
		}
		enc, _ := json.MarshalIndent(job, "", "  ")
		fmt.Println(string(enc))
		return nil
	},
}

var aiTrainListCmd = &cobra.Command{
	Use:  "list",
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctrl := &AITrainingController{}
		jobs, err := ctrl.List()
		if err != nil {
			return err
		}
		enc, _ := json.MarshalIndent(jobs, "", "  ")
		fmt.Println(string(enc))
		return nil
	},
}

var aiTrainCancelCmd = &cobra.Command{
	Use:  "cancel [jobID]",
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctrl := &AITrainingController{}
		if err := ctrl.Cancel(args[0]); err != nil {
			return err
		}
		fmt.Println("cancelled")
		return nil
	},
}

func init() {
	aiTrainCmd.AddCommand(aiTrainStartCmd)
	aiTrainCmd.AddCommand(aiTrainStatusCmd)
	aiTrainCmd.AddCommand(aiTrainListCmd)
	aiTrainCmd.AddCommand(aiTrainCancelCmd)
}

// Exported command group
var AITrainingCmd = aiTrainCmd
