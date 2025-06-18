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
ORDER BY COALESCE(g.confirmed_start_date_time, g.adjustment_start_date_time) DESC