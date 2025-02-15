package main

import (
	"a/handlers"
	"context"
	"database/sql"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s", r.Method, r.RequestURI)
		next.ServeHTTP(w, r)
	})
}

func main() {

	db, err := DB_Connect()
	if err != nil {
		log.Fatalf("error connecting to database: %v", err)
	}

	defer func() {
		if err := db.Close(); err != nil {
			log.Printf("Error closing database connection: %v", err)
		}
		log.Println("Database connection closed.")
	}()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	err = runServer(port, db)
	if err != nil {
		log.Fatalf("Server exited with error: %v", err)
	}

}

func runServer(port string, db *sql.DB) error {

	server := &http.Server{
		Addr:         ":" + port,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	userHandler := handlers.NewHandler(db)

	mux := http.NewServeMux()
	mux.Handle("GET /ping", loggingMiddleware(http.HandlerFunc(userHandler.Ping)))
	mux.Handle("GET /users/{id}", loggingMiddleware(http.HandlerFunc(userHandler.GetUser)))
	mux.Handle("GET /users", loggingMiddleware(http.HandlerFunc(userHandler.GetUsers)))
	mux.Handle("POST /users", loggingMiddleware(http.HandlerFunc(userHandler.CreateUser)))
	mux.Handle("DELETE /users/{id}", loggingMiddleware(http.HandlerFunc(userHandler.DeleteUser)))

	server.Handler = mux

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
