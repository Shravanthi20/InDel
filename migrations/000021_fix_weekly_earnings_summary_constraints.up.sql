-- Fix weekly_earnings_summary for ON CONFLICT(worker_id, week_start) upserts.
-- 1) Ensure updated_at exists because upsert updates it.
ALTER TABLE weekly_earnings_summary
ADD COLUMN IF NOT EXISTS updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP;

-- 2) Backfill null updated_at values.
UPDATE weekly_earnings_summary
SET updated_at = CURRENT_TIMESTAMP
WHERE updated_at IS NULL;

-- 3) Remove potential duplicates before adding unique constraint.
WITH ranked AS (
  SELECT id,
         ROW_NUMBER() OVER (
           PARTITION BY worker_id, week_start
           ORDER BY id DESC
         ) AS rn
  FROM weekly_earnings_summary
)
DELETE FROM weekly_earnings_summary
WHERE id IN (
  SELECT id FROM ranked WHERE rn > 1
);

-- 4) Add the unique key required by applyWorkerEarningsIncrement upsert.
ALTER TABLE weekly_earnings_summary
ADD CONSTRAINT uq_weekly_earnings_summary_worker_week_start UNIQUE (worker_id, week_start);
