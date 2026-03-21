-- migrations/000009_create_audit_tables.down.sql
DROP TABLE IF EXISTS auth_tokens;
DROP TABLE IF EXISTS api_request_logs;
DROP TABLE IF EXISTS kafka_event_logs;
DROP TABLE IF EXISTS idempotency_keys;
DROP TABLE IF EXISTS fcm_tokens;
DROP TABLE IF EXISTS notifications;
