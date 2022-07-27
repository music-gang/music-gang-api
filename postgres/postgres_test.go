package postgres_test

import (
	"testing"

	"github.com/music-gang/music-gang-api/config"
	"github.com/music-gang/music-gang-api/postgres"
)

func TestDB(t *testing.T) {
	db := MustOpenDB(t)
	MustCloseDB(t, db)
}

func MustOpenDB(tb testing.TB) *postgres.DB {

	tb.Helper()

	dsn := config.BuildDSNFromDatabaseConfigForPostgres(config.GetConfig().APP.Databases.Postgres)

	db := postgres.NewDB(dsn)
	if err := db.Open(); err != nil {
		tb.Fatal(err)
	}

	return db
}

func MustCloseDB(tb testing.TB, db *postgres.DB) {

	tb.Helper()

	if err := db.Close(); err != nil {
		tb.Fatal(err)
	}
}

func MustExec(tb testing.TB, db *postgres.DB, sql string, args ...interface{}) {

	tb.Helper()

	if _, err := postgres.GetConn(db).Exec(sql, args...); err != nil {
		tb.Fatal(err)
	}
}

func MustTruncateTable(tb testing.TB, db *postgres.DB, table string) {

	tb.Helper()

	if _, err := postgres.GetConn(db).Exec("TRUNCATE TABLE " + table + " RESTART IDENTITY CASCADE"); err != nil {
		tb.Fatal(err)
	}
}
