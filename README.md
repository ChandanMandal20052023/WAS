# WAS — Multi-Tenant WhatsApp CRM Platform

**WAS** is a high-performance, multi-tenant WhatsApp CRM platform built from scratch. It empowers businesses to connect multiple WhatsApp accounts via QR-code device linking, manage customer conversations, organize contacts, and coordinate support teams through an intuitive, centralized dashboard.

---

## Key Highlights

- **Multi-Tenant SaaS Architecture:** Strict data isolation per organization with role-based access control (`Owner`, `Admin`, `Manager`, `Agent`).
- **Native WhatsApp Integration:** Powered by [Evolution Go](https://github.com/EvolutionAPI/evolution-go) using the lightweight `whatsmeow` engine.
- **Centralized Three-Panel Inbox:** Seamless agent workspace featuring conversation filtering, customer profile sidebars, and internal collaboration notes.
- **Real-Time Streaming:** Sub-second inbound and outbound message synchronization powered by WebSockets and Redis pub/sub.
- **Automated Contact Intelligence:** Contact deduplication via E.164 normalization, customer tagging, and conversation history.
- **Production-Ready Containerization:** Docker Compose setup orchestrating PostgreSQL, Redis, Evolution Go, Go Backend, and React Frontend.

---

## Technology Stack

### Frontend
- **Framework:** React.js 18+ with TypeScript
- **Tooling:** Vite, Tailwind CSS, PostCSS
- **Component Primitives:** shadcn/ui & Radix UI
- **Routing & State:** React Router v6, TanStack Query v5, Zustand
- **Forms & Validation:** React Hook Form, Zod

### Backend
- **Language:** Golang (1.23+)
- **Router:** Chi router (`/api/v1/`)
- **Database Driver:** `pgx` with `pgxpool`
- **Migrations:** `golang-migrate`
- **Real-Time:** Gorilla WebSocket, Redis Pub/Sub
- **Architecture:** Modular Monolith (domain-driven handlers, services, repositories)

### Messaging Gateway
- **Engine:** Evolution Go (`whatsmeow` library)
- **Adapter Support:** Dual-mode gateway supporting both live Evolution Go instances and a high-fidelity mock adapter for isolated testing.

### Infrastructure
- **Databases:** PostgreSQL 16 (persistent relational state), Redis 7 (in-memory caching & event distribution)
- **Deployment:** Docker & Docker Compose

---

## Monorepo Directory Structure

```
WAS/
├── frontend/                     # React + TypeScript + Vite SPA
│   ├── public/                   # Static assets & icons
│   ├── src/
│   │   ├── components/           # shadcn/ui primitives & reusable UI
│   │   ├── layouts/              # Auth & Dashboard app layouts
│   │   ├── pages/                # Dashboard, Inbox, Channels, Contacts, Team
│   │   ├── features/             # Domain UI modules (inbox, channels, team)
│   │   ├── hooks/                # Custom React & WebSocket hooks
│   │   ├── services/             # Axios/fetch API clients
│   │   ├── stores/               # Zustand client stores
│   │   ├── types/                # TypeScript interfaces
│   │   └── utils/                # Formatting & utility functions
│   ├── package.json
│   └── Dockerfile
│
├── backend/                      # Golang Chi Modular Monolith
│   ├── cmd/
│   │   ├── api/                  # Main HTTP & WebSocket server entrypoint
│   │   └── migrate/              # Database migration CLI runner
│   ├── internal/
│   │   ├── config/               # Environment & configuration parser
│   │   ├── database/             # pgxpool & Redis client singletons
│   │   ├── middleware/           # Tenant isolation, JWT, RBAC, logging
│   │   ├── auth/                 # Authentication & token management
│   │   ├── organizations/        # Organization hierarchy & settings
│   │   ├── channels/             # WhatsApp channel instances
│   │   ├── evolution/            # Evolution Go API client & webhook handler
│   │   ├── contacts/             # Contact repository & deduplication
│   │   ├── conversations/        # Conversation threads & status state machine
│   │   ├── messages/             # Real-time messages & media handling
│   │   ├── assignments/          # Agent workloads & conversation assignment
│   │   ├── websocket/            # WebSocket hub & client connection pool
│   │   ├── notifications/        # User notifications
│   │   └── audit/                # Immutable audit logging
│   ├── migrations/               # PostgreSQL schema migration scripts
│   ├── go.mod
│   └── Dockerfile
│
├── third-party/
│   └── evolution-go/             # Preserved Evolution Go WhatsApp engine
│
├── infrastructure/
│   └── docker/
│       └── postgres/             # Database initialization scripts
│
├── docs/                         # Architecture, Database, API, and Setup specs
│   ├── architecture.md
│   ├── database.md
│   ├── api.md
│   └── setup.md
│
├── docker-compose.yml            # Complete orchestration of all services
├── .env.example                  # Environment configuration template
├── .gitignore
├── README.md
└── AI_USAGE.md
```

---

## Development Roadmap

| Phase | Description | Status |
|:---:|---|:---:|
| **1** | Repository Inspection & System Architecture | **Complete** |
| **2** | Initialize React, Go, PostgreSQL, Redis, and Docker | Up Next |
| **3** | Database Schema & SQL Migrations | Planned |
| **4** | Authentication & Multi-Tenancy Isolation | Planned |
| **5** | WhatsApp Evolution Go Integration & QR Pairing | Planned |
| **6** | Contacts Management & Conversation Pipelines | Planned |
| **7** | Centralized 3-Panel Inbox & WhatsApp Messaging | Planned |
| **8** | Real-Time WebSockets & In-App Notifications | Planned |
| **9** | Team Management & Agent Workload Assignment | Planned |
| **10** | End-to-End Testing, Security Audits & Deployment | Planned |

---

## Quick Start (Docker Compose)

### 1. Prerequisites
- Docker Engine 24+ & Docker Compose v2+
- Git

### 2. Configure Environment
```bash
cp .env.example .env
# Edit .env and supply your preferred secrets
```

### 3. Launch the Stack
```bash
docker compose up -d --build
```

### 4. Access Services
- **WAS Frontend:** `http://localhost:3000`
- **WAS Backend API:** `http://localhost:8080/api/v1`
- **Evolution Go Engine:** `http://localhost:8081`

---

## License

This project is licensed under the Apache 2.0 License.

