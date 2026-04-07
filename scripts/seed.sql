-- Insert users (Required for foreign keys in profiles, policies, etc.)
INSERT INTO users (id, phone, role) VALUES
(1, '+919999990001', 'worker'),
(2, '+919999990002', 'worker'),
(3, '+919999990003', 'worker'),
(4, '+919999990004', 'worker');

-- Insert zones
INSERT INTO zones (id, name, city, state, risk_rating) VALUES
(1, 'Zone-A', 'Bangalore', 'Karnataka', 0.3),
(2, 'Zone-B', 'Bangalore', 'Karnataka', 0.5),
(3, 'Zone-C', 'Mumbai', 'Maharashtra', 0.4),
(4, 'Zone-D', 'Delhi', 'Delhi', 0.6);

-- Insert policy cycles
INSERT INTO weekly_policy_cycles (id, week_start, week_end) VALUES
(1, CURRENT_DATE - INTERVAL '7 days', CURRENT_DATE);

-- Insert workers
INSERT INTO worker_profiles (worker_id, name, zone_id, vehicle_type, upi_id) VALUES
(1, 'Rajesh Kumar', 1, 'bike', 'rajesh@upi'),
(2, 'Priya Singh', 2, 'bike', 'priya@upi'),
(3, 'Amit Patel', 3, 'car', 'amit@upi'),
(4, 'Sneha Gupta', 4, 'bike', 'sneha@upi');

-- Insert baseline earnings
INSERT INTO earnings_baseline (worker_id, baseline_amount, last_updated_at) VALUES
(1, 5000, NOW()),
(2, 6000, NOW()),
(3, 8000, NOW()),
(4, 4500, NOW());

-- Insert policies
INSERT INTO policies (worker_id, status, premium_amount, policy_cycle_id) VALUES
(1, 'active', 300, 1),
(2, 'active', 350, 1),
(3, 'paused', 450, 1),
(4, 'active', 280, 1);

-- Insert sample orders
INSERT INTO orders (worker_id, zone_id, order_value, created_at) VALUES
(1, 1, 450, NOW() - INTERVAL '2 days'),
(1, 1, 500, NOW() - INTERVAL '1 day'),
(2, 2, 600, NOW()),
(3, 3, 750, NOW() - INTERVAL '3 days');

-- Insert earnings records
INSERT INTO earnings_records (worker_id, date, hours_worked, amount_earned) VALUES
(1, CURRENT_DATE - INTERVAL '6 days', 8, 4500),
(1, CURRENT_DATE - INTERVAL '5 days', 9, 5200),
(1, CURRENT_DATE - INTERVAL '4 days', 7, 3800),
(1, CURRENT_DATE - INTERVAL '3 days', 0, 0),  -- Disruption day
(2, CURRENT_DATE - INTERVAL '2 days', 10, 6500),
(2, CURRENT_DATE - INTERVAL '1 day', 8, 5000);

-- Insert disruptions
INSERT INTO disruptions (id, zone_id, type, severity, signal_timestamp) VALUES
(1, 1, 'weather', 'high', NOW() - INTERVAL '3 days'),
(2, 2, 'aqi', 'medium', NOW() - INTERVAL '1 day');

-- Insert claims
INSERT INTO claims (disruption_id, worker_id, claim_amount, status) VALUES
(1, 1, 5000, 'pending'),
(2, 2, 1500, 'approved');
