-- migrations/000007_create_claims_payouts.down.sql
DROP TABLE IF EXISTS payouts;
DROP TABLE IF EXISTS maintenance_check;
DROP TABLE IF EXISTS claim_fraud_scores;
DROP TABLE IF EXISTS claims;
