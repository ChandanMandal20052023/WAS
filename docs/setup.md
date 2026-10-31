# WAS — Development & Deployment Setup Guide

---

## 1. Prerequisites

Ensure you have the following installed locally:
- **Docker Engine** 24.0+ and **Docker Compose** v2.20+
- **Golang** 1.23+ (for bare-metal development)
- **Node.js** 20+ and **npm** or **pnpm** (for bare-metal frontend development)
- **Git**

---

## 2. Docker Compose Setup (Recommended)

### Step 1: Clone & Configure
```bash
git clone https://github.com/ChandanMandal20052023/WAS.git
cd WAS
cp .env.example .env
```

### Step 2: Build & Start Containers
```bash
docker compose up -d --build
```

### Step 3: Verify Container Health
```bash
docker compose ps
```
All containers (`was-postgres`, `was-redis`, `was-evolution`, `was-backend`, `was-frontend`) should display `healthy` or `running`.

### Step 4: Access Application
- **WAS Web Dashboard:** `http://localhost:3000`
- **Backend API:** `http://localhost:8080/api/v1`
- **Evolution Go Manager / Engine:** `http://localhost:8081`

---

## 3. Local Bare-Metal Development

### 1. Database & Cache Services
Run only PostgreSQL and Redis via Docker Compose:
```bash
docker compose up -d was-postgres was-redis
```

### 2. Backend Server
```bash
cd backend
cp ../.env.example .env
go mod download
go run cmd/api/main.go
```

### 3. Frontend Development Server
```bash
cd frontend
npm install
npm run dev
```
Open `http://localhost:5173` in your browser.

---

## 4. Evolution Go Configuration

Evolution Go acts as the WhatsApp gateway service.
- **Port:** `8080` (mapped to `8081` on host in Docker Compose).
- **Global API Key:** Configured via `EVOLUTION_GLOBAL_API_KEY` in `.env`.
- **Database Connection:** Connects to PostgreSQL databases `evogo_auth` and `evogo_users`.
- **Mock Mode:** If you do not have WhatsApp test devices available, set `USE_MOCK_EVOLUTION_ADAPTER=true` in `.env` to enable the built-in emulator.
