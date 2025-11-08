package jobs

import (
	"encoding/json"
	"time"

	"github.com/hibiken/asynq"
)

const (
	TaskTypePriceUpdate = "price:update"
	TaskTypeCleanupData = "data:cleanup"
)

const (
	QueueCritical = "critical"
	QueueDefault  = "default"
	QueueLow      = "low"
)

type PriceUpdatePayload struct {
	TokenAddresses []string `json:"token_addresses"`
}

type CleanupDataPayload struct {
	OlderThan time.Duration `json:"older_than"`
}

func NewPriceUpdateTask(addresses []string) (*asynq.Task, error) {
	payload, err := json.Marshal(PriceUpdatePayload{TokenAddresses: addresses})
	if err != nil {
		return nil, err
	}

	return asynq.NewTask(TaskTypePriceUpdate, payload, asynq.Queue(QueueDefault)), nil
}

func NewCleanupDataTask(olderThan time.Duration) (*asynq.Task, error) {
	payload, err := json.Marshal(CleanupDataPayload{OlderThan: olderThan})
	if err != nil {
		return nil, err
	}

	return asynq.NewTask(TaskTypeCleanupData, payload, asynq.Queue(QueueLow)), nil
}
