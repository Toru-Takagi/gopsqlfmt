package formatter_test

import (
	"github.com/Toru-Takagi/gopsqlfmt/fmtconf"
	"github.com/Toru-Takagi/gopsqlfmt/formatter"

	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/assert"
)

func TestFormat(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		sql  string
		conf *fmtconf.Config
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
			name: "JOIN_USING",
			sql:  `SELECT orders.order_id, customers.customer_name, orders.total_amount FROM orders JOIN customers USING (customer_id)`,
			want: `
SELECT
  orders.order_id,
  customers.customer_name,
  orders.total_amount
FROM orders
  INNER JOIN customers USING(customer_id)
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
			name: "exist",
			sql:  `select user_uuid from users u where exists(select * from today_login_user tlu where u.user_uuid = tlu.user_uuid)`,
			want: `
SELECT
  user_uuid
FROM users u
WHERE EXISTS(
  SELECT
    *
  FROM today_login_user tlu
  WHERE u.user_uuid = tlu.user_uuid
)
`,
		},
		{
			name: "not exist",
			sql:  `select user_uuid from users u where not exists(select * from today_login_user tlu where u.user_uuid = tlu.user_uuid)`,
			want: `
SELECT
  user_uuid
FROM users u
WHERE NOT EXISTS(
  SELECT
    *
  FROM today_login_user tlu
  WHERE u.user_uuid = tlu.user_uuid
)
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
ORDER BY min(u.registered_at)
`,
		},
		{
			name: "group by",
			sql:  `select count(*) from users u group by u.name, u.age`,
			want: `
SELECT
  count(*)
FROM users u
GROUP BY u.name, u.age
`,
		},
		{
			name: "SELECT_COUNT_DISTINCT",
			sql:  `select count(distinct user_name) from users`,
			want: `
SELECT
  count(DISTINCT user_name)
FROM users
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
  current_setting('search_path') AS search_path
`,
		},
		{
			name: "set_config",
			sql:  `select set_config('test', $1, false)`,
			want: `
SELECT
  set_config('test', $1, false)
`,
		},
		{
			name: "array_agg",
			sql:  `select array_agg(t.tablename ORDER BY t.tablename) from pg_catalog.pg_tables AS t`,
			want: `
SELECT
  array_agg(t.tablename ORDER BY t.tablename)
FROM pg_catalog.pg_tables t
`,
		},
		{
			name: "COALESCE",
			sql:  `SELECT name, salary, coalesce(bonus, 0) AS bonus FROM employees`,
			want: `
SELECT
  name,
  salary,
  COALESCE(bonus, 0) AS bonus
FROM employees
`,
		},
		{
			name: "COALESCE_JSON_AGG",
			sql: `select coalesce((
            select json_agg(json_build_object('userUUID', gu.user_uuid, 'userName', gu.user_name)) as results from gest_users gu
          ), '[]') as results from users u`,
			want: `
SELECT
  COALESCE((
    SELECT
      json_agg(json_build_object('userUUID', gu.user_uuid, 'userName', gu.user_name)) AS results
    FROM gest_users gu
  ), '[]') AS results
FROM users u
`,
		},
		{
			name: "COALESCE_ARRAY_LENGTH",
			sql:  `select coalesce(array_length(u.user_uuids, 1),0) as result_count from users u`,
			want: `
SELECT
  COALESCE(array_length(u.user_uuids, 1), 0) AS result_count
FROM users u
`},
		{
			name: "ARRAY",
			sql:  `SELECT ARRAY(select user_uuid from users u WHERE u.user_uuid = $1) as user_uuids FROM login_users`,
			want: `
SELECT
  ARRAY(
    SELECT
      user_uuid
    FROM users u
    WHERE u.user_uuid = $1
  ) AS user_uuids
FROM login_users
`,
		},
		{
			name: "ARRAY_WITH_TYPECAST",
			sql:  `SELECT ARRAY(select ged.day::text from gather_enable_day ged where ged.gather_uuid = g.gather_uuid) as enable_days FROM gather g`,
			want: `
SELECT
  ARRAY(
    SELECT
      ged.day::text
    FROM gather_enable_day ged
    WHERE ged.gather_uuid = g.gather_uuid
  ) AS enable_days
FROM gather g
`,
		},
		{
			name: "FUNC_NAME_TYPE_CASE_UPPER",
			sql:  `select array_agg(user_uuid), now(), gen_random_uuid() from users`,
			conf: fmtconf.NewDefaultConfig().WithFuncNameTypeCaseUpper(),
			want: `
SELECT
  ARRAY_AGG(user_uuid),
  NOW(),
  GEN_RANDOM_UUID()
FROM users
`,
		},
		{
			name: "CARDINALITY",
			sql:  `select cardinality(u.user_uuids) as user_count from users u`,
			want: `
SELECT
  cardinality(u.user_uuids) AS user_count
FROM users u
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
			name: "ANY_string param",
			sql:  `select user_name from users where user_uuid = any('{72c6b8e0-c2f1-4fdd-835a-253fe92cbbd6}'::uuid[])`,
			want: `
SELECT
  user_name
FROM users
WHERE user_uuid = ANY('{72c6b8e0-c2f1-4fdd-835a-253fe92cbbd6}'::uuid[])
`,
		},
		{
			name: "ANY_named paramter",
			sql:  `select user_name from users where user_uuid = ANY(:user_uuids::uuid[])`,
			want: `
SELECT
  user_name
FROM users
WHERE user_uuid = ANY(:user_uuids::uuid[])
`,
		},
		{
			name: "ANY_positional_parameter",
			sql:  `select user_name from users where user_uuid = ANY($1)`,
			want: `
SELECT
  user_name
FROM users
WHERE user_uuid = ANY($1)
`,
		},
		{
			name: "select with subquery",
			sql:  `select u.user_uuid, (select ull.last_login_at, current_setting('test') from user_last_login ull where ull.user_uuid = u.user_uuid and u.email = :email) as last_login_at from users u`,
			want: `
SELECT
  u.user_uuid,
  (
    SELECT
      ull.last_login_at,
      current_setting('test')
    FROM user_last_login ull
    WHERE ull.user_uuid = u.user_uuid
      AND u.email = :email
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
  INNER JOIN user_last_login ull
    ON u.user_uuid = ull.user_uuid
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
  INNER JOIN user_last_login ull
    ON u.user_uuid = ull.user_uuid
  INNER JOIN user_age uage
    ON u.user_uuid = uage.user_uuid
  INNER JOIN user_address uadd
    ON u.user_uuid = uadd.user_uuid
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
  INNER JOIN user_last_login ull
    ON u.user_uuid = ull.user_uuid
  LEFT JOIN user_age uage
    ON u.user_uuid = uage.user_uuid
  LEFT JOIN user_address uadd
    ON u.user_uuid = uadd.user_uuid
WHERE u.user_uuid = $1
`,
		},
		{
			name: "JOIN_LINE_BREAK_OFF",
			sql: `select u.user_name, ull.last_login_at, uage.user_age, uadd.address from users u
						inner join user_last_login ull on u.user_uuid = ull.user_uuid
						left join user_age uage on u.user_uuid = uage.user_uuid
						left join user_address uadd on u.user_uuid = uadd.user_uuid
						where u.user_uuid = $1`,
			conf: fmtconf.NewDefaultConfig().WithJoinLineBreakOff(),
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
			name: "JOIN_START_INDENT_TYPE_NONE",
			sql: `select u.user_name, ull.last_login_at, uage.user_age, uadd.address from users u
						inner join user_last_login ull on u.user_uuid = ull.user_uuid
						left join user_age uage on u.user_uuid = uage.user_uuid
						left join user_address uadd on u.user_uuid = uadd.user_uuid
						where u.user_uuid = $1`,
			conf: fmtconf.NewDefaultConfig().WithJoinStartIndentTypeNone(),
			want: `
SELECT
  u.user_name,
  ull.last_login_at,
  uage.user_age,
  uadd.address
FROM users u
INNER JOIN user_last_login ull
  ON u.user_uuid = ull.user_uuid
LEFT JOIN user_age uage
  ON u.user_uuid = uage.user_uuid
LEFT JOIN user_address uadd
  ON u.user_uuid = uadd.user_uuid
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
  INNER JOIN user_last_login ull
    ON u.user_uuid = ull.user_uuid
  AND u.email = ull.email
`,
		},
		{
			name: "window function of count",
			sql:  `select COUNT(*) OVER () AS total, user_uuid from users`,
			want: `
SELECT
  count(*) OVER() AS total,
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
  now()
)
`,
		},
		{
			name: "INSERT_NO_COLUMNS",
			sql:  `insert into users values ($1, $2)`,
			want: `
INSERT INTO users
VALUES (
  $1,
  $2
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
  now()
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
  now()
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
  now()
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
  gen_random_uuid(),
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
  current_setting('tenant_name_id')
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
  updated_at = now()
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
			name: "insert: on conflict multiple columns do nothing",
			sql: `
				insert into users (user_uuid, google_account_email, user_name) values ($1, $2, $3)
				on conflict (user_uuid, google_account_email) do nothing
			`,
			want: `
INSERT INTO users(
  user_uuid,
  google_account_email,
  user_name
) VALUES (
  $1,
  $2,
  $3
)
ON CONFLICT(user_uuid, google_account_email)
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
  updated_at = now()
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
WHERE locale = current_setting('locale')
`,
		},
		{
			name: "WHERE with IS NULL",
			sql:  `SELECT tweet_uuid, body FROM tweet WHERE operated_by = $1 AND deleted_at IS NULL ORDER BY created_at DESC`,
			want: `
SELECT
  tweet_uuid,
  body
FROM tweet
WHERE operated_by = $1
  AND deleted_at IS NULL
ORDER BY created_at DESC
`,
		},
		{
			name: "WHERE with IS NOT NULL",
			sql:  `SELECT user_uuid FROM users WHERE email IS NOT NULL`,
			want: `
SELECT
  user_uuid
FROM users
WHERE email IS NOT NULL
`,
		},
		{
			name: "INDENT_TYPE_TAB",
			sql:  `select user_uuid from users`,
			conf: fmtconf.NewDefaultConfig().WithIndentTypeTab(),
			want: `
SELECT
	user_uuid
FROM users
`,
		},
		{
			name: "UNION_ALL",
			sql:  `SELECT name FROM (SELECT ga.attendance_name as name FROM gather_attendance ga UNION ALL SELECT gp.participant_name as name FROM gather_participant gp) AS combined_names`,
			want: `
SELECT
  name
FROM (
  SELECT
    ga.attendance_name AS name
  FROM gather_attendance ga
  UNION ALL
  SELECT
    gp.participant_name AS name
  FROM gather_participant gp
) combined_names
`,
		},
		{
			name: "UNION",
			sql:  `SELECT user_name FROM users UNION SELECT guest_name FROM guests`,
			want: `
SELECT
  user_name
FROM users
UNION
SELECT
  guest_name
FROM guests
`,
		},
		{
			name: "COMPLEX_UNION_ALL_SUBQUERY",
			sql:  `SELECT g.gather_uuid, COALESCE((SELECT COUNT(DISTINCT name) FROM (SELECT ga.attendance_name as name FROM gather_attendance ga WHERE g.gather_uuid = ga.gather_uuid UNION ALL SELECT gp.participant_name as name FROM gather_participant gp WHERE g.gather_uuid = gp.gather_uuid) AS combined_names), 0) AS number_of_participants FROM gather g WHERE g.gather_uuid = ANY($1) AND g.deleted_at IS NULL`,
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
`,
		},
		{
			name: "CASE_SIMPLE",
			sql:  `SELECT name, CASE WHEN age >= 18 THEN 'adult' ELSE 'minor' END AS age_category FROM users`,
			want: `
SELECT
  name,
  CASE WHEN age >= 18 THEN 'adult' ELSE 'minor' END AS age_category
FROM users
`,
		},
		{
			name: "CASE_WITH_IS_NOT_NULL",
			sql:  `SELECT todo_setting_uuid, label, day_of_week, CASE WHEN setting_uuid IS NOT NULL THEN 'COMPLETE' ELSE 'INCOMPLETE' END AS today_status FROM todo_settings`,
			want: `
SELECT
  todo_setting_uuid,
  label,
  day_of_week,
  CASE WHEN setting_uuid IS NOT NULL THEN 'COMPLETE' ELSE 'INCOMPLETE' END AS today_status
FROM todo_settings
`,
		},
		{
			name: "CASE_WITH_IS_NULL",
			sql:  `SELECT user_uuid, CASE WHEN email IS NULL THEN 'no email' ELSE email END AS email_display FROM users`,
			want: `
SELECT
  user_uuid,
  CASE WHEN email IS NULL THEN 'no email' ELSE email END AS email_display
FROM users
`,
		},
		{
			name: "CASE_MULTIPLE_WHEN",
			sql:  `SELECT name, CASE WHEN age < 13 THEN 'child' WHEN age < 18 THEN 'teen' ELSE 'adult' END AS age_group FROM users`,
			want: `
SELECT
  name,
  CASE WHEN age < 13 THEN 'child' WHEN age < 18 THEN 'teen' ELSE 'adult' END AS age_group
FROM users
`,
		},
		{
			name: "CASE_WITH_COLUMN_RESULT",
			sql:  `SELECT user_uuid, CASE WHEN preferred_name IS NOT NULL THEN preferred_name ELSE full_name END AS display_name FROM users`,
			want: `
SELECT
  user_uuid,
  CASE WHEN preferred_name IS NOT NULL THEN preferred_name ELSE full_name END AS display_name
FROM users
`,
		},
		{
			name: "CASE_COMPLEX_QUERY_WITH_JOINS",
			sql:  `SELECT ts.todo_setting_uuid, ts.label, ts.day_of_week, CASE WHEN t.setting_uuid IS NOT NULL THEN 'COMPLETE' ELSE 'INCOMPLETE' END AS today_status FROM todo_settings ts LEFT JOIN todo t ON ts.todo_setting_uuid = t.setting_uuid AND t.operate_date = $2 WHERE ts.operated_by_user_uuid = $1`,
			want: `
SELECT
  ts.todo_setting_uuid,
  ts.label,
  ts.day_of_week,
  CASE WHEN t.setting_uuid IS NOT NULL THEN 'COMPLETE' ELSE 'INCOMPLETE' END AS today_status
FROM todo_settings ts
  LEFT JOIN todo t
    ON ts.todo_setting_uuid = t.setting_uuid
  AND t.operate_date = $2
WHERE ts.operated_by_user_uuid = $1
`,
		},
		{
			name: "DATE_FUNCTION_IN_WHERE",
			sql:  `SELECT g.gather_uuid, l.user_id, l.group_id FROM gather g INNER JOIN line_user_gather_relation l ON g.gather_uuid = l.gather_uuid WHERE DATE(g.deadline_date_time) = $1 AND g.deadline_date_time IS NOT NULL AND g.deleted_at IS NULL AND g.confirmed_start_date_time IS NULL AND l.user_id IS NOT NULL`,
			want: `
SELECT
  g.gather_uuid,
  l.user_id,
  l.group_id
FROM gather g
  INNER JOIN line_user_gather_relation l
    ON g.gather_uuid = l.gather_uuid
WHERE date(g.deadline_date_time) = $1
  AND g.deadline_date_time IS NOT NULL
  AND g.deleted_at IS NULL
  AND g.confirmed_start_date_time IS NULL
  AND l.user_id IS NOT NULL
`,
		},
		{
			name: "GROUP_BY_WITH_TYPECAST",
			sql:  `SELECT ga.attendance_date::date::text AS attendance_date, ga.status, count(DISTINCT ga.attendance_name) AS count FROM gather_attendance ga WHERE ga.gather_uuid = $1 GROUP BY ga.attendance_date::date, ga.status ORDER BY ga.attendance_date::date ASC`,
			want: `
SELECT
  ga.attendance_date::date::text AS attendance_date,
  ga.status,
  count(DISTINCT ga.attendance_name) AS count
FROM gather_attendance ga
WHERE ga.gather_uuid = $1
GROUP BY ga.attendance_date::date, ga.status
ORDER BY ga.attendance_date::date ASC
`,
		},
		{
			name: "COMPLEX_SUBQUERY_WITH_UNION_ALL_IN_FUNCTION",
			sql: `SELECT
    g.gather_uuid,
    g.title,
    g.mode,
    g.adjustment_start_date_time,
    g.adjustment_end_date_time,
    g.confirmed_start_date_time,
    g.confirmed_end_date_time,
    g.min_number_of_participants,
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
            ) combined_names
    ), 0) AS number_of_participants
FROM gather g
WHERE g.gather_uuid = ANY($1)
    AND g.deleted_at IS NULL
ORDER BY COALESCE(g.confirmed_start_date_time, g.adjustment_start_date_time) DESC`,
			want: `
SELECT
  g.gather_uuid,
  g.title,
  g.mode,
  g.adjustment_start_date_time,
  g.adjustment_end_date_time,
  g.confirmed_start_date_time,
  g.confirmed_end_date_time,
  g.min_number_of_participants,
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
			name: "CASE_WITH_COMPLEX_SUBQUERY",
			sql: `SELECT
    g.gather_uuid,
    g.title,
    CASE WHEN g.operated_by = $2 THEN 'CREATOR' ELSE 'VIEWER' END AS relationship_type,
    CASE 
        WHEN g.operated_by = $2 THEN NULL 
        ELSE (
            SELECT gvh.gather_view_history_uuid
            FROM gather_view_history gvh
            WHERE gvh.gather_uuid = g.gather_uuid AND gvh.user_uuid = $2
            LIMIT 1
        )
    END AS gather_history_uuid
FROM gather g
WHERE g.gather_uuid = ANY($1)
    AND g.deleted_at IS NULL`,
			want: `
SELECT
  g.gather_uuid,
  g.title,
  CASE WHEN g.operated_by = $2 THEN 'CREATOR' ELSE 'VIEWER' END AS relationship_type,
  CASE WHEN g.operated_by = $2 THEN NULL ELSE (
    SELECT
    gvh.gather_view_history_uuid
  FROM gather_view_history gvh
  WHERE gvh.gather_uuid = g.gather_uuid AND gvh.user_uuid = $2
  LIMIT 1
) END AS gather_history_uuid
FROM gather g
WHERE g.gather_uuid = ANY($1)
  AND g.deleted_at IS NULL
`,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			actual, err := formatter.Format(tt.sql, tt.conf)
			assert.NoError(t, err)
			t.Log(actual)
			if diff := cmp.Diff(tt.want, actual); diff != "" {
				t.Errorf("diff: %s", diff)
			}
		})
	}
}
