# 🦷 SimilePro API

> **High-Performance Dental Clinic Management System (SaaS Backend)**

![Go Version](https://img.shields.io/badge/Go-1.25+-00ADD8?style=flat&logo=go)
![Docker](https://img.shields.io/badge/Docker-Enabled-2496ED?style=flat&logo=docker)
![PostgreSQL](https://img.shields.io/badge/PostgreSQL-15-4169E1?style=flat&logo=postgresql)
![Redis](https://img.shields.io/badge/Redis-7-DC382D?style=flat&logo=redis)
![Swagger](https://img.shields.io/badge/Swagger-OpenAPI-85EA2D?style=flat&logo=swagger)

**SimilePro** is a robust, production-ready RESTful API built in **Go (Golang)**. It is designed to serve as the highly scalable backend for a multi-tenant Dental Clinic Management SaaS.

This repository demonstrates advanced architectural patterns, strict security compliance, and performance optimization techniques tailored for modern cloud environments.

---

## 🏛️ Engineering & Architecture Highlights

This API was engineered with a strong focus on performance and security, mitigating standard OWASP Top 10 vulnerabilities.

### ⚡ Performance & Scalability
* **Zero-Allocation DTOs & Projections:** Heavy endpoints bypass full ORM preloads (GORM) in favor of lightweight DTOs and optimized SQL Joins, significantly reducing memory footprint and GC (Garbage Collection) pressure.
* **Strict Pagination & Bounded Queries:** Engineered date-bound query constraints and offset/limit pagination to prevent full-table scans (e.g., bounding schedule queries to 24-hour windows), dropping response times from minutes to milliseconds.
* **Connection Pooling:** Finely tuned database connection pools for high-concurrency environments.

### 🛡️ Security & Compliance (OWASP Top 10)
* **RSA-Signed JWT Authentication:** Stateless and secure authentication using asymmetric RSA keys.
* **Data Encryption at Rest:** Sensitive patient data (such as National IDs/CPFs) are encrypted in the database using **AES-GCM**.
* **Redis-Backed Rate Limiting:** Distributed rate limiting blocks brute-force and DoS attacks at the middleware layer.
* **Strict RBAC & Multi-Tenancy:** Role-Based Access Control isolates permissions (e.g., Dentists vs. HR vs. Admins), while a strict Tenant DB Middleware guarantees data isolation between different clinics.
* **Audit Logging:** Global middleware captures and logs every sensitive state change in the system for accountability.

---

## 🛠️ Tech Stack

* **Language:** Go (Golang) 1.25+
* **Framework:** Gin Gonic (High-performance HTTP web framework)
* **Database / ORM:** PostgreSQL 15 & GORM
* **Cache & Security:** Redis 7 (Rate limiting & ephemeral storage)
* **CI/CD:** GitHub Actions
* **Infrastructure:** Docker, Docker Compose, Cloudflare Tunnels (Zero Trust Network Access)
* **Documentation:** Swaggo (OpenAPI 2.0)

---

## 🚀 Key Business Modules

### 🏥 Clinical Core
* **Patients & Records:** Full lifecycle management with dynamic health alerts (Allergies, Diabetes, etc.).
* **Scheduling:** Conflict-free appointment management with real-time status tracking.
* **Treatment Plans & Evolutions:** Itemized odontological budgets and chronological clinical notes.

### 💰 Financial & HR (Strict RBAC)
* **Accounts Payable/Receivable:** Full transaction tracking and invoice (Faturas) management.
* **Cash Flow Dashboard:** Real-time BI metrics and cash flow aggregations.
* **Payroll (Folha):** Dynamic salary calculations, pay stubs, and HR management.

---

## 🔄 CI/CD Pipeline

This project utilizes **GitHub Actions** to enforce continuous integration and code quality. The pipeline triggers on every `push` and `pull request` to the main branches, performing:

1. **Environment Setup:** Provisions the Go 1.25 environment.
2. **Dependency Resolution:** Downloads and caches Go modules.
3. **Swagger Generation:** Automatically regenerates the `docs/swagger.json` file ensuring documentation is never out of sync with the code.
4. **Build & Test:** Compiles the binary and runs the automated unit and integration test suites.

---

## ⚙️ Getting Started (Local Development)

The easiest way to spin up the entire ecosystem (Backend, PostgreSQL, Redis, and Cloudflare Tunnel) is via Docker.

### Prerequisites
* Docker & Docker Compose
* Go 1.25+ (If running locally without Docker)

### Running with Docker Compose
1. **Clone the repository.**
2. **Setup Environment Variables:** Copy `.env.example` to `.env` in the root directory and fill in the required values. The Cloudflare Tunnel token must be set as `CLOUDFLARE_TUNNEL_TOKEN`; Compose stops with a clear error when it is missing. Keep `.env` local and never commit it:
   ```env
   ENV=development
   DB_HOST=similepro_postgres
   DB_USER=postgres
   DB_PASS=root
   DB_NAME=SimilePro
   REDIS_ADDR=similepro_redis:6379
   SERVER_PORT=8080
   JWT_SECRET=your_jwt_secret
   ```
   Add your Cloudflare Tunnel token to the local `.env` file:
   ```env
   CLOUDFLARE_TUNNEL_TOKEN=<paste-your-tunnel-token-here>
   ```
3. **Build and Run:**
   ```bash
   docker compose up -d --build
   ```
   *The database tables will be auto-migrated on startup.*

### Gateway deployment

The standalone `SimilePro-gateway` Docker image is the single ingress for the
web application and API. It routes `/api/*` to this service and all other
paths to the Next.js frontend. The gateway and both applications join the
external `similepro_global_net` Docker network.

Configure the Cloudflare Tunnel public hostname to use
`http://similepro_gateway:8080`; do not route the public hostname directly to
the API or frontend. Start the standalone gateway Compose project as well as
the frontend and backend projects. Server-side frontend API calls use
`API_INTERNAL_URL` to reach the gateway.

The Tunnel connector runs in the backend Compose project and reads
`CLOUDFLARE_TUNNEL_TOKEN` from that project's root `.env` file. Do not put the
token in the gateway repository or send it through chat.

The gateway applies basic HTTP limits and routing. Authentication, clinic
isolation, role checks, and business rules remain in the Go API. Keep the API's
existing authentication and rate-limit checks enabled. Add gateway IP-based
rate limiting only after the real client IP forwarding and trusted-proxy
chain are configured for the Cloudflare Tunnel.

---

## 📖 API Documentation (Swagger)

The API is fully documented using Swagger, providing an interactive UI to test endpoints, explore data models, and view required JWT scopes.

* **Interactive UI:** `http://localhost:8080/api/swagger/index.html`
* **Raw JSON (For Front-end Generators like Flutter/Dart):** `http://localhost:8080/api/swagger/doc.json`

---

## ⚖️ License
This project is proprietary software. All rights reserved.
