package postgres

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"io/fs"
	"sort"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/music-gang/music-gang-api/app/apperr"
)

//go:embed migration/*.sql
var migrationsFS embed.FS

// Tx wraps the SQL Tx object to provide a timestamp at the start of the transaction.
type Tx struct {
	*sqlx.Tx
	db  *DB
	now time.Time
}

// DB represents the database connection.
type DB struct {
	conn   *sqlx.DB
	ctx    context.Context
	cancel func()

	DSN string

	Now func() time.Time
}

// NewDB returns a new instance of DB with the given DSN.
func NewDB(dsn string) *DB {
	db := &DB{
		DSN: dsn,
		Now: time.Now,
	}
	db.ctx, db.cancel = context.WithCancel(context.Background())
	return db
}

// createMigrationsTable creates the migrations table if it doesn't exist.
func (db *DB) createMigrationsTable() error {
	if _, err := db.conn.Exec("CREATE TABLE IF NOT EXISTS migrations (name TEXT PRIMARY KEY);"); err != nil {
		return apperr.Errorf(apperr.EINTERNAL, "failed to create migrations table: %s", err)
	}
	return nil
}

// migrate sets up migration tracking and executes pending migration files.
//
// Migration files are embedded in the sqlite/migration folder and are executed
// in lexigraphical order.
//
// Once a migration is run, its name is stored in the 'migrations' table so it
// is not re-executed. Migrations run in a transaction to prevent partial
// migrations
func (db *DB) migrate() error {

	if err := db.createMigrationsTable(); err != nil {
		return err
	}

	names, err := fs.Glob(migrationsFS, "migration/*.sql")
	if err != nil {
		return apperr.Errorf(apperr.EINTERNAL, "failed to glob migrations: %s", err)
	}

	sort.Strings(names)

	tx, err := db.conn.Beginx()
	if err != nil {
		return apperr.Errorf(apperr.EINTERNAL, "failed to begin transaction: %s", err)
	}

	for _, name := range names {
		if err := db.migrateFile(tx, name); err != nil {
			tx.Rollback()
			return err
		}
	}

	return tx.Commit()
}

// migrate runs a single migration file within a transaction. On success, the
// migration file name is saved to the "migrations" table to prevent re-running.
func (db *DB) migrateFile(tx *sqlx.Tx, name string) error {

	var n int
	if err := tx.QueryRow("SELECT COUNT(*) FROM migrations WHERE name = $1", name).Scan(&n); err != nil {
		return apperr.Errorf(apperr.EINTERNAL, "failed to query migrations: %s", err)
	} else if n != 0 {
		return nil
	}

	if buf, err := fs.ReadFile(migrationsFS, name); err != nil {

		return apperr.Errorf(apperr.EINTERNAL, "failed to read migration file: %s", err)

	} else if _, err := tx.Exec(string(buf)); err != nil {

		return apperr.Errorf(apperr.EINTERNAL, "failed exec migration %s: %s", name, err)
	}

	if _, err := tx.Exec("INSERT INTO migrations (name) VALUES ($1)", name); err != nil {
		return apperr.Errorf(apperr.EINTERNAL, "failed insert migration: %s", err)
	}

	return nil
}

// BeginTx starts a transaction and returns a wrapper Tx type. This type
// provides a reference to the database and a fixed timestamp at the start of
// the transaction. The timestamp allows us to mock time during tests as well.
func (db *DB) BeginTx(ctx context.Context, opts *sql.TxOptions) (*Tx, error) {
	tx, err := db.conn.BeginTxx(ctx, opts)
	if err != nil {
		return nil, apperr.Errorf(apperr.EINTERNAL, "failed to begin transaction: %v", err)
	}

	// Return wrapper Tx that includes the transaction start time.
	return &Tx{
		Tx:  tx,
		db:  db,
		now: db.Now().UTC().Truncate(time.Second),
	}, nil
}

// Close closes the database connection.
func (db *DB) Close() error {
	db.cancel()
	if db.conn != nil {
		return db.conn.Close()
	}
	return nil
}

// MustOpen opens the database and panics on error.
func (db *DB) MustOpen() {
	if err := db.Open(); err != nil {
		panic(err)
	}
}

// Open database connection
func (db *DB) Open() error {

	if db.DSN == "" {
		return apperr.Errorf(apperr.EINVALID, "no DSN provided")
	}

	var err error

	if db.conn, err = sqlx.Open("postgres", db.DSN); err != nil {
		return err
	}

	if err := db.migrate(); err != nil {
		return err
	}

	return nil
}

// FormatLimitOffset returns a SQL string for a given limit & offset.
// Clauses are only added if limit and/or offset are greater than zero.
func FormatLimitOffset(limit, offset int) string {
	if limit > 0 && offset > 0 {
		return fmt.Sprintf(`LIMIT %d OFFSET %d`, limit, offset)
	} else if limit > 0 {
		return fmt.Sprintf(`LIMIT %d`, limit)
	} else if offset > 0 {
		return fmt.Sprintf(`OFFSET %d`, offset)
	}
	return ""
}
