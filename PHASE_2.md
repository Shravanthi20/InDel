# InDel — Implementation Phase 2

This document provides a practical, execution-focused overview of the InDel platform as implemented in **Phase 2**. For the theoretical background and project vision, refer to the [original README](file:///f:/DevTrails/InDel/README.md).

## 🏗️ Functional Architecture

InDel is built as a high-precision, event-driven ecosystem. The Phase 2 implementation focuses on three core pillars:

### 1. The Core Infrastructure (Backend)
- **Technology**: Go (Gin), PostgreSQL (GORM), JWT Authentication.
- **Port**: `8003`
- **Key Logic**: Handles worker registration, policy management, real-time zone monitoring (`DisruptionEngine`), and automated claim/payout generation.

### 2. Insurer Command Center (Insurer Dashboard)
- **Technology**: React (Vite), Tailwind CSS, Lucide Icons.
- **Port**: `5173`
- **Goal**: Provides insurers with high-density data on risk, premium health, and a dedicated queue for manual fraud verification.
- **Theme**: Enterprise High-Precision (Slate/Orange).

### 3. Platform Mission Control (Platform Dashboard)
- **Technology**: React (Vite), Tailwind CSS, Lucide Icons.
- **Port**: `5174`
- **Goal**: Regional telemetry monitoring and the **Chaos Engine** — a specialized tool for simulating environmental and demand disruptions.

---

## ⚡ Key Phase 2 Features

### 🎨 Unified Enterprise Design
A cohesive design system implemented across all dashboards:
- **Typography**: Integrated `Outfit` font for maximum legibility.
- **Dark/Light Modes**: Standardized theme switching using CSS variables and React Context.
- **High-Density Layouts**: "Mission Control" style panels for displaying complex telemetry and workers.
- **Snappy Response**: Global transitions minimized for a fast, enterprise-grade feel.

### ⚙️ Automation & Chaos Engine
- **Parametric Triggers**: Automated detection of Rain, AQI, and Curfew events via backend engines.
- **Simulation Layer**: The Chaos Engine allows developers to manually:
  - **Collapse Demand**: Force a drop in regional order activity.
  - **Inject Signals**: Trigger simulated weather or local alerts.
- **Instant Processing**: Claims are automatically generated and queued for payout the moment a multi-signal disruption is confirmed.

### 🔍 Dynamic Data & Operations
- **Real-Time Telemetry**: Dashboards use 2-second polling to reflect live backend state.
- **Active Search & Filter**: Robust client-side filtering for searching workers, zones, and historical analytics.
- **Data Export**: Built-in CSV generation for auditing worker lists and disruption feeds.

---

## 🚀 Local Development Setup

### 1. Prerequisites
- **Go** (1.25.0+)
- **Node.js** (v18+)
- **PostgreSQL** (running locally or via Docker)

### 2. Backend Setup
```bash
cd backend
go mod download
# Update .env with your credentials
go run cmd/api/main.go
```

### 3. Dashboard Setup (Run for both Insurer and Platform)
```bash
cd insurer-dashboard # or platform-dashboard
npm install
npm run dev
```

### 4. Database Initialization
Migrations and seed data are managed via the backend's internal logic. On first startup, the system initializes:
- Standardised InDel Zones (Tambaram, Koramangala, etc.)
- Mock Workers and active policies
- Baseline telemetry benchmarks

---

## 📂 Project Directory Map

- `/backend`: Go source code, API routes, and business logic.
- `/insurer-dashboard`: React frontend for insurance providers.
- `/platform-dashboard`: React frontend for platform operators (Mission Control).
- `/worker-app`: Native mobile application for delivery workers.
- `/migrations`: SQL schema and baseline data.
- `/ml`: Model training scripts and synthetic data generators.

---

> [!NOTE]
> Phase 2 implementation successfully unified the disparate MVP components into a professional "InDel Suite" with a consistent enterprise skin and functional telemetry.
