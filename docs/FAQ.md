# Frequently Asked Questions

## General

### Q: What is InDel?
A: InDel is an income protection insurance platform for gig workers. It automatically detects disruptions (weather, AQI, order drops) and generates claims with ML-powered fraud detection.

### Q: What gig platforms does InDel support?
A: V1 supports Amazon Flex via order webhooks. V2 will add DoorDash, Uber Eats, etc.

### Q: How is the baseline income calculated?
A: Baseline = Average of last 4 weeks of earnings during non-disruptive hours for the worker's zone.

## Worker App

### Q: When am I eligible for a claim?
A: You're eligible when there's a confirmed disruption in your zone AND your earnings fell below baseline.

### Q: How long does it take to get paid?
A: Claims are typically processed within 24 hours, payouts within 2-3 hours via Razorpay.

### Q: What should I do if my claim is denied?
A: You can request a maintenance check within 7 days. An insurer will review your claim.

### Q: How is the premium calculated?
A: Premium is calculated by an XGBoost model considering your:
- Historical earnings volatility
- Disruption frequency in your zone
- Zone risk rating
- Age of account

## Insurer Portal

### Q: What does "Loss Ratio" mean?
A: Loss Ratio = Total Claims Paid / Total Premiums Collected. Lower is better.

### Q: How do I review a fraudulent claim?
A: Claims with high fraud scores appear in the Fraud Queue. Click to see signals (historical claims, unusual patterns, rule violations).

### Q: What does the 7-day forecast show?
A: It predicts disruption probability for each zone using Prophet, helping with reserve planning.

## Technical

### Q: Can I run InDel locally?
A: Yes! Run `docker-compose up -d` and all services start locally on ports 8000-8003, 9001-9003.

### Q: How is the database structured?
A: PostgreSQL with 9 migration sets. See DATABASE_DESIGN.md.

### Q: How does fraud detection work?
A: 3-layer approach:
1. Isolation Forest (global anomaly detection)
2. DBSCAN (temporal clustering)
3. Rules (hard disqualifiers like duplicate claims)

### Q: How are messages published?
A: Via Kafka. Topics include:
- `indel.claims.generated`
- `indel.claims.scored`
- `indel.payouts.queued`
- `indel.weather.alerts`
- `indel.aqi.alerts`

## Deployment

### Q: Where can I deploy InDel?
A: GitHub Actions CI/CD deploys to Render. Can also deploy to AWS, GCP, Azure.

### Q: How do I seed demo data?
A: Run `docker-compose exec postgres psql < scripts/seed.sql`

### Q: How do I reset the database?
A: Run `scripts/reset-demo.sh` (creates fresh tables and seeds demo data)

## Troubleshooting

### Q: Database connection fails
A: Ensure PostgreSQL is running. Check `.env` file has correct `DB_HOST`, `DB_PASSWORD`.

### Q: Kafka connection fails
A: Ensure Zookeeper and Kafka are running. Check `KAFKA_BROKERS` in `.env`.

### Q: Worker app can't connect to API
A: Ensure backend gateways are running. Check `VITE_WORKER_API_URL` in frontend `.env`.

### Q: ML service fails to start
A: Ensure Python 3.11+ and dependencies installed. Run `pip install -r ml/requirements.txt`.

## Roadmap Questions

### Q: When will you support [platform]?
A: Open an issue on GitHub. V2 is planned for Q3 2026.

### Q: Can I contribute?
A: Yes! See CONTRIBUTING.md (to be created).

### Q: Is there a license?
A: TBD. Check LICENSE file.
