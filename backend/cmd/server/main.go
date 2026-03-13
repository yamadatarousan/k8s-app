package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/user/k8s-app/backend/todoapi"
)

type appConfig struct {
	Port            string
	ApplicationName string
	ShutdownTimeout time.Duration
}

func main() {
	config, err := loadConfig()
	if err != nil {
		log.Fatalf("todo api server config error: %v", err)
	}

	apiServer := todoapi.NewServer(todoapi.Config{
		ApplicationName: config.ApplicationName,
	})

	server := &http.Server{
		Addr:    ":" + config.Port,
		Handler: apiServer,
	}

	log.Printf("todo api server started on :%s name=%s", config.Port, config.ApplicationName)

	serverErrorChannel := make(chan error, 1)
	go func() {
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErrorChannel <- err
		}
		close(serverErrorChannel)
	}()

	signalContext, stop := signalContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	select {
	case err := <-serverErrorChannel:
		if err != nil {
			log.Fatalf("todo api server failed: %v", err)
		}
	case <-signalContext.Done():
		shutdownContext, cancel := context.WithTimeout(context.Background(), config.ShutdownTimeout)
		defer cancel()

		if err := waitForShutdownSignal(shutdownContext, func(shutdownContext context.Context) error {
			if err := apiServer.Shutdown(shutdownContext); err != nil {
				return err
			}

			return server.Shutdown(shutdownContext)
		}); err != nil {
			log.Fatalf("todo api server shutdown failed: %v", err)
		}
	}
}

func loadConfig() (appConfig, error) {
	port := os.Getenv("PORT")
	if port == "" {
		// k8s の Service から扱いやすい既定ポートに固定しておく。
		port = "8080"
	}

	applicationName := os.Getenv("TODO_API_NAME")
	if applicationName == "" {
		applicationName = "todo-api"
	}

	shutdownTimeoutSeconds := 10
	if rawShutdownTimeoutSeconds := os.Getenv("SHUTDOWN_TIMEOUT_SECONDS"); rawShutdownTimeoutSeconds != "" {
		parsedShutdownTimeoutSeconds, err := strconv.Atoi(rawShutdownTimeoutSeconds)
		if err != nil {
			return appConfig{}, err
		}
		if parsedShutdownTimeoutSeconds <= 0 {
			return appConfig{}, errors.New("shutdown timeout must be greater than zero")
		}
		shutdownTimeoutSeconds = parsedShutdownTimeoutSeconds
	}

	return appConfig{
		Port:            port,
		ApplicationName: applicationName,
		ShutdownTimeout: time.Duration(shutdownTimeoutSeconds) * time.Second,
	}, nil
}

func signalContext(parentContext context.Context, signals ...os.Signal) (context.Context, context.CancelFunc) {
	return signal.NotifyContext(parentContext, signals...)
}

func waitForShutdownSignal(parentContext context.Context, shutdown func(context.Context) error) error {
	shutdownTimeout := 10 * time.Second
	if deadline, ok := parentContext.Deadline(); ok {
		shutdownTimeout = time.Until(deadline)
	}

	if shutdownTimeout <= 0 {
		shutdownTimeout = 10 * time.Second
	}

	shutdownContext, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	log.Printf("todo api server shutting down")

	return shutdown(shutdownContext)
}
