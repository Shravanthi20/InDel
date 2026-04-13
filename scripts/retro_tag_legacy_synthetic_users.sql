-- One-time admin script: retro-tag legacy synthetic users.
--
-- Safety model:
-- 1) DRY RUN by default (do_update = false).
-- 2) Uses strict multi-signal heuristics to avoid touching real app users.
-- 3) Updates only users with empty email so no existing real email is overwritten.
--
-- Target outcome:
-- Legacy synthetic workers get email:
--   synthetic+legacy-<user_id>@synthetic.indel.local
--
-- Run example:
--   psql "$DATABASE_URL" -f scripts/retro_tag_legacy_synthetic_users.sql

BEGIN;

-- Preview candidates first.
WITH candidate_users AS (
    SELECT u.id
    FROM users u
    JOIN worker_profiles wp ON wp.worker_id = u.id
    LEFT JOIN auth_tokens at ON at.user_id = u.id
    WHERE u.role = 'worker'
      AND COALESCE(TRIM(u.email), '') = ''
      AND u.phone ~ '^\+9199[0-9]{8}$'
      AND wp.name ~ '^Worker [0-9]{3}$'
      AND wp.upi_id ~ '^worker[0-9]{3}@upi$'
      AND wp.vehicle_type IN ('two_wheeler', 'bike', 'scooter')
      -- Legacy synthetic users were generally not actively logged in recently.
      AND (at.user_id IS NULL OR at.expires_at < NOW())
)
SELECT u.id, u.phone, u.email, wp.name, wp.upi_id, wp.vehicle_type
FROM candidate_users c
JOIN users u ON u.id = c.id
JOIN worker_profiles wp ON wp.worker_id = u.id
ORDER BY u.id;

DO $$
DECLARE
    do_update boolean := false; -- change to true only after reviewing preview rows above
    affected_count integer := 0;
BEGIN
    IF do_update THEN
        WITH candidate_users AS (
            SELECT u.id
            FROM users u
            JOIN worker_profiles wp ON wp.worker_id = u.id
            LEFT JOIN auth_tokens at ON at.user_id = u.id
            WHERE u.role = 'worker'
              AND COALESCE(TRIM(u.email), '') = ''
              AND u.phone ~ '^\+9199[0-9]{8}$'
              AND wp.name ~ '^Worker [0-9]{3}$'
              AND wp.upi_id ~ '^worker[0-9]{3}@upi$'
              AND wp.vehicle_type IN ('two_wheeler', 'bike', 'scooter')
              AND (at.user_id IS NULL OR at.expires_at < NOW())
        )
        UPDATE users u
        SET email = 'synthetic+legacy-' || u.id || '@synthetic.indel.local',
            updated_at = NOW()
        FROM candidate_users c
        WHERE u.id = c.id;

        GET DIAGNOSTICS affected_count = ROW_COUNT;
        RAISE NOTICE 'Retro-tag complete. Rows updated: %', affected_count;
    ELSE
        RAISE NOTICE 'Dry run only. No rows updated. Set do_update=true in script after verification.';
    END IF;
END $$;

COMMIT;
