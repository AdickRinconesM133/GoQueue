package main

import (
	"time"

	serv "GoQueue/cmd/worker"
	handler "GoQueue/internal/api"
	"GoQueue/internal/redisclient"
	"GoQueue/internal/store"
	"GoQueue/internal/worker"
)

func main() {
	rdb := redisclient.InitRedis()
	s := store.NewStore(rdb)
	w := worker.NewWorker(s)
	h := handler.NewHandler(s, w)

	serv.StartServer(h.CreateJobHandler)

	go func() {
		ticker := time.NewTicker(2 * time.Second)
		for range ticker.C {
			w.CheckPendingJobs()
		}
	}()
}
