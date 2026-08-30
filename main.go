package main

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type JobStatus int

type Job struct {
	ID          string          `json:"id"`
	Type        string          `json:"type"`
	Payload     json.RawMessage `json:"payload"`
	Status      JobStatus       `json:"status"`
	RetryCount  int             `json:"retry_count"`
	NextRetryAt time.Time       `json:"next_retry_at"`
}

var (
	jobStore = make(map[string]Job)
	mu       sync.Mutex
)

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

func createJobHandler(w http.ResponseWriter, r *http.Request) {
	var job Job
	if err := json.NewDecoder(r.Body).Decode(&job); err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}

	job.Status = StatusPending
	job.ID = uuid.NewString()

	mu.Lock()
	jobStore[job.ID] = job
	mu.Unlock()

	log.Printf("Received job: %+v\n", job)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(job)
}

func main() {
	r := chi.NewRouter()
	r.Post("/jobs", createJobHandler)

	log.Printf("server running on: 8080")
	http.ListenAndServe(":8080", r)
}
