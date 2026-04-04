-- Seed demo data for testing
-- Insert 500 workers if not already present
INSERT INTO users (phone, email, role, password_hash) 
SELECT '9999' || LPAD(i::text, 6, '0'), 'worker' || i || '@indel.com', 'worker', 'hash'
FROM generate_series(1, 500) as i
WHERE NOT EXISTS (SELECT 1 FROM users WHERE role = 'worker' AND phone LIKE '9999%')
ON CONFLICT (phone) DO NOTHING;

-- Create worker profiles for all users
INSERT INTO worker_profiles (worker_id, name, zone_id, upi_id)
SELECT 
  u.id,
  'Worker ' || u.id,
  (u.id % 12) + 1 as zone_id,
  'upi' || LPAD(u.id::text, 8, '0') || '@okaxis'
FROM users u
WHERE u.role = 'worker' AND NOT EXISTS (SELECT 1 FROM worker_profiles WHERE worker_id = u.id)
LIMIT 500;

-- Create active policies for all workers
INSERT INTO policies (worker_id, status, premium_amount)
SELECT 
  wp.worker_id,
  'active',
  500
FROM worker_profiles wp
WHERE NOT EXISTS (SELECT 1 FROM policies WHERE worker_id = wp.worker_id)
LIMIT 500;

-- Report
\echo '✓ Demo Data Seeding Complete'
SELECT COUNT(*) as "Workers Created" FROM users WHERE role = 'worker';
SELECT COUNT(*) as "Profiles Created" FROM worker_profiles;
SELECT COUNT(*) as "Active Policies Created" FROM policies WHERE status = 'active';
SELECT COUNT(*) as "Total Zones" FROM zones;
