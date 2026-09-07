package main

import (
	"log"
	"net/http"
	"time"

	"GoQueue/internal/jobs"
	"GoQueue/internal/redisclient"

	"github.com/go-chi/chi/v5"
)

func main() {
	rdb := redisclient.InitRedis()
	s := jobs.NewStore(rdb)
	r := chi.NewRouter()

	r.Post("/jobs", s.CreateJobHandler)

	go func() {
		ticker := time.NewTicker(2 * time.Second)
		for range ticker.C {
			s.CheckPendingJobs()
		}
	}()

	log.Printf("server running on: 8080")
	http.ListenAndServe(":8080", r)
}
