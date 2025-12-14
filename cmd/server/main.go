package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/cazlo/go-outside-in-testing-strategy-example/internal/app"
	"github.com/cazlo/go-outside-in-testing-strategy-example/internal/httpclient"
)

func main() {
	externalURL := getenv("EXTERNAL_URL", "https://httpbin.org/status/204")
	addr := getenv("ADDR", ":8080")

	a := &app.App{
		ExternalURL: externalURL,
		HTTPClient: &httpclient.DefaultClient{
			Client: &http.Client{
				Timeout: 5 * time.Second,
			},
		},
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/hello", a.HelloHandler)

	server := &http.Server{
		Addr:              addr,
		Handler:           mux,
		ReadHeaderTimeout: 10 * time.Second,
	}

	// Channel to listen for interrupt or terminate signals
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	// Run server in a goroutine
	go func() {
		log.Printf("listening on %s", addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
	}()

	// Wait for interrupt signal
	<-stop
	log.Println("shutting down server...")

	// Create a deadline for graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Gracefully shutdown the server
	if err := server.Shutdown(ctx); err != nil {
		log.Printf("server shutdown error: %v", err)
	}

	log.Println("server stopped")
}

func getenv(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}
