package postgres

import (
	"testing"

	"github.com/music-gang/music-gang-api/config"
)

func TestDB(t *testing.T) {
	db := MustOpenDB(t)
	MustCloseDB(t, db)
}

func MustOpenDB(tb testing.TB) *DB {

	tb.Helper()

	config.LoadConfigWithOptions(config.LoadOptions{ConfigFilePath: "../config.yaml"})

	dsn := config.BuildDSNFromDatabaseConfigForPostgres(config.GetConfig().TEST.Databases.Postgres)

	db := NewDB(dsn)
	if err := db.Open(); err != nil {
		tb.Fatal(err)
	}

	return db
}

func MustCloseDB(tb testing.TB, db *DB) {

	tb.Helper()

	if err := db.Close(); err != nil {
		tb.Fatal(err)
	}
}

func MustExec(tb testing.TB, db *DB, sql string, args ...interface{}) {

	tb.Helper()

	if _, err := db.conn.Exec(sql, args...); err != nil {
		tb.Fatal(err)
	}
}

func MustTruncateTable(tb testing.TB, db *DB, table string) {

	tb.Helper()

	if _, err := db.conn.Exec("TRUNCATE TABLE " + table + " RESTART IDENTITY CASCADE"); err != nil {
		tb.Fatal(err)
	}
}
