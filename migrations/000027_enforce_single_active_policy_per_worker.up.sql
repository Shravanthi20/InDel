WITH ranked_active AS (
    SELECT
        id,
        ROW_NUMBER() OVER (
            PARTITION BY worker_id
            ORDER BY COALESCE(updated_at, created_at) DESC, id DESC
        ) AS rn
    FROM policies
    WHERE status = 'active'
)
UPDATE policies p
SET status = 'inactive',
    updated_at = CURRENT_TIMESTAMP
FROM ranked_active r
WHERE p.id = r.id
  AND r.rn > 1;

CREATE UNIQUE INDEX IF NOT EXISTS idx_policies_one_active_per_worker
ON policies(worker_id)
WHERE status = 'active';
