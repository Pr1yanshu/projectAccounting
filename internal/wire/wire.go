//go:build wire
// +build wire

package wire

import (
	"github.com/google/wire"

	"accounting/internal/db"
	"accounting/internal/handlers"
	"accounting/internal/service"

	"github.com/jackc/pgx/v5/pgxpool"
)

// ProviderSet for the application wiring.
var ProviderSet = wire.NewSet(
	db.New,
	wire.Bind(new(db.Store), new(*db.DB)),
	service.New,
	handlers.NewServer,
)

func InitializeServer(pool *pgxpool.Pool) (*handlers.Server, error) {
	wire.Build(ProviderSet)
	return &handlers.Server{}, nil
}
