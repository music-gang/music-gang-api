package query

import "strings"

func DeleteContractQuery() string {
	return `
		DELETE FROM contracts WHERE id = $1
	`
}

func InsertContractQuery() string {
	return `
		INSERT INTO contracts (
			name,
			description,
			user_id,
			visibility,
			max_fuel,
			stateful,
			created_at,
			updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id;
	`
}

func InsertRevisionQuery() string {
	return `
		INSERT INTO revisions (
			rev,
			version,
			contract_id,
			notes,
			compiled_code,
			max_fuel,
			created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id
	`
}

func SelectContractsQuery(whereCondtions []string, limit, offset int) string {
	return `		
		SELECT
			id,
			name,
			description,
			user_id,
			visibility,
			max_fuel,
			stateful,
			created_at,
			updated_at,
			COUNT(*) OVER() as count
		FROM contracts
		WHERE ` + strings.Join(whereCondtions, " AND ") + `
		ORDER BY id ASC
		` + FormatLimitOffset(limit, offset)
}

func SelectRevisionsQuery(whereCondtions []string, limit, offset int) string {
	return `
		SELECT
			id,
			rev,
			version,
			contract_id,
			notes,
			compiled_code,
			max_fuel,
			created_at,
			COUNT(*) OVER() as count
		FROM revisions
		WHERE ` + strings.Join(whereCondtions, " AND ") + `
		ORDER BY rev DESC
		` + FormatLimitOffset(limit, offset)
}

func UpdateContractQuery() string {
	return 	`
		UPDATE contracts SET
			name = $1,
			description = $2,
			max_fuel = $3,
			updated_at = $4
		WHERE id = $5
	`
}