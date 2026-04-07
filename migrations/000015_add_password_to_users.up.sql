-- migrations/000015_add_password_to_users.up.sql
ALTER TABLE users ADD COLUMN password_hash VARCHAR(255);
