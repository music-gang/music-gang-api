package query

import "strings"

func DeleteUserQuery() string {
	return `DELETE FROM users WHERE id = $1`
}

func InsertUserQuery() string {
	return `
		INSERT INTO users (
			name,
			email,
			password,
			created_at,
			updated_at
		) VALUES ( $1, $2, $3, $4, $5 ) RETURNING id
	`
}

func SelectUsersQuery(whereCondtions []string, limit, offset int) string {
	return `
		SELECT 
		    id,
		    name,
		    email,
			password,
		    created_at,
		    updated_at,
		    COUNT(*) OVER() as count
		FROM users
		WHERE ` + strings.Join(whereCondtions, " AND ") + `
		ORDER BY id ASC
		` + FormatLimitOffset(limit, offset)
}

func UpdateUserQuery() string {
	return `
		UPDATE users SET
			name = $1,
			updated_at = $2
		WHERE id = $3
	`
}
