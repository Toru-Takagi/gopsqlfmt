package formatter_test

import (
	"github.com/Toru-Takagi/sql_formatter_go/formatter"

	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/assert"
)

func TestFormat(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		sql  string
		want string
	}{
		{
			name: "simple select sql",
			sql:  "SELECT user_name FROM users",
			want: `
SELECT
	user_name
FROM users
`,
		},
		{
			name: "multiple column select sql",
			sql: `
				SELECT   
					user_uuid      ,
						user_name
					, user_age FROM users
			`,
			want: `
SELECT
	user_uuid,
	user_name,
	user_age
FROM users
`,
		},
		{
			name: "column name alias",
			sql:  "select u.user_name un from users u",
			want: `
SELECT
	u.user_name AS un
FROM users u
`,
		},
		{
			name: "alias with as",
			sql:  "select u.user_name as un from users as u",
			want: `
SELECT
	u.user_name AS un
FROM users u
`,
		},
		{
			name: "select sql with where",
			sql: `
				SELECT
					user_uuid
				FROM users
				WHERE user_uuid = $1 
			`,
			want: `
SELECT
	user_uuid
FROM users
WHERE user_uuid = $1
`,
		},
		{
			name: "select sql with where and multiple columns",
			sql: `
				SELECT user_uuid, user_name
				FROM users WHERE user_uuid = $1 AND user_email = $2
			`,
			want: `
SELECT
	user_uuid,
	user_name
FROM users
WHERE user_uuid = $1
	AND user_email = $2
`,
		},
		{
			name: "select sql with where or multiple columns",
			sql: `
							SELECT user_uuid, user_name
				FROM users WHERE user_uuid = $1 OR user_email = $2
			`,
			want: `
SELECT
	user_uuid,
	user_name
FROM users
WHERE user_uuid = $1
	OR user_email = $2
`,
		},
		{
			name: "select sql with where and, or multiple columns",
			sql: `
							SELECT user_uuid, user_name
				FROM users WHERE user_uuid = $1 AND (user_email = $2 OR user_age = $3)
			`,
			want: `
SELECT
	user_uuid,
	user_name
FROM users
WHERE user_uuid = $1
	AND (
		user_email = $2
			OR user_age = $3
	)
`,
		},
		{
			name: "not equal and boolean",
			sql:  `SELECT user_name FROM users WHERE name != 'taro' AND is_active = true`,
			want: `
SELECT
	user_name
FROM users
WHERE name != 'taro'
	AND is_active = true
`,
		},
		{
			name: "select with named parameter",
			sql: `
				select user_uuid, user_name from users where user_uuid = :user_uuid and user_email = :user_email
			`,
			want: `
SELECT
	user_uuid,
	user_name
FROM users
WHERE user_uuid = :user_uuid
	AND user_email = :user_email
`,
		},
		{
			name: "select with alias table",
			sql:  `select u.user_uuid, u.user_name from users u`,
			want: `
SELECT
	u.user_uuid,
	u.user_name
FROM users u
`,
		},
		{
			name: "limit",
			sql:  `select user_uuid from users limit 10`,
			want: `
SELECT
	user_uuid
FROM users
LIMIT 10
`,
		},
		{
			name: "order by",
			sql:  `select user_uuid from users u order by u.user_uuid desc, u.user_name asc`,
			want: `
SELECT
	user_uuid
FROM users u
ORDER BY u.user_uuid DESC,
	u.user_name ASC
`,
		},
		{
			name: "select with subquery",
			sql:  `select u.user_uuid, (select ull.last_login_at from user_last_login ull where ull.user_uuid = u.user_uuid) as last_login_at from users u`,
			want: `
SELECT
	u.user_uuid,
	(
		SELECT
			ull.last_login_at
		FROM user_last_login ull
		WHERE ull.user_uuid = u.user_uuid
	) AS last_login_at
FROM users u
`,
		},
		{
			name: "select with inner join",
			sql:  `select u.user_name, ull.last_login_at from users u inner join user_last_login ull on u.user_uuid = ull.user_uuid where u.user_uuid = $1`,
			want: `
SELECT
	u.user_name,
	ull.last_login_at
FROM users u
INNER JOIN user_last_login ull ON u.user_uuid = ull.user_uuid
WHERE u.user_uuid = $1
`,
		},
		{
			name: "three inner join",
			sql: `select u.user_name, ull.last_login_at, uage.user_age, uadd.address from users u
						inner join user_last_login ull on u.user_uuid = ull.user_uuid
						inner join user_age uage on u.user_uuid = uage.user_uuid
						inner join user_address uadd on u.user_uuid = uadd.user_uuid
						where u.user_uuid = $1`,
			want: `
SELECT
	u.user_name,
	ull.last_login_at,
	uage.user_age,
	uadd.address
FROM users u
INNER JOIN user_last_login ull ON u.user_uuid = ull.user_uuid
INNER JOIN user_age uage ON u.user_uuid = uage.user_uuid
INNER JOIN user_address uadd ON u.user_uuid = uadd.user_uuid
WHERE u.user_uuid = $1
`,
		},
		{
			name: "inner join and left join",
			sql: `select u.user_name, ull.last_login_at, uage.user_age, uadd.address from users u
						inner join user_last_login ull on u.user_uuid = ull.user_uuid
						left join user_age uage on u.user_uuid = uage.user_uuid
						left join user_address uadd on u.user_uuid = uadd.user_uuid
						where u.user_uuid = $1`,
			want: `
SELECT
	u.user_name,
	ull.last_login_at,
	uage.user_age,
	uadd.address
FROM users u
INNER JOIN user_last_login ull ON u.user_uuid = ull.user_uuid
LEFT JOIN user_age uage ON u.user_uuid = uage.user_uuid
LEFT JOIN user_address uadd ON u.user_uuid = uadd.user_uuid
WHERE u.user_uuid = $1
`,
		},
		{
			name: "join on and",
			sql:  `SELECT u.user_name, ull.last_login_at FROM users u INNER JOIN user_last_login ull ON u.user_uuid = ull.user_uuid and u.email = ull.email`,
			want: `
SELECT
	u.user_name,
	ull.last_login_at
FROM users u
INNER JOIN user_last_login ull ON u.user_uuid = ull.user_uuid
	AND u.email = ull.email
`,
		},
		{
			name: "window function of count",
			sql:  `select COUNT(*) OVER () AS total, user_uuid from users`,
			want: `
SELECT
	COUNT(*) OVER() AS total,
	user_uuid
FROM users
`,
		},
		{
			name: "simple insert",
			sql: `
				insert into users (user_uuid, user_name, user_age, created_at) values ($1, $2, $3, now())
			`,
			want: `
INSERT INTO users(
	user_uuid,
	user_name,
	user_age,
	created_at
) VALUES (
	$1,
	$2,
	$3,
	NOW()
)
`,
		},
		{
			name: "insert from select",
			sql: `
				insert into deleted_users (user_uuid, user_name, user_age) select user_uuid, user_name, user_age from users where user_uuid = $1
			`,
			want: `
INSERT INTO deleted_users(
	user_uuid,
	user_name,
	user_age
) SELECT
	user_uuid,
	user_name,
	user_age
FROM users
WHERE user_uuid = $1
`,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			actual, err := formatter.Format(tt.sql)
			assert.NoError(t, err)
			t.Log(actual)
			if diff := cmp.Diff(tt.want, actual); diff != "" {
				t.Errorf("diff: %s", diff)
			}
		})
	}
}
