package api

import (
	"a/handlers"
	"context"
	"database/sql"
	"errors"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"
)

func RunServer(port string, handler http.Handler) error {

	server := &http.Server{
		Addr:         ":" + port,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	server.Handler = handler

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	defer cancel()

	serr := make(chan error, 1)

	go func() { serr <- server.ListenAndServe() }()

	var servErr error
	select {
	case servErr = <-serr:
		if servErr != nil {
			log.Printf("Server error: %v", servErr)
		}
	case <-ctx.Done():
		log.Println("Shutdown signal received")
	}

	sdctx, sdcancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer sdcancel()

	shutdownErr := server.Shutdown(sdctx)
	switch {
	case shutdownErr == context.DeadlineExceeded:
		return errors.New("graceful shutdown timed out")
	case shutdownErr != nil:
		log.Printf("Error during shutdown: %v", shutdownErr)
		return errors.Join(servErr, shutdownErr)
	case servErr != nil && !errors.Is(servErr, http.ErrServerClosed):
		return servErr
	default:
		log.Println("Server shutdown completed successfully")
		return nil
	}
}

func Mount(db *sql.DB) http.Handler {

	mux := http.NewServeMux()

	userHandler := handlers.NewHandler(db)

	mux.Handle("GET /ping", loggingMiddleware(http.HandlerFunc(userHandler.Ping)))
	mux.Handle("GET /users/{id}", loggingMiddleware(http.HandlerFunc(userHandler.GetUser)))
	mux.Handle("GET /users", loggingMiddleware(http.HandlerFunc(userHandler.GetUsers)))
	mux.Handle("POST /users", loggingMiddleware(http.HandlerFunc(userHandler.CreateUser)))
	mux.Handle("DELETE /users/{id}", loggingMiddleware(http.HandlerFunc(userHandler.DeleteUser)))

	return mux

}
