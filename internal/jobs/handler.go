package jobs

import (
	"encoding/json"
	"log"
	"maps"
	"net/http"
	"time"

	"github.com/google/uuid"
)

func (s *Store) CreateJobHandler(w http.ResponseWriter, r *http.Request) {
	var job Job
	if err := json.NewDecoder(r.Body).Decode(&job); err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}

	job.Status = StatusPending
	job.ID = uuid.NewString()

	s.Create(job)

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
