package worker

import (
	"encoding/json"
	"log"
	"math"
	"math/rand/v2"
	"time"

	"GoQueue/internal/job"
	"GoQueue/internal/store"
)

type Worker struct {
	store *store.Store
}

func NewWorker(s *store.Store) *Worker {
	return &Worker{
		store: s,
	}
}

func (w *Worker) ProcessJob(id string) {
	const (
		maxDelay = 30 * time.Second
		maxRetry = 10
	)

	jobStored, exist := w.store.Get(id)

	if !exist {
		log.Printf("Job %s not found\n", id)
		return
	}

	if jobStored.RetryCount == maxRetry {
		jobStored.Status = job.StatusDeadLetter

		w.store.Update(jobStored)
		return
	}

	randNum := rand.IntN(2)
	log.Printf("%v", randNum)

	jobStored.RetryCount++

	if randNum == 0 {
		jobStored.Status = job.StatusFailed

		base := math.Pow(2, float64(jobStored.RetryCount))
		delay := time.Duration(base * float64(time.Second))
		delay = min(delay, maxDelay)

		delay += time.Duration(rand.Float64() * 0.25 * float64(delay))
		jobStored.NextRetryAt = time.Now().Add(delay)
	} else {
		jobStored.Status = job.StatusCompleted
	}
	w.store.Update(jobStored)
	log.Printf("Job %s Procesed with Status: %s\n", jobStored.ID, jobStored.Status)

	jobJSON, err := json.Marshal(jobStored)
	if err != nil {
		log.Printf("Failed to Marshal: %s", err.Error())
		return
	}

	_, err = w.store.GetRdb().Set(w.store.GetCtx(), jobStored.IdempotencyKey, jobJSON, 0).Result()
	if err != nil {
		log.Printf("Failed to store Job in Redis: %s", jobJSON)
	}
}

func (w *Worker) CheckPendingJobs() {
	snapshot := make(map[string]job.Job, len(w.store.GetJobStore()))
	w.store.Copy(snapshot)

	for id, jobStored := range snapshot {
		if jobStored.Status == job.StatusFailed && jobStored.NextRetryAt.Before(time.Now()) {
			log.Printf("Retrying job: %s", id)
			go w.ProcessJob(id)
		}
	}
}
