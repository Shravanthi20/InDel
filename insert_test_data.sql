-- Insert test data directly for payout testing
-- This bypasses the disruption/claims flow and tests payouts directly

-- Get a worker ID
WITH worker AS (
  SELECT wp.worker_id, u.id
  FROM worker_profiles wp
  JOIN users u ON u.id = wp.worker_id
  LIMIT 1
)
-- Insert a test disruption
INSERT INTO disruptions (zone_id, type, severity, status, confidence, signal_timestamp, confirmed_at)
SELECT z.id, 'heavy_rain', 'high', 'confirmed', 0.95, NOW(), NOW()
FROM zones z
WHERE z.name = 'Tambaram'
ON CONFLICT DO NOTHING;

-- Insert test claims for eligible workers
WITH eligible_workers AS (
  SELECT wp.worker_id, wp.zone_id, z.id as zone_id_2
  FROM worker_profiles wp
  JOIN zones z ON z.id = wp.zone_id
  WHERE z.name = 'Tambaram'
  LIMIT 30
)
INSERT INTO claims (disruption_id, worker_id, claim_amount, status, fraud_verdict, created_at)
SELECT 
  (SELECT id FROM disruptions WHERE type='heavy_rain' LIMIT 1) as disruption_id,
  ew.worker_id,
  (RANDOM() * 1000 + 500)::numeric(10,2) as claim_amount,
  'approved',
  'clear',
  NOW()
FROM eligible_workers ew
ON CONFLICT DO NOTHING;

-- Insert payouts for created claims
INSERT INTO payouts (claim_id, worker_id, amount, status, created_at)
SELECT 
  c.id,
  c.worker_id,
  c.claim_amount,
  'queued',
  NOW()
FROM claims c
WHERE c.status = 'approved' 
  AND NOT EXISTS (SELECT 1 FROM payouts WHERE claim_id = c.id);

-- Report
\echo '=== TEST DATA INSERTED ==='
SELECT COUNT(*) as "Disruptions" FROM disruptions WHERE type='heavy_rain';
SELECT COUNT(*) as "Claims Created (all)" FROM claims WHERE status IN ('approved', 'manual_review');
SELECT COUNT(*) as "Payouts Queued" FROM payouts WHERE status='queued';
\echo '=== Ready for Phase 6 testing (Payout Processing) ==='
