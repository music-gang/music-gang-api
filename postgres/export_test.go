package postgres

import (
	"context"

	"github.com/jmoiron/sqlx"
	"github.com/music-gang/music-gang-api/app/entity"
)

func GetConn(db *DB) *sqlx.DB {
	return db.conn
}

var AttachUserAssociations = attachUserAssociations
var AttachContractAssociations = attachContractAssociations

var FindUserByEmail = func(ctx context.Context, db *DB, email string) (*entity.User, error) {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	return findUserByEmail(ctx, tx, email)
}
