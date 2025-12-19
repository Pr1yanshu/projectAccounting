package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"accounting/internal/wire"

	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		log.Fatal("DATABASE_URL environment variable is required")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		log.Fatalf("unable to connect to database: %v", err)
	}
	defer pool.Close()

	r := mux.NewRouter()
	srv, err := wire.InitializeServer(pool)
	if err != nil {
		log.Fatalf("failed to initialize server: %v", err)
	}

	r.HandleFunc("/accounts", srv.CreateAccount).Methods("POST")
	r.HandleFunc("/accounts/{id}", srv.GetAccount).Methods("GET")
	r.HandleFunc("/transactions", srv.CreateTransaction).Methods("POST")
	// created additional routes for testing purposes
	r.HandleFunc("/accounts", srv.GetAllAccounts).Methods("GET")
	r.HandleFunc("/accounts/{id}/transactions", srv.GetTransactions).Methods("GET")

	addr := ":8080"
	log.Printf("listening on %s", addr)
	if err := http.ListenAndServe(addr, r); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
