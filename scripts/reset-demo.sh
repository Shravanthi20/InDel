#!/bin/bash
# Reset demo database to clean state

set -e

echo "Resetting InDel demo database..."

# Drop and recreate database
docker-compose exec -T postgres psql -U indel -c "DROP DATABASE IF EXISTS indel_demo;"
docker-compose exec -T postgres psql -U indel -c "CREATE DATABASE indel_demo;"

# Run migrations
docker-compose exec -T backend migrate -path migrations -database "postgresql://indel:password@postgres:5432/indel_demo?sslmode=disable" up

# Seed demo data
docker-compose exec -T postgres psql -U indel -d indel_demo < scripts/seed.sql

# Generate synthetic data
docker-compose exec -T core python scripts/generate-synthetic-data.py

echo "Demo database reset complete!"
echo "Services available at:"
echo "  Worker API: http://localhost:8001"
echo "  Insurer API: http://localhost:8002"
echo "  Platform API: http://localhost:8003"
echo "  PgAdmin: http://localhost:5050"
