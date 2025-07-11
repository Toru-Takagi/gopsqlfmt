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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := formatter.Format(tt.sql, nil)
			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}