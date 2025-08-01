package core

import (
	"fmt"
	"sync"
)

// Workflow represents a sequence of opcode names executed in order.
type Workflow struct {
	ID      string
	Actions []string
	Trigger string
	Webhook string
}

var (
	workflows   = make(map[string]*Workflow)
	workflowsMu sync.RWMutex
)

// NewWorkflow creates a new workflow identified by id.
func NewWorkflow(id string) (*Workflow, error) {
	workflowsMu.Lock()
	defer workflowsMu.Unlock()
	if _, exists := workflows[id]; exists {
		return nil, fmt.Errorf("workflow %s already exists", id)
	}
	wf := &Workflow{ID: id}
	workflows[id] = wf
	return wf, nil
}

// AddWorkflowAction appends an opcode name to the workflow.
func AddWorkflowAction(id, fn string) error {
	workflowsMu.Lock()
	defer workflowsMu.Unlock()
	wf, ok := workflows[id]
	if !ok {
		return fmt.Errorf("workflow %s not found", id)
	}
	if _, ok := nameToOp[fn]; !ok {
		return fmt.Errorf("unknown function %s", fn)
	}
	wf.Actions = append(wf.Actions, fn)
	return nil
}

// SetWorkflowTrigger sets a cron expression or event trigger for the workflow.
func SetWorkflowTrigger(id, trigger string) error {
	workflowsMu.Lock()
	defer workflowsMu.Unlock()
	wf, ok := workflows[id]
	if !ok {
		return fmt.Errorf("workflow %s not found", id)
	}
	wf.Trigger = trigger
	return nil
}

// SetWebhook registers a webhook URL to be called after execution.
func SetWebhook(id, url string) error {
	workflowsMu.Lock()
	defer workflowsMu.Unlock()
	wf, ok := workflows[id]
	if !ok {
		return fmt.Errorf("workflow %s not found", id)
	}
	wf.Webhook = url
	return nil
}

// ExecuteWorkflow executes each action sequentially using the provided context.
func ExecuteWorkflow(ctx OpContext, id string) error {
	workflowsMu.RLock()
	wf, ok := workflows[id]
	workflowsMu.RUnlock()
	if !ok {
		return fmt.Errorf("workflow %s not found", id)
	}
	for _, fn := range wf.Actions {
		op, ok := nameToOp[fn]
		if !ok {
			return fmt.Errorf("unknown function %s", fn)
		}
		if err := Dispatch(ctx, op); err != nil {
			return fmt.Errorf("execute %s: %w", fn, err)
		}
	}
	return nil
}

// ListWorkflows returns all registered workflow IDs.
func ListWorkflows() []string {
	workflowsMu.RLock()
	defer workflowsMu.RUnlock()
	ids := make([]string, 0, len(workflows))
	for id := range workflows {
		ids = append(ids, id)
	}
	return ids
}
