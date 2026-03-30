-- Add more zones for enhanced coverage
-- Chennai zones
INSERT INTO zones (name, city, state, risk_rating) VALUES
('Tambaram', 'Chennai', 'Tamil Nadu', 0.45),
('Selaiyur', 'Chennai', 'Tamil Nadu', 0.40),
('Pallikaranai', 'Chennai', 'Tamil Nadu', 0.50),
('Chromepet', 'Chennai', 'Tamil Nadu', 0.55),
('Velachery', 'Chennai', 'Tamil Nadu', 0.35),
('Medavakkam', 'Chennai', 'Tamil Nadu', 0.48)
ON CONFLICT (name) DO NOTHING;

-- Additional Bangalore zones
INSERT INTO zones (name, city, state, risk_rating) VALUES
('Whitefield', 'Bangalore', 'Karnataka', 0.42),
('Koramangala', 'Bangalore', 'Karnataka', 0.38),
('Indiranagar', 'Bangalore', 'Karnataka', 0.40)
ON CONFLICT (name) DO NOTHING;

-- Additional Mumbai zones
INSERT INTO zones (name, city, state, risk_rating) VALUES
('Bandra', 'Mumbai', 'Maharashtra', 0.45),
('Dadar', 'Mumbai', 'Maharashtra', 0.52),
('Andheri', 'Mumbai', 'Maharashtra', 0.48)
ON CONFLICT (name) DO NOTHING;
