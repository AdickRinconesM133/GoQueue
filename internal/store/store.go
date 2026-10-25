package store

import (
	"context"
	"maps"
	"sync"
	"time"

	"GoQueue/internal/job"

	"github.com/redis/go-redis/v9"
)

type Store struct {
	mu       sync.Mutex
	jobStore map[string]job.Job
	rdb      *redis.Client
	ctx      context.Context
}

func NewStore(rdb *redis.Client) *Store {
	return &Store{
		jobStore: make(map[string]job.Job),
		rdb:      rdb,
		ctx:      context.Background(),
	}
}

func (s *Store) GetRdb() *redis.Client {
	return s.rdb
}

func (s *Store) GetCtx() context.Context {
	return s.ctx
}

func (s *Store) GetJobStore() map[string]job.Job {
	return s.jobStore
}

func (s *Store) Get(id string) (job.Job, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	job, exist := s.jobStore[id]
	return job, exist
}

func (s *Store) GetFailedReady() []job.Job {
	s.mu.Lock()
	defer s.mu.Unlock()

	var failed []job.Job
	for _, jobStored := range s.jobStore {
		if jobStored.Status == job.StatusFailed && jobStored.NextRetryAt.Before(time.Now()) {
			failed = append(failed, jobStored)
		}
	}

	return failed
}

func (s *Store) Create(job job.Job) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exist := s.jobStore[job.ID]; exist {
		return false
	}

	s.jobStore[job.ID] = job
	return true
}

func (s *Store) Update(job job.Job) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.jobStore[job.ID] = job
}

func (s *Store) Copy(snapshot map[string]job.Job) {
	s.mu.Lock()
	defer s.mu.Unlock()

	maps.Copy(snapshot, s.jobStore)
}
