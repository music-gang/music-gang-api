package query

import "strings"

func DeleteAuthQuery() string {
	return `
		DELETE FROM auths WHERE id = $1
	`
}

func InsertAuthQuery() string {
	return `
		INSERT INTO auths (
			user_id,
			source,
			source_id,
			access_token,
			refresh_token,
			expiry,
			created_at,
			updated_at
		) VALUES ( $1, $2, $3, $4, $5, $6, $7, $8 ) RETURNING id
	`
}

func SelectAuthsQuery(whereConditions []string, limit, offset int) string {
	return `
		SELECT
			id,
			user_id,
			source,
			source_id,
			access_token,
			refresh_token,
			expiry,
			created_at,
			updated_at,
			COUNT(*) OVER()
		FROM auths
		WHERE ` + strings.Join(whereConditions, " AND ") + `
		ORDER BY id ASC
		` + FormatLimitOffset(limit, offset)
}

func UpdateAuthQuery() string {
	return `
		UPDATE auths SET
			access_token = $1,
			refresh_token = $2,
			expiry = $3,
			updated_at = $4
		WHERE id = $5
	`
}