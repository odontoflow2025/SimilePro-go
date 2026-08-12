# 🦷 Simile Pro API

**Simile Pro** is a robust, production-ready backend built in Go, designed to streamline dental clinic management. From clinical records to financial auditing, it provides a comprehensive suite of features for modern dental practices.

---

## 🚀 Key Modules

### 🏥 Clinical Management
- **Patients & Records:** Full lifecycle management of patient profiles.
- **Anamnesis & Alerts:** Detailed health history with critical alerts (Allergies, Diabetes, etc.) that override standard views.
- **Treatment Plans:** Creation of complex treatment plans with itemized procedures and cost tracking.
- **Evolution Notes:** chronological tracking of patient progress during sessions.

### 💰 Financial & Accounting
- **Transactions & Invoices:** Complete control over revenue and expenses.
- **Account Chart:** Hierarchical chart of accounts for professional bookkeeping.
- **Cash Flow Dashboard:** Date-filtered insights into the clinic's liquidity.
- **Insurance (Convênios):** Integration with major health insurance providers.

### 🛡️ Security & Auditing
- **JWT Authentication:** Secure access control with user roles.
- **Audit Logs:** Global middleware capturing every sensitive state change in the system.
- **Privacy Protocol:** A unique "Protocolo de Auditoria" system that allows temporary access to patient data originating from other clinics during global searches.

### ☁️ SaaS & Integration
- **Multi-tenant:** SaaS-ready architecture focused on multi-clinic environments.
- **Subscription Management:** Built-in plan levels and status tracking.
- **RNDS Integration:** Ready for Brazil's National Health Data Network integration.

---

## 🛠 Tech Stack
- **Language:** Go (Golang) 1.21+
- **Framework:** Gin Gonic (High-performance HTTP router)
- **ORM:** GORM (PostgreSQL focused)
- **Documentation:** Swagger (Swaggo)
- **Security:** JWT (JSON Web Tokens) & Bcrypt

---

## 📂 Project Structure
```bash
├── cmd/api           # Application entry point (main.go)
├── docs/             # Auto-generated Swagger documentation
├── internal/
│   ├── config        # Environment and configuration logic
│   ├── database      # Database connection & migration scripts
│   ├── handlers      # Business logic / Controllers
│   ├── middleware    # Auth, CORS, Audit & Security guards
│   ├── models        # GORM entities & data structures
│   └── routes        # API Route definitions
└── run.bat           # Quick startup script for Windows
```

---

## ⚙️ Getting Started

### Prerequisites
- Go installed
- PostgreSQL instance running

### Installation
1.  **Clone the repository**
2.  **Configure Environment:** Create a `.env` file (see `internal/config`) with:
    ```env
    DB_HOST=localhost
    DB_USER=postgres
    DB_PASS=yourpassword
    DB_NAME=similepro
    SERVER_PORT=8080
    JWT_SECRET=your_jwt_secret
    ```
3.  **Install dependencies:**
    ```bash
    go mod tidy
    ```
4.  **Run the application:**
    ```bash
    ./run.bat
    ```
    *The database tables will be auto-migrated on startup.*

---

## 📖 API Documentation

The API is fully documented using Swagger. You can explore the interactive UI to test endpoints and view data models.

**Access it at:**  
`http://localhost:8080/api/swagger/index.html`

**Frontend Integration:**  
Full `swagger.json` available at `docs/swagger.json` or via `GET /api/swagger/doc.json`.

## 🔄 CI/CD (Integração Contínua)

Este projeto possui testes automatizados integrados com **GitHub Actions**. O pipeline de CI é acionado automaticamente a cada _push_ ou _pull request_ nas branches principais.

### Pipeline Configurado:
1. **Checkout:** Baixa o código fonte.
2. **Setup Go:** Instala a versão correta do Golang (1.25).
3. **Build:** Compila a API para garantir ausência de erros de sintaxe.
4. **Testes:** Roda os testes unitários e de integração (`go test`).

Você pode conferir as regras de execução no arquivo `.github/workflows/ci.yml`.

---

## ⚖️ License
This project is for internal use. All rights reserved.
