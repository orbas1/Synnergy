package core

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"time"
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

// StartTraining creates a new training job by invoking the remote AI service
// and persists the metadata on chain.
func (ai *AIEngine) StartTraining(datasetCID, modelCID string, params map[string]string, creator Address) (string, error) {
	if ai == nil {
		return "", fmt.Errorf("AI engine not initialised")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	resp, err := ai.client.StartTraining(ctx, &TrainingRequest{DatasetCID: datasetCID, ModelCID: modelCID, Params: params})
	if err != nil {
		return "", err
	}

	id := resp.JobID
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

	if err := ai.persistTrainingJob(job); err != nil {
		return "", err
	}
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
	if !ok {
		key := fmt.Sprintf("ai_training:%s", id)
		raw, _ := ai.led.GetState([]byte(key))
		if raw == nil {
			return TrainingJob{}, fmt.Errorf("job %s not found", id)
		}
		if err := json.Unmarshal(raw, &job); err != nil {
			return TrainingJob{}, err
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if resp, err := ai.client.TrainingStatus(ctx, &TrainingStatusRequest{JobID: id}); err == nil {
		job.Status = resp.Status
		if resp.Status == "completed" || resp.Status == "failed" {
			job.EndedAt = time.Now()
		}
	}

	ai.mu.Lock()
	if ai.jobs == nil {
		ai.jobs = make(map[string]TrainingJob)
	}
	ai.jobs[id] = job
	ai.mu.Unlock()
	if err := ai.persistTrainingJob(job); err != nil {
		return job, err
	}
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

	return ai.persistTrainingJob(job)
}

// CompleteTraining marks a job as finished and persists resulting parameters securely.
func (ai *AIEngine) CompleteTraining(id string, modelParams []byte) error {
	if ai == nil {
		return fmt.Errorf("AI engine not initialised")
	}
	ai.mu.Lock()
	job, ok := ai.jobs[id]
	if !ok {
		ai.mu.Unlock()
		return fmt.Errorf("job %s not found", id)
	}
	job.Status = "completed"
	job.EndedAt = time.Now()
	ai.jobs[id] = job
	ai.mu.Unlock()

	if err := ai.persistTrainingJob(job); err != nil {
		return err
	}

	hash := sha256.Sum256([]byte(job.ModelCID))
	if err := ai.StoreModelParams(hash, modelParams); err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, err := ai.client.UploadModel(ctx, &ModelUploadRequest{Model: modelParams, CID: job.ModelCID})
	return err
}

func (ai *AIEngine) persistTrainingJob(job TrainingJob) error {
	b, err := json.Marshal(job)
	if err != nil {
		return err
	}
	key := fmt.Sprintf("ai_training:%s", job.ID)
	return ai.led.SetState([]byte(key), b)
}
