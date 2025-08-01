package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	core "synnergy-network/core"
)

// middleware to ensure nothing for now

// Controller wraps core functions.
type workflowController struct{}

func (w *workflowController) new(id string) error {
	_, err := core.NewWorkflow(id)
	return err
}
func (w *workflowController) add(id, action string) error {
	return core.AddWorkflowAction(id, action)
}
func (w *workflowController) trigger(id, expr string) error {
	return core.SetWorkflowTrigger(id, expr)
}
func (w *workflowController) webhook(id, url string) error {
	return core.SetWebhook(id, url)
}
func (w *workflowController) run(cmd *cobra.Command, id string) error {
	ctx := &cliOpCtx{cmd}
	return core.ExecuteWorkflow(ctx, id)
}

// cliOpCtx is a minimal OpContext implementation for CLI use.
type cliOpCtx struct{ cmd *cobra.Command }

func (c *cliOpCtx) Call(name string) error { return fmt.Errorf("call %s not implemented", name) }
func (c *cliOpCtx) Gas(g uint64) error     { return nil }

var wfCmd = &cobra.Command{
	Use:   "workflow",
	Short: "Manage automation workflows",
}

var wfNewCmd = &cobra.Command{
	Use:  "new [id]",
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctrl := &workflowController{}
		if err := ctrl.new(args[0]); err != nil {
			return err
		}
		fmt.Fprintln(cmd.OutOrStdout(), "workflow created")
		return nil
	},
}

var wfAddCmd = &cobra.Command{
	Use:  "add [id] [function]",
	Args: cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctrl := &workflowController{}
		if err := ctrl.add(args[0], args[1]); err != nil {
			return err
		}
		fmt.Fprintln(cmd.OutOrStdout(), "action added")
		return nil
	},
}

var wfTriggerCmd = &cobra.Command{
	Use:  "trigger [id] [cron]",
	Args: cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctrl := &workflowController{}
		if err := ctrl.trigger(args[0], args[1]); err != nil {
			return err
		}
		fmt.Fprintln(cmd.OutOrStdout(), "trigger set")
		return nil
	},
}

var wfWebhookCmd = &cobra.Command{
	Use:  "webhook [id] [url]",
	Args: cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctrl := &workflowController{}
		if err := ctrl.webhook(args[0], args[1]); err != nil {
			return err
		}
		fmt.Fprintln(cmd.OutOrStdout(), "webhook set")
		return nil
	},
}

var wfRunCmd = &cobra.Command{
	Use:  "run [id]",
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctrl := &workflowController{}
		return ctrl.run(cmd, args[0])
	},
}

func init() {
	wfCmd.AddCommand(wfNewCmd)
	wfCmd.AddCommand(wfAddCmd)
	wfCmd.AddCommand(wfTriggerCmd)
	wfCmd.AddCommand(wfWebhookCmd)
	wfCmd.AddCommand(wfRunCmd)
}

// WorkflowCmd exported for index registration
var WorkflowCmd = wfCmd
