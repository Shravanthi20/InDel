-- Rollback for 000011_add_more_zones.sql
DELETE FROM zones WHERE name IN (
    'Tambaram', 'Selaiyur', 'Pallikaranai', 'Chromepet', 'Velachery', 'Medavakkam',
    'Whitefield', 'Koramangala', 'Indiranagar',
    'Bandra', 'Dadar', 'Andheri'
);
