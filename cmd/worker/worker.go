package worker

import (
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
)

func StartServer(h http.HandlerFunc) {
	r := chi.NewRouter()
	r.Post("/jobs", h)

	if err := http.ListenAndServe(":8080", r); err != nil {
		log.Fatalf("Error starting server: %v", err)
	} else {
		log.Printf("server running on: 8080")
	}
}
