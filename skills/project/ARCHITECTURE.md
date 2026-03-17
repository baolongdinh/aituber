# Architecture Skill

## TRIGGER
Read this file when you need to understand overall architecture before designing new features or refactoring.

---

## BACKEND N-LAYER ARCHITECTURE

```
HTTP Request
     │
     ▼
┌──────────────┐
│   Middleware  │  cors, auth_jwt, logger, rate_limit
└──────┬───────┘
       │
       ▼
┌──────────────┐
│   Handler    │  Receive request, validate input, call Service, return response
│  (Controller)│  No business logic
└──────┬───────┘
       │
       ▼
┌──────────────┐
│   Service    │  Business logic, orchestrate multiple Repository
│              │  No knowledge of HTTP or specific DB queries
└──────┬───────┘
       │
       ▼
┌──────────────┐
│  Repository  │  Data access layer, only CRUD with DB
│              │  Return model struct
└──────┬───────┘
       │
       ▼
┌──────────────┐
│  PostgreSQL  │
└──────────────┘
```

### Layer separation rules (STRICT)
- **Handler** → only calls **Service**, never calls **Repository** directly
- **Service** → can call multiple **Repository**
- **Repository** → only interacts with **DB (GORM)**
- **DTO** → used for request/response at Handler layer
- **Model** → used for GORM entity, not exposed outside API

---

## FOLDER STRUCTURE (BACKEND)

```
backend/
├── cmd/
│   └── server/
│       └── main.go               # Bootstrap: load config, init DB, start server
├── internal/
│   ├── handler/                  # Layer 1: HTTP handlers (1 file/domain)
│   │   ├── auth_handler.go
│   │   ├── user_handler.go
│   │   └── nft_handler.go
│   ├── service/                  # Layer 2: Business logic (1 file/domain)
│   │   ├── auth_service.go
│   │   ├── user_service.go
│   │   └── nft_service.go
│   ├── repository/               # Layer 3: Data access (1 file/domain)
│   │   ├── user_repository.go
│   │   └── nft_repository.go
│   ├── model/                    # GORM DB models
│   │   ├── user.go
│   │   └── nft.go
│   ├── dto/                      # Request & Response structs
│   │   ├── auth_dto.go
│   │   └── user_dto.go
│   ├── middleware/               # Gin middlewares
│   │   ├── auth_middleware.go
│   │   └── cors_middleware.go
│   └── router/
│       └── router.go             # Route registration
├── pkg/                          # Shared, reusable packages (không phụ thuộc internal)
│   ├── database/
│   │   └── postgres.go           # DB connection pool
│   ├── jwt/
│   │   └── jwt.go                # JWT sign/verify
│   ├── response/
│   │   └── response.go           # Standard API response helpers
│   └── validator/
│       └── validator.go          # Custom validators
├── migrations/                   # SQL migration files (số thứ tự tăng dần)
│   ├── 001_create_users.sql
│   └── 002_create_nfts.sql
├── config/
│   └── config.go                 # Struct config + load từ env
├── .env.example
├── Dockerfile
├── go.mod
└── go.sum
```

---

## FOLDER STRUCTURE (FRONTEND)

```
frontend/
├── src/
│   ├── api/                      # Layer 1: Axios API calls (1 file/domain)
│   │   ├── index.js              # Axios instance + interceptors
│   │   ├── auth.api.js
│   │   └── user.api.js
│   ├── stores/                   # Layer 2: Pinia stores (1 file/domain)
│   │   ├── auth.store.js
│   │   └── nft.store.js
│   ├── composables/              # Layer 3: Business logic hooks
│   │   ├── useAuth.js
│   │   └── useNft.js
│   ├── views/                    # Route-level pages
│   │   ├── HomeView.vue
│   │   ├── auth/
│   │   │   └── LoginView.vue
│   │   └── nft/
│   │       └── NftListView.vue
│   ├── components/               # Reusable UI components
│   │   ├── common/               # AppButton, AppModal, AppToast...
│   │   └── nft/                  # NftCard, NftGrid...
│   ├── layouts/                  # Layout wrappers
│   │   ├── DefaultLayout.vue
│   │   └── AuthLayout.vue
│   ├── router/
│   │   └── index.js              # Vue Router definitions + guards
│   └── utils/
│       ├── formatters.js         # Pure helper functions
│       └── web3.js               # Wallet/contract helpers
├── public/
├── index.html
├── vite.config.js
├── tailwind.config.js
└── .env.example
```

---

## DEPENDENCY INJECTION PATTERN (BACKEND)

Dùng constructor injection, **không** dùng global singletons:

```go
// main.go bootstrap order:
// 1. Load config
// 2. Connect DB
// 3. Init Repositories  (nhận *gorm.DB)
// 4. Init Services      (nhận Repositories)
// 5. Init Handlers      (nhận Services)
// 6. Register Routes    (nhận Handlers)
// 7. Start server
```
