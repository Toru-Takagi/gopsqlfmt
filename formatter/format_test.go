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
			name: "order by with function",
			sql:  `select user_uuid from users u order by min(u.registered_at)`,
			want: `
SELECT
	user_uuid
FROM users u
ORDER BY MIN(u.registered_at)
`,
		},
		{
			name: "group by",
			sql:  `select count(*) from users u group by u.name, u.age`,
			want: `
SELECT
	COUNT(*)
FROM users u
GROUP BY u.name, u.age
`,
		},
		{
			name: "FOR UPDATE SKIP LOCKED",
			sql:  `select user_uuid from users for update skip locked`,
			want: `
SELECT
	user_uuid
FROM users
FOR UPDATE SKIP LOCKED
`,
		},
		{
			name: "current_setting",
			sql:  `select current_setting('search_path') as search_path`,
			want: `
SELECT
	CURRENT_SETTING('search_path') AS search_path
`,
		},
		{
			name: "set_config",
			sql:  `select set_config('test', $1, false)`,
			want: `
SELECT
	SET_CONFIG('test', $1, false)
`,
		},
		{
			name: "array_agg",
			sql:  `select array_agg(t.tablename ORDER BY t.tablename) from pg_catalog.pg_tables AS t`,
			want: `
SELECT
	ARRAY_AGG(t.tablename ORDER BY t.tablename)
FROM pg_catalog.pg_tables t
`,
		},
		{
			name: "reserved word: user",
			sql:  "SELECT u.user_name FROM user u ",
			want: `
SELECT
	u.user_name
FROM user u
`,
		},
		{
			name: "select with subquery",
			sql:  `select u.user_uuid, (select ull.last_login_at, current_setting('test') from user_last_login ull where ull.user_uuid = u.user_uuid) as last_login_at from users u`,
			want: `
SELECT
	u.user_uuid,
	(
		SELECT
			ull.last_login_at,
			CURRENT_SETTING('test')
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
			name: "insert: named parameter",
			sql:  `insert into users (user_name, user_age, created_at) values (:user_name, :user_age, now())`,
			want: `
INSERT INTO users(
	user_name,
	user_age,
	created_at
) VALUES (
	:user_name,
	:user_age,
	NOW()
)
`,
		},
		{
			name: "insert: primitive string",
			sql:  `insert into users (user_name, user_age, created_at) values ('taro', 20, now())`,
			want: `
INSERT INTO users(
	user_name,
	user_age,
	created_at
) VALUES (
	'taro',
	20,
	NOW()
)
`,
		},
		{
			name: "insert from select",
			sql: `
				insert into deleted_users (user_uuid, user_name, user_age, registered_at) select user_uuid, user_name, user_age, now() from users where user_uuid = $1
			`,
			want: `
INSERT INTO deleted_users(
	user_uuid,
	user_name,
	user_age,
	registered_at
) SELECT
	user_uuid,
	user_name,
	user_age,
	NOW()
FROM users
WHERE user_uuid = $1
`,
		},
		{
			name: "insert: gen_random_uuid",
			sql:  `insert into users (user_uuid, user_name) values (gen_random_uuid(), $1)`,
			want: `
INSERT INTO users(
	user_uuid,
	user_name
) VALUES (
	GEN_RANDOM_UUID(),
	$1
)
`,
		},
		{
			name: "insert: current_setting",
			sql:  `insert into users (tenant_name_id) values (current_setting('tenant_name_id'))`,
			want: `
INSERT INTO users(
	tenant_name_id
) VALUES (
	CURRENT_SETTING('tenant_name_id')
)
`,
		},
		{
			name: "insert: on conflict do update",
			sql: `
				insert into users (user_uuid, user_name, user_age) values ($1, $2, $3)
				on conflict (user_uuid) do update set user_name = EXCLUDED.user_name, user_age = EXCLUDED.user_age, updated_at = now()
			`,
			want: `
INSERT INTO users(
	user_uuid,
	user_name,
	user_age
) VALUES (
	$1,
	$2,
	$3
)
ON CONFLICT(user_uuid)
DO UPDATE SET
	user_name = EXCLUDED.user_name,
	user_age = EXCLUDED.user_age,
	updated_at = NOW()
`,
		},
		{
			name: "insert: on conflict do nothing",
			sql: `
				insert into users (user_uuid, user_name, user_age) values ($1, $2, $3)
				on conflict on constraint user_unique_key do nothing
			`,
			want: `
INSERT INTO users(
	user_uuid,
	user_name,
	user_age
) VALUES (
	$1,
	$2,
	$3
)
ON CONFLICT ON CONSTRAINT user_unique_key
DO NOTHING
`,
		},
		{
			name: "simple update",
			sql:  `update users set user_name = $1, user_age = $2, updated_at = now() where user_uuid = $3`,
			want: `
UPDATE users
SET
	user_name = $1,
	user_age = $2,
	updated_at = NOW()
WHERE user_uuid = $3
`,
		},
		{
			name: "simple delete",
			sql:  `delete from users where user_uuid = $1`,
			want: `
DELETE FROM users
WHERE user_uuid = $1
`,
		},
		{
			name: "delete: current_setting",
			sql:  `delete from users where locale = current_setting('locale')`,
			want: `
DELETE FROM users
WHERE locale = CURRENT_SETTING('locale')
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
