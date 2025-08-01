package core

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// TrainingJob represents a long running training process for an AI model.
type TrainingJob struct {
	ID         string            `json:"id"`
	DatasetCID string            `json:"dataset_cid"`
	ModelCID   string            `json:"model_cid"`
	Params     map[string]string `json:"params"`
	Creator    Address           `json:"creator"`
	Status     string            `json:"status"`
	StartedAt  time.Time         `json:"started_at"`
	EndedAt    time.Time         `json:"ended_at,omitempty"`
}

// StartTraining creates a new training job and persists it on chain.
func (ai *AIEngine) StartTraining(datasetCID, modelCID string, params map[string]string, creator Address) (string, error) {
	if ai == nil {
		return "", fmt.Errorf("AI engine not initialised")
	}

	id := uuid.New().String()
	job := TrainingJob{
		ID:         id,
		DatasetCID: datasetCID,
		ModelCID:   modelCID,
		Params:     params,
		Creator:    creator,
		Status:     "running",
		StartedAt:  time.Now(),
	}

	ai.mu.Lock()
	if ai.jobs == nil {
		ai.jobs = make(map[string]TrainingJob)
	}
	ai.jobs[id] = job
	ai.mu.Unlock()

	key := fmt.Sprintf("ai_training:%s", id)
	_ = ai.led.SetState([]byte(key), toJSON(job))
	return id, nil
}

// TrainingStatus returns the stored information for a given training job.
func (ai *AIEngine) TrainingStatus(id string) (TrainingJob, error) {
	if ai == nil {
		return TrainingJob{}, fmt.Errorf("AI engine not initialised")
	}
	ai.mu.RLock()
	job, ok := ai.jobs[id]
	ai.mu.RUnlock()
	if ok {
		return job, nil
	}
	key := fmt.Sprintf("ai_training:%s", id)
	raw, _ := ai.led.GetState([]byte(key))
	if raw == nil {
		return TrainingJob{}, fmt.Errorf("job %s not found", id)
	}
	if err := json.Unmarshal(raw, &job); err != nil {
		return TrainingJob{}, err
	}
	ai.mu.Lock()
	if ai.jobs == nil {
		ai.jobs = make(map[string]TrainingJob)
	}
	ai.jobs[id] = job
	ai.mu.Unlock()
	return job, nil
}

// ListTrainingJobs returns all known training jobs.
func (ai *AIEngine) ListTrainingJobs() ([]TrainingJob, error) {
	if ai == nil {
		return nil, fmt.Errorf("AI engine not initialised")
	}
	ai.mu.RLock()
	out := make([]TrainingJob, 0, len(ai.jobs))
	for _, job := range ai.jobs {
		out = append(out, job)
	}
	ai.mu.RUnlock()
	return out, nil
}

// CancelTraining marks a training job as cancelled.
func (ai *AIEngine) CancelTraining(id string) error {
	if ai == nil {
		return fmt.Errorf("AI engine not initialised")
	}
	ai.mu.Lock()
	job, ok := ai.jobs[id]
	if !ok {
		ai.mu.Unlock()
		return fmt.Errorf("job %s not found", id)
	}
	job.Status = "cancelled"
	job.EndedAt = time.Now()
	ai.jobs[id] = job
	ai.mu.Unlock()

	key := fmt.Sprintf("ai_training:%s", id)
	_ = ai.led.SetState([]byte(key), toJSON(job))
	return nil
}

func toJSON(v interface{}) []byte {
	b, _ := json.Marshal(v)
	return b
}
