package sqlconst

func main() {
	const sql = `
		SELECT
			user_uuid
		FROM users
	`

	const sql2 = "SELECT user_uuid FROM users2"

	const sql3, sql4 = `
		SELECT 
			user_uuid
		FROM users3
	`, `
		SELECT
			user_uuid
		FROM users4
	`

	// q := `
	// 	INSERT INTO users (
	// 		user_uuid
	// 	) VALUES (
	// 		$1
	// 	)
	// `

	// var query = `
	// 	INSERT INTO users(
	// 		user_uuid
	// 	) VALUES (
	// 		$1
	// 	)
	// `
}
