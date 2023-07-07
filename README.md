# gopsqlfmt

Format SQL strings in go files.

# install command

`go install github.com/Toru-Takagi/gopsqlfmt@latest`


# exec

1. `$ cd [your go project root]`
2. `$ gopsqlfmt ./...`

# Sample

### before

```go
func main() {
	const selectSQL = `select u.user_name, ull.last_login_at, uage.user_age, uadd.address   ,
					array_agg(user_uuid), now(), gen_random_uuid() ,
			  COALESCE(( SELECT json_agg(json_build_object('userUUID', gu.user_uuid, 'userName', gu.user_name)) AS results FROM gest_users gu), '[]') AS results,
					(select ull.last_login_at, current_setting('test') from user_last_login ull where ull.user_uuid = u.user_uuid and u.email = :email) as last_login_at
from users u
						inner join user_last_login ull on u.user_uuid = ull.user_uuid
						left join user_age uage on u.user_uuid = uage.user_uuid
						left join user_address uadd on u.user_uuid = uadd.user_uuid
						where u.user_uuid = $1`

	const insertSQL = `
					insert into users (user_uuid, user_name, user_age) values ($1, $2, $3)
								on conflict (user_uuid) do update set user_name = EXCLUDED.user_name, user_age = EXCLUDED.user_age, updated_at = now()
	`
}

```

### after

```go
func main() {
	const selectSQL = `
SELECT
  u.user_name,
  ull.last_login_at,
  uage.user_age,
  uadd.address,
  array_agg(user_uuid),
  now(),
  gen_random_uuid(),
  COALESCE((
    SELECT
      json_agg(json_build_object('userUUID', gu.user_uuid, 'userName', gu.user_name)) AS results
    FROM gest_users gu
  ), '[]') AS results,
  (
    SELECT
      ull.last_login_at,
      current_setting('test')
    FROM user_last_login ull
    WHERE ull.user_uuid = u.user_uuid
      AND u.email = :email
  ) AS last_login_at
FROM users u
  INNER JOIN user_last_login ull
    ON u.user_uuid = ull.user_uuid
  LEFT JOIN user_age uage
    ON u.user_uuid = uage.user_uuid
  LEFT JOIN user_address uadd
    ON u.user_uuid = uadd.user_uuid
WHERE u.user_uuid = $1
`
	const insertSQL = `
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
`
}
```
