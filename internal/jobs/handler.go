package jobs

import (
	"encoding/json"
	"fmt"
	"log"
	"maps"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

func (s *Store) CreateJobHandler(w http.ResponseWriter, r *http.Request) {
	var job Job
	idempotencyKey := r.Header.Get("Idempotency-Key")

	if idempotencyKey == "" {
		http.Error(w, "Header must contain Idempotency-Key", http.StatusBadRequest)
		return
	}

	if err := json.NewDecoder(r.Body).Decode(&job); err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}

	jobIdempotencyKey := fmt.Sprintf("job:%s", idempotencyKey)
	previousValue, err := s.rdb.SetArgs(s.ctx, jobIdempotencyKey, "PROCESSING", redis.SetArgs{
		TTL:  5 * time.Minute,
		Mode: "NX",
		Get:  true,
	}).Result()

	if err == nil {
		log.Printf("Job already exist: %s", previousValue)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusAccepted)
		w.Write([]byte(previousValue))
		return
	} else if err != redis.Nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	job.Status = StatusPending
	job.ID = uuid.NewString()
	job.IdempotencyKey = jobIdempotencyKey

	s.Create(job)
	s.rdb.Set(s.ctx, jobIdempotencyKey, job, 0)

	go s.ProcessJob(job.ID)

	log.Printf("Received job: %+v\n", job)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(job)
}

func (s *Store) CheckPendingJobs() {
	s.mu.Lock()
	snapshot := make(map[string]Job, len(s.jobStore))
	maps.Copy(snapshot, s.jobStore)

	s.mu.Unlock()

	for id, job := range snapshot {
		if job.Status == StatusFailed && job.NextRetryAt.Before(time.Now()) {
			log.Printf("Retrying job: %s", id)
			go s.ProcessJob(id)
		}
	}
}
