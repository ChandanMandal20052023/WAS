# AI Usage & Engineering Record — WAS Platform

This document logs the AI collaboration guidelines, architectural rationale, and execution phases followed during the development of the **WAS** WhatsApp CRM platform.

---

## System Guidelines & Constraints

1. **Strict Multi-Tenancy:**
   - Every database query touching tenant-scoped entities must explicitly filter by `organization_id`.
   - Never trust frontend assertions regarding tenant identity or role capabilities.
   - Tenant context must be securely injected via verified server-side JWT authentication.

2. **Integration Fidelity:**
   - WhatsApp integration targets the official `evolution-go` repository specifications.
   - Endpoints, headers (`apikey`), and webhook payload schemas match actual `evolution-go` source code rather than documentation assumptions.
   - For offline development and automated CI testing, a high-fidelity `MockEvolutionAdapter` must be provided that faithfully simulates the Evolution Go contract without bypassing production checks.

3. **Incremental Modular Implementation:**
   - Each phase is verified independently before proceeding to subsequent layers.
   - Handlers, business logic services, and database repositories maintain clean separation of concerns.

---

## Log of Execution Phases

### Phase 1: Repository Inspection & Architectural Blueprint
- **Scope:** Complete analysis of current workspace, inspection of `evolution-go` codebase, design of database schema, ER diagrams, REST API specs, and monorepo structure.
- **Key Findings:**
  - `evolution-go` requires API key header `apikey` rather than Bearer token.
  - Text messaging route is `POST /send/text` with JSON body `{"number": "...", "text": "..."}`.
  - Webhooks dispatch to configured URL with root object `{ event, instanceToken, instanceId, instanceName, data }`.
  - Evolution Go v0.4.x includes runtime licensing gate (`core.GateMiddleware`); integration includes dual adapter design (Live + Mock) to guarantee testability.
- **Artifacts:**
  - `README.md`, `AI_USAGE.md`, `.env.example`, `.gitignore`
  - `docs/architecture.md`, `docs/database.md`, `docs/api.md`, `docs/setup.md`
  - `docker-compose.yml`

