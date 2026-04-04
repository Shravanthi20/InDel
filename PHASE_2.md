# InDel — Implementation Phase 2

This document provides a technical and execution-focused overview of the InDel platform as implemented in Phase 2.

## Functional Architecture

InDel is an event-driven parametric insurance ecosystem designed to protect gig-worker income from regional disruptions. The architecture is composed of a high-performance Go backend, dual enterprise React dashboards, and a native mobile application.

### System Components

The following diagram illustrates the high-level relationship between the platform's core components:

```mermaid
graph TD
    subgraph "External Layers"
        WA[Worker Mobile App]
        ES[External Signals - Weather/AQI]
    end

    subgraph "InDel Core (Backend)"
        API[Gin REST API]
        DE[Disruption Engine]
        CO[Core Ops Service]
        DB[(PostgreSQL)]
    end

    subgraph "Management Layers"
        PD[Platform Dashboard]
        ID[Insurer Dashboard]
    end

    WA <-->|Real-time Tracking| API
    ES -->|Event Ingestion| API
    API <--> DB
    DE <--> DB
    CO <--> DB
    PD <-->|Simulation / Telemetry| API
    ID <-->|Risk Management| API
```

###  Economic Impact Lifecycle

This sequence diagram traces the flow from a triggered disruption to an automated payout:

```mermaid
sequenceDiagram
    participant C as Chaos Engine (Simulator)
    participant B as Backend (Disruption Engine)
    participant W as Worker Activity Telemetry
    participant G as Claim Generator
    participant P as Payout Processor

    C->>B: Inject Signal (e.g., Heavy Rain)
    B->>B: Monitor Zone Baseline vs Actuals
    W->>B: Report Drop in Order Volume
    Note over B: Multi-Signal Validation (Signal + Telemetry)
    B->>B: Confirm Disruption Event
    B->>G: Trigger Claim Scoping
    G->>G: Identify Eligible Workers in Zone
    G->>G: Calculate Income Loss (Baseline - Actual)
    G->>B: Save Approved Claims
    B->>P: Queue Asynchronous Payouts
    P->>P: Process via Kafka/Razorpay Mock
```

---

## Repository Structure

The repository is organized into distinct modules for backend logic, dashboard management, and mobile operations.

###  Repository Root
- `backend/`: Core Go API and business logic.
- `insurer-dashboard/`: React application for insurance providers.
- `platform-dashboard/`: React application for platform operators and simulator control.
- `worker-app/`: Native Kotlin implementation for worker-side tracking and notifications.
- `migrations/`: SQL schema definitions and baseline data seeds.
- `ml/`: Model training scripts and synthetic data generators.
- `PHASE_2.md`: This implementation-focused documentation.
- `README.md`: High-level project vision and theoretical background.

###  Backend Deep Dive (`/backend`)
- `cmd/api/`: Entry point for the Gin server.
- `internal/handlers/`: Domain-specific API endpoint logic.
    - `platform/`: Zone monitoring and Chaos Engine endpoints.
    - `insurer/`: Policy and risk analytics endpoints.
    - `worker/`: Identity and notification endpoints.
    - `demo/`: Specialized simulation and seeding handlers.
- `internal/services/`: The "Engine" layer containing core business logic.
    - `disruption_engine.go`: Logic for real-time baseline calculation and disruption confirmation.
    - `core_ops_service.go`: Batch processing for claim generation, eligibility checks, and payouts.
    - `premium_pricing.go`: Dynamic risk-based pricing logic.
- `internal/models/`: GORM-based entity definitions (Users, Zones, Policies, Claims).
- `internal/router/`: Centralized Gin route definitions.

###  Dashboard Deep Dive (`/platform-dashboard` & `/insurer-dashboard`)
- `src/api/`: Axios-based clients for backend communication.
- `src/components/`: Reusable UI components.
    - `layout/`: Shared Sidebar and Navbar with theme management.
    - `ui/`: Atomic components like panels, badges, and metrics.
- `src/pages/`: Feature-specific views.
    - `Overview.tsx`: Holistic system health telemetry.
    - `Disruptions.tsx`: The "Chaos Engine" simulation interface.
    - `Workers.tsx`: Native-style searchable node directory.
- `src/context/`: React Context for Theme (Light/Dark) and Global State.

---

##  Core Functional Components

### 1. Disruption Engine (The Brain)
The engine maintains a sliding 10-minute window of regional order volume. It calculates a **Dynamic Baseline** for each zone. A disruption is confirmed only through **Multi-Signal Validation**:
- **Environmental**: External signal (Weather, AQI, Curfew).
- **Economic**: Internal telemetry showing a >30% drop in order volume relative to the baseline.

### 2. Core Ops Service (The Scale)
This service handles high-volume batch operations. When a disruption is confirmed, it:
1. Scans the database for all workers with **Active Policies** in the affected zone.
2. Identifies workers who were **Logged In** during the disruption.
3. Computes **Income Loss** using the worker's historical 4-week average vs. actual earnings.
4. Generates a **Claim Record** with an automated fraud verdict based on the signal strength.

### 3. Chaos Engine (The Simulator)
Integrated directly into the Platform Dashboard, the Chaos Engine provides a practical way to test the entire parametric pipeline. It allows for:
- **Demand Collapse**: Artificially resetting the baseline to simulate a massive order drop.
- **Signal Injection**: Sending real-time weather or local restriction events to the backend.

---
Platform Dashboard
<img width="2550" height="1376" alt="image" src="https://github.com/user-attachments/assets/eb6861dc-981a-4421-974c-f0f906a308e2" />
<img width="2558" height="1344" alt="image" src="https://github.com/user-attachments/assets/78f9734b-e6f1-4474-9797-b99360365a19" />
<img width="2560" height="1374" alt="image" src="https://github.com/user-attachments/assets/964e5256-7f20-4a2f-953a-c0d80081d896" />
<img width="2552" height="1370" alt="image" src="https://github.com/user-attachments/assets/cb0c9ec4-ec01-4a94-ad90-9bf3be66f369" />
<img width="2560" height="1372" alt="image" src="https://github.com/user-attachments/assets/4777ca60-9f84-4827-9a67-57bc910362ee" />

Insurer Dashboard
<img width="2548" height="1374" alt="image" src="https://github.com/user-attachments/assets/62f26bb3-c1b7-4686-9c86-80f05a13ae5b" />
<img width="2554" height="1372" alt="image" src="https://github.com/user-attachments/assets/ee38be36-7289-4c80-bb41-66c2ae521e55" />
<img width="2560" height="1362" alt="image" src="https://github.com/user-attachments/assets/840e9bfd-b5f9-494b-8b68-adfb4eb03afc" />
<img width="2558" height="1372" alt="image" src="https://github.com/user-attachments/assets/3300467d-52bf-4db1-a340-5c68543cf28b" />
<img width="2456" height="1372" alt="image" src="https://github.com/user-attachments/assets/eac3ad8d-6677-4bb8-ac60-8cef07693071" />

Mobile App
<img width="702" height="1600" alt="image" src="https://github.com/user-attachments/assets/b4b9d0f8-4dda-40e0-90b9-9aba2c48b065" />
<img width="702" height="1600" alt="image" src="https://github.com/user-attachments/assets/34664da4-b267-43c0-8ee3-2bd181abfd3c" />
<img width="702" height="1600" alt="image" src="https://github.com/user-attachments/assets/64637df5-53ab-4964-b3d6-55d578f4ae77" />
<img width="702" height="1600" alt="image" src="https://github.com/user-attachments/assets/7f62b262-60b9-42e0-b967-0fc35e88f415" />

---
## Technical Specifications

- **Backend**: Go (Gin), PostgreSQL (GORM), JWT.
- **Frontend**: React 18 (Vite), Tailwind CSS, Lucide Icons.
- **Theme**: Enterprise High-Precision (Slate/Orange).
- **Communication**: 2-second polling interval for real-time telemetry updates.

> [!NOTE]
> All automated systems are idempotent, ensuring that mass disruption events do not cause duplicate claim generation or payout errors.
