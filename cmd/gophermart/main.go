package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/riouske/gophermart/internal/config"
	"github.com/riouske/gophermart/internal/db"
	"github.com/riouske/gophermart/internal/handler/gophermart/user"
	"github.com/riouske/gophermart/internal/repository"
	"github.com/riouske/gophermart/internal/service"
)

func main() {
	log.Println("Starting gophermart app")

	cfg := config.New()

	database, err := db.NewDB(cfg.DatabaseURI)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer database.Close()

	userRepo := repository.NewUserRepository(database)
	authService := service.NewAuthService(userRepo, cfg.JWTSecretKey)

	registerHandler := user.NewRegisterHandler(authService)
	loginHandler := user.NewLoginHandler(authService)

	mux := http.NewServeMux()

	mux.Handle("/api/user/register", registerHandler)
	mux.Handle("/api/user/login", loginHandler)

	// Create an auth middleware
	//authMiddleware := middleware.Auth(authService)

	server := &http.Server{
		Addr:    cfg.ServerAddress,
		Handler: mux,
	}

	go func() {
		log.Printf("Server started on %s", cfg.ServerAddress)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed to start: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shut down the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server gracefully stopped")
}
