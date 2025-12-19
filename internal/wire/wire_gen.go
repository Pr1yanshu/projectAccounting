package wire

import (
	"accounting/internal/db"
	"accounting/internal/handlers"
	"accounting/internal/service"

	"github.com/jackc/pgx/v5/pgxpool"
)

// InitializeServer is the manual, generated equivalent of what wire would produce.
func InitializeServer(pool *pgxpool.Pool) (*handlers.Server, error) {
	dbClient := db.New(pool)
	svc := service.New(dbClient)
	srv := handlers.NewServer(svc)
	return srv, nil
}
