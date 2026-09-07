package jobs

import (
	"log"
	"math"
	"math/rand/v2"
	"time"
)

func (s *Store) ProcessJob(id string) {
	const (
		maxDelay = 30 * time.Second
		maxRetry = 10
	)

	job, exist := s.Get(id)

	if !exist {
		log.Printf("Job %s not found\n", id)
		return
	}

	if job.RetryCount == maxRetry {
		job.Status = StatusDeadLetter

		s.Update(job)
		return
	}

	randNum := rand.IntN(2)
	log.Printf("%v", randNum)

	job.RetryCount++

	if randNum == 0 {
		job.Status = StatusFailed

		base := math.Pow(2, float64(job.RetryCount))
		delay := time.Duration(base * float64(time.Second))
		delay = min(delay, maxDelay)

		delay += time.Duration(rand.Float64() * 0.25 * float64(delay))
		job.NextRetryAt = time.Now().Add(delay)
	} else {
		job.Status = StatusCompleted
	}

	s.Update(job)

	log.Printf("Job %s Procesed with Status: %s\n", job.ID, job.Status)
}
