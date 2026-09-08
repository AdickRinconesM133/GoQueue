// Package jobs gestiona la creacion, almacenamiento y
// procesamiento de tareas en segundo plano (workers)
package jobs

import (
	"encoding/json"
	"fmt"
	"time"
)

type JobStatus int

type Job struct {
	ID             string          `json:"id"`
	IdempotencyKey string          `json:"idempotencykey"`
	Type           string          `json:"type"`
	Payload        json.RawMessage `json:"payload"`
	Status         JobStatus       `json:"status"`
	RetryCount     int             `json:"retry_count"`
	NextRetryAt    time.Time       `json:"next_retry_at"`
}

const (
	StatusPending JobStatus = iota
	StatusProcessing
	StatusCompleted
	StatusFailed
	StatusDeadLetter
)

func (s JobStatus) String() string {
	switch s {
	case StatusPending:
		return "pending"
	case StatusProcessing:
		return "processing"
	case StatusCompleted:
		return "completed"
	case StatusFailed:
		return "failed"
	case StatusDeadLetter:
		return "dead_letter"
	default:
		return "unknown"
	}
}

func (s JobStatus) MarshalJSON() ([]byte, error) {
	return json.Marshal(s.String())
}

func (s *JobStatus) UnmarshalJSON(b []byte) error {
	var str string
	if err := json.Unmarshal(b, &str); err != nil {
		return err
	}

	switch str {
	case "pending":
		*s = StatusPending
	case "processing":
		*s = StatusProcessing
	case "completed":
		*s = StatusCompleted
	case "failed":
		*s = StatusFailed
	case "dead_letter":
		*s = StatusDeadLetter
	default:
		return fmt.Errorf("invalid JobStatus: %s", str)
	}

	return nil
}
