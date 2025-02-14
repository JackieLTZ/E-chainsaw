package main

import (
	"a/handlers"
	"context"
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
	defer db.Close()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	server := &http.Server{
		Addr:         ":" + port,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	userHandler := handlers.NewHandler(db)

	mux := http.NewServeMux()
	mux.Handle("/ping", loggingMiddleware(http.HandlerFunc(userHandler.Ping)))
	mux.Handle("/users", loggingMiddleware(http.HandlerFunc(userHandler.CreateUser)))
	mux.Handle("/u", loggingMiddleware(http.HandlerFunc(userHandler.GetUsers)))

	server.Handler = mux

	serverCtx, serverStopCtx := context.WithCancel(context.Background())

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	go func() {
		<-sig

		shutdownCtx, cancelShutdown := context.WithTimeout(serverCtx, 30*time.Second)
		defer cancelShutdown()

		go func() {
			<-shutdownCtx.Done()
			if shutdownCtx.Err() == context.DeadlineExceeded {
				log.Fatal("graceful shutdown timed out.. forcing exit.")
			}
		}()

		err := server.Shutdown(shutdownCtx)
		if err != nil {
			log.Printf("error shutting down server: %v", err)
		}
		serverStopCtx()
	}()

	log.Printf("Server starting on port %s...\n", port)
	err = server.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		log.Printf("error starting server: %v", err)
		os.Exit(1)
	}

	<-serverCtx.Done()
	log.Println("Server stopped")
}
