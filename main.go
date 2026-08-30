package main

import (
	"encoding/json"
	"log"
	"math"
	"math/rand/v2"
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

func jobHandler(id string) {
	const (
		maxDelay = 30 * time.Second
		maxRetry = 10
	)

	mu.Lock()
	job, exist := jobStore[id]
	mu.Unlock()

	if !exist {
		log.Printf("Job %s not found\n", id)
		return
	}

	if job.RetryCount == maxRetry {
		job.Status = StatusDeadLetter

		mu.Lock()
		jobStore[id] = job
		mu.Unlock()
		return
	}

	randNum := rand.IntN(2)
	log.Printf("%v", randNum)

	job.RetryCount++

	if randNum == 0 {
		job.Status = StatusFailed

		base := math.Pow(2, float64(job.RetryCount))
		delay := time.Duration(base * float64(time.Second))

		if delay > maxDelay {
			delay = maxDelay
		}

		delay += time.Duration(rand.Float64() * 0.25 * float64(delay))
		job.NextRetryAt = time.Now().Add(delay)
	} else {
		job.Status = StatusCompleted
	}

	mu.Lock()
	jobStore[id] = job
	mu.Unlock()

	log.Printf("Job %s Procesed with Status: %s\n", job.ID, job.Status)
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

	go jobHandler(job.ID)

	log.Printf("Received job: %+v\n", job)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(job)
}

func checkPendingJobs() {
	mu.Lock()
	snapshot := make(map[string]Job, len(jobStore))

	for id, job := range jobStore {
		snapshot[id] = job
	}
	mu.Unlock()

	for id, job := range snapshot {
		if job.Status == StatusFailed && job.NextRetryAt.Before(time.Now()) {
			log.Printf("Retrying job: %s", id)
			go jobHandler(id)
		}
	}
}

func main() {
	r := chi.NewRouter()
	r.Post("/jobs", createJobHandler)

	go func() {
		ticker := time.NewTicker(2 * time.Second)
		for range ticker.C {
			checkPendingJobs()
		}
	}()

	log.Printf("server running on: 8080")
	http.ListenAndServe(":8080", r)
}
