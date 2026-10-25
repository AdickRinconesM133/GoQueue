package handler

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"

	"GoQueue/internal/job"
	"GoQueue/internal/store"
	"GoQueue/internal/worker"
)

type Handler struct {
	store  *store.Store
	worker *worker.Worker
}

func NewHandler(s *store.Store, w *worker.Worker) *Handler {
	return &Handler{
		store:  s,
		worker: w,
	}
}

func (h *Handler) CreateJobHandler(w http.ResponseWriter, r *http.Request) {
	var newJob job.Job
	idempotencyKey := r.Header.Get("Idempotency-Key")

	if idempotencyKey == "" {
		http.Error(w, "Header must contain Idempotency-Key", http.StatusBadRequest)
		return
	}

	if err := json.NewDecoder(r.Body).Decode(&newJob); err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}

	jobIdempotencyKey := fmt.Sprintf("job:%s", idempotencyKey)
	previousValue, err := h.store.GetRdb().SetArgs(h.store.GetCtx(), jobIdempotencyKey, "PROCESSING", redis.SetArgs{
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

	newJob.Status = job.StatusPending
	newJob.ID = uuid.NewString()
	newJob.IdempotencyKey = jobIdempotencyKey

	h.store.Create(newJob)
	h.store.GetRdb().Set(h.store.GetCtx(), jobIdempotencyKey, newJob, 0)

	go h.worker.ProcessJob(newJob.ID)

	log.Printf("Received job: %+v\n", newJob)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(newJob)
}
