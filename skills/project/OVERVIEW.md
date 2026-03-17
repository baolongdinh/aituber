# Project Overview Skill

## TRIGGER
Read this file FIRST when starting any task in this project to understand overall context.

---

## CONTEXT

This is a **DApp** (Decentralized Application) running on **AIOZ Network** combining Web3 and Web2:

- **Backend**: RESTful API written in Golang
- **Frontend**: SPA written in Vue 3
- **Blockchain**: AIOZ Network (EVM-compatible), supports both Mainnet and Testnet
- **Wallet**: Reown AppKit (formerly WalletConnect Web3Modal)
- **Database**: PostgreSQL

---

## TECH STACK

### Backend
| Layer       | Technology              |
|-------------|-------------------------|
| Language    | Go 1.22+                |
| Framework   | Gin (HTTP router)       |
| ORM         | GORM                    |
| Database    | PostgreSQL              |
| Auth        | JWT (golang-jwt/jwt/v5) |
| Cache       | Redis (nếu cần)         |
| Config      | godotenv / viper        |

### Frontend
| Layer         | Technology              |
|---------------|-------------------------|
| Framework     | Vue 3 (Composition API) |
| Build tool    | Vite                    |
| Styling       | Tailwind CSS            |
| State         | Pinia                   |
| HTTP Client   | Axios                   |
| Router        | Vue Router 4            |
| Web3          | ethers.js v6            |
| Wallet UI     | Reown AppKit            |

### Blockchain
| Property      | Mainnet                              | Testnet                                   |
|---------------|--------------------------------------|-------------------------------------------|
| Network       | AIOZ Network                         | AIOZ Network Testnet                      |
| Chain ID      | `168`                                | `4102`                                    |
| RPC URL       | `https://eth-dataseed.aioz.network`  | `https://eth-ds.testnet.aioz.network`     |
| Explorer      | `https://explorer.aioz.network`      | `https://testnet.explorer.aioz.network`   |
| Native Token  | AIOZ (18 decimals)                   | AIOZ (18 decimals)                        |

---

## OVERALL ARCHITECTURE

```
[Browser / Wallet]
       │
       ▼
[Vue 3 Frontend]  ──── Web3 ────▶  [Smart Contracts on Chain]
       │
       │  REST API (JSON)
       ▼
[Go Gin Backend]
       │
       ├──▶ [PostgreSQL]
       └──▶ [Redis Cache]
```

---

## RELATED SKILLS
- Backend patterns → `.antigravity/skills/backend/SKILL.md`
- Frontend patterns → `.antigravity/skills/frontend/SKILL.md`
- DApp / Web3 patterns → `.antigravity/skills/dapp/SKILL.md`
- Coding conventions → `skills/project/CONVENTIONS.md`
- Architecture detail → `skills/project/ARCHITECTURE.md`
- Performance Patterns → `skills/backend/performance/SKILL.md`
