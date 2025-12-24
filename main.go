package main

import (
	"log"
	"net/http"
)

func main() {
	const filePathRoot = "."
	const port = "8080"

	cfg := &apiConfig{}

	mux := http.NewServeMux()

	mux.HandleFunc("GET /api/healthz", cfg.handlerHealthz)
	mux.HandleFunc("GET /admin/metrics", cfg.handlerMetrics)
	mux.HandleFunc("POST /admin/reset", cfg.handlerReset)

	fileServer := http.FileServer(http.Dir(filePathRoot))
	mux.Handle("/app/", http.StripPrefix("/app/", cfg.middlewareMetricsInc(fileServer)))

	server := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	log.Printf("Starting server on %s", server.Addr)
	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
