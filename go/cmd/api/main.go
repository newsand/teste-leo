package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/users-api-go/internal/database"
	"github.com/users-api-go/internal/handlers"
)

func main() {
	port := flag.String("port", "8080", "Server port")
	dbPath := flag.String("db", "users.db", "SQLite database path")
	flag.Parse()

	if envPort := os.Getenv("PORT"); envPort != "" {
		*port = envPort
	}
	if envDB := os.Getenv("DB_PATH"); envDB != "" {
		*dbPath = envDB
	}

	db, err := database.NewDB(*dbPath)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	userRepo := database.NewUserRepository(db)
	userHandler := handlers.NewUserHandler(userRepo)

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.SetHeader("Content-Type", "application/json"))

	r.Route("/users", func(r chi.Router) {
		r.Post("/", userHandler.Create)
		r.Get("/", userHandler.List)
		r.Get("/{id}", userHandler.GetByID)
		r.Put("/{id}", userHandler.Update)
		r.Patch("/{id}", userHandler.Update)
		r.Delete("/{id}", userHandler.Delete)
	})

	addr := fmt.Sprintf(":%s", *port)
	log.Printf("Starting server on %s", addr)
	if err := http.ListenAndServe(addr, r); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
