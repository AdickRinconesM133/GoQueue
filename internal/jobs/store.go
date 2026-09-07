package jobs

import (
	"context"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

type Store struct {
	mu       sync.Mutex
	jobStore map[string]Job
	rdb      *redis.Client
	ctx      context.Context
}

func NewStore(rdb *redis.Client) *Store {
	return &Store{
		jobStore: make(map[string]Job),
		rdb:      rdb,
		ctx:      context.Background(),
	}
}

func (s *Store) Get(id string) (Job, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	job, exist := s.jobStore[id]
	return job, exist
}

func (s *Store) GetFailedReady() []Job {
	s.mu.Lock()
	defer s.mu.Unlock()

	var failed []Job
	for _, job := range s.jobStore {
		if job.Status == StatusFailed && job.NextRetryAt.Before(time.Now()) {
			failed = append(failed, job)
		}
	}

	return failed
}

func (s *Store) Create(job Job) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exist := s.jobStore[job.ID]; exist {
		return false
	}

	s.jobStore[job.ID] = job
	return true
}

func (s *Store) Update(job Job) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.jobStore[job.ID] = job
}
