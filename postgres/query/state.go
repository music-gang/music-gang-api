package query

func InsertStateQuery() string {
	return `
		INSERT INTO states (
			revision_id,
			value,
			user_id,
			created_at,
			updated_at
		) VALUES ($1, $2, $3, $4, $5) RETURNING id
	`
}

func SelectStateByRevisionIDAndUserIDQuery() string {
	return `
		SELECT
			id,
			revision_id,
			value,
			user_id,
			created_at,
			updated_at
		FROM states
		WHERE revision_id = $1 AND user_id = $2
	`
}

func UpdateStateQuery() string {
	return `
		UPDATE states SET
			value = $1,
			updated_at = $2
		WHERE id = $3
	`
}