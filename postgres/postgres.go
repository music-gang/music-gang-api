package postgres

import (
	"context"
	"embed"
	"time"

	"github.com/jmoiron/sqlx"
)

// go:embed migration/*.sql
var migrationsFS embed.FS

type DB struct {
	db     *sqlx.DB
	ctx    context.Context
	cancel func()

	DSN string

	Now func() time.Time
}
