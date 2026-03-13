package main

import (
	"log"
	"net/http"
	"os"

	"github.com/user/k8s-app/backend/todoapi"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		// k8s の Service から扱いやすい既定ポートに固定しておく。
		port = "8080"
	}

	server := &http.Server{
		Addr:    ":" + port,
		Handler: todoapi.NewServer(),
	}

	log.Printf("todo api server started on :%s", port)

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("todo api server failed: %v", err)
	}
}
