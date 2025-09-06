package formatter_test

import (
	"testing"

	"github.com/Toru-Takagi/gopsqlfmt/formatter"
	"github.com/stretchr/testify/assert"
)

func TestFormatAConst(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		sql  string
		want string
	}{
		{
			name: "integer constant",
			sql:  "SELECT 42 FROM users",
			want: `
SELECT
  42
FROM users
`,
		},
		{
			name: "float constant",
			sql:  "SELECT 3.14 FROM users",
			want: `
SELECT
  3.14
FROM users
`,
		},
		{
			name: "string constant",
			sql:  "SELECT 'hello' FROM users",
			want: `
SELECT
  'hello'
FROM users
`,
		},
		{
			name: "boolean constant true",
			sql:  "SELECT true FROM users",
			want: `
SELECT
  true
FROM users
`,
		},
		{
			name: "boolean constant false",
			sql:  "SELECT false FROM users",
			want: `
SELECT
  false
FROM users
`,
		},
		{
			name: "null constant",
			sql:  "SELECT NULL FROM users",
			want: `
SELECT
  NULL
FROM users
`,
		},
		{
			name: "complex query with subquery and aliases",
			sql: `SELECT
    g.gather_uuid,
    COALESCE((
        SELECT
            COUNT(DISTINCT name)
        FROM
            (
                SELECT
                    ga.attendance_name as name
                FROM gather_attendance ga
                WHERE g.gather_uuid = ga.gather_uuid
                UNION ALL
                SELECT
                    gp.participant_name as name
                FROM gather_participant gp
                WHERE g.gather_uuid = gp.gather_uuid
            ) AS combined_names
    ), 0) AS number_of_participants
FROM gather g
WHERE g.gather_uuid = ANY($1)
    AND g.deleted_at IS NULL
ORDER BY COALESCE(g.confirmed_start_date_time, g.adjustment_start_date_time) DESC`,
			want: `
SELECT
  g.gather_uuid,
  COALESCE((
    SELECT
      count(DISTINCT name)
    FROM (
      SELECT
        ga.attendance_name AS name
      FROM gather_attendance ga
      WHERE g.gather_uuid = ga.gather_uuid
      UNION ALL
      SELECT
        gp.participant_name AS name
      FROM gather_participant gp
      WHERE g.gather_uuid = gp.gather_uuid
    ) combined_names
  ), 0) AS number_of_participants
FROM gather g
WHERE g.gather_uuid = ANY($1)
  AND g.deleted_at IS NULL
ORDER BY COALESCE(g.confirmed_start_date_time, g.adjustment_start_date_time) DESC
`,
		},
		{
			name: "INSERT with CURRENT_TIMESTAMP",
			sql:  "INSERT INTO hoge (created_at, updated_at) VALUES (CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)",
			want: `
INSERT INTO hoge(
  created_at,
  updated_at
) VALUES (
  CURRENT_TIMESTAMP,
  CURRENT_TIMESTAMP
)
`,
		},
		{
			name: "UPDATE with CURRENT_TIMESTAMP",
			sql:  "UPDATE users SET updated_at = CURRENT_TIMESTAMP WHERE id = $1",
			want: `
UPDATE users
SET
  updated_at = CURRENT_TIMESTAMP
WHERE id = $1
`,
		},
		{
			name: "INSERT with ON CONFLICT and parameters",
			sql:  "INSERT INTO todo (title, created_at, updated_at) VALUES ($1, now(), now()) ON CONFLICT (todo_uuid) DO UPDATE SET title = $2, updated_at = now()",
			want: `
INSERT INTO todo(
  title,
  created_at,
  updated_at
) VALUES (
  $1,
  now(),
  now()
)
ON CONFLICT(todo_uuid)
DO UPDATE SET
  title = $2,
  updated_at = now()
`,
		},
		{
			name: "SELECT with LIMIT parameter",
			sql:  "SELECT * FROM user LIMIT $1",
			want: `
SELECT
  *
FROM user
LIMIT $1
`,
		},
		{
			name: "SELECT with WHERE COALESCE",
			sql:  "SELECT user_uuid, name FROM user WHERE user_uuid = $1 AND COALESCE(name, '') = COALESCE($2, '')",
			want: `
SELECT
  user_uuid,
  name
FROM user
WHERE user_uuid = $1
  AND COALESCE(name, '') = COALESCE($2, '')
`,
		},
		{
			name: "SELECT with HAVING COUNT DISTINCT",
			sql:  "SELECT user_id, COUNT(*) FROM users GROUP BY user_id HAVING COUNT(DISTINCT name) < 5",
			want: `
SELECT
  user_id,
  count(*)
FROM users
GROUP BY user_id
HAVING count(DISTINCT name) < 5
`,
		},
		{
			name: "SELECT with LEFT JOIN subquery DISTINCT",
			sql:  "SELECT * FROM users LEFT JOIN (SELECT DISTINCT department FROM employees) dept ON users.dept_id = dept.department",
			want: `
SELECT
  *
FROM users
  LEFT JOIN (
    SELECT DISTINCT
      department
    FROM employees
  ) dept
    ON users.dept_id = dept.department
`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := formatter.Format(tt.sql, nil)
			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}
