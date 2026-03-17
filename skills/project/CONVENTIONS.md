# Coding Conventions Skill

## TRIGGER
Read this file when you need to know naming rules, file structure, or coding style.

---

## BACKEND (Go) CONVENTIONS

### File Naming
| File Type        | Convention          | Example                   |
|------------------|---------------------|---------------------------|
| Go source file   | `snake_case`        | `user_handler.go`         |
| Test file        | `_test.go` suffix   | `user_handler_test.go`    |
| SQL migration    | `NNN_description`   | `001_create_users.sql`    |
| Config file      | `snake_case`        | `config.go`               |

### Type / Variable Naming
| Type               | Convention           | Example                  |
|--------------------|----------------------|--------------------------|
| Struct             | `PascalCase`         | `UserHandler`            |
| Interface          | `PascalCase` + `er`  | `UserRepository`         |
| Function/Method    | `PascalCase` (export)| `GetUserByID`            |
| Private function   | `camelCase`          | `parseToken`             |
| Constants          | `PascalCase`         | `MaxRetryCount`          |
| Env variables      | `UPPER_SNAKE_CASE`   | `DATABASE_URL`           |
| DB table name      | `snake_case` plural  | `user_profiles`          |
| DB column name     | `snake_case`         | `created_at`             |

### Package Naming
- Package name: **lowercase, single word** (no underscores)
- Package `internal/handler` → package `handler`
- Package `internal/service` → package `service`
- Package `internal/repository` → package `repository`

### Error Variable Naming
```go
// Sentinel errors: capitalize first letter, prefix "Err"
var ErrNotFound = errors.New("resource not found")
var ErrUnauthorized = errors.New("unauthorized")
```

---

## FRONTEND (Vue 3 / JS) CONVENTIONS

### File Naming
| File Type          | Convention           | Example                   |
|--------------------|----------------------|---------------------------|
| Vue component      | `PascalCase.vue`     | `NftCard.vue`             |
| Composable         | `camelCase.js`       | `useAuth.js`              |
| Store (Pinia)      | `noun.store.js`      | `auth.store.js`           |
| API module         | `noun.api.js`        | `user.api.js`             |
| Utility            | `camelCase.js`       | `formatters.js`          |
| Router file        | `index.js`           | `router/index.js`         |
| CSS / styles       | `kebab-case.css`     | `main.css`                |

### Vue Component Conventions
```vue
<!-- ALWAYS use Composition API + <script setup> -->
<script setup>
// 1. Imports
// 2. Props & Emits
// 3. Store / Router / Route
// 4. Reactive state
// 5. Computed properties
// 6. Watchers
// 7. Lifecycle hooks
// 8. Methods / handlers
</script>
```

### CSS Class Naming
- Use Tailwind utility classes
- Custom class: `kebab-case` (when need to extract to CSS)
- BEM: DON'T use (Tailwind replaces it)

### Variable Naming (JS)
| Type               | Convention      | Example               |
|--------------------|-----------------|-----------------------|
| Variable           | `camelCase`     | `userProfile`         |
| Constant           | `UPPER_SNAKE`   | `MAX_RETRY`           |
| Boolean            | `is/has/can` prefix | `isLoading`      |
| Async function     | `camelCase`     | `fetchUserData`       |
| Event handler      | `handle` prefix | `handleSubmit`        |
| Composable         | `use` prefix    | `useWallet`           |

---

## GIT CONVENTIONS

### Commit Messages (Conventional Commits)
```
<type>(<scope>): <subject>

Types:
  feat     → New feature
  fix      → Bug fix
  refactor → Refactor without adding feature/fix
  chore    → Maintenance (config, deps...)
  docs     → Documentation
  test     → Test

Examples:
  feat(nft): add mint NFT endpoint
  fix(auth): handle expired JWT token
  refactor(user): extract user validation to service layer
```

### Branch Naming
```
feature/<short-description>    → feature/nft-minting
fix/<short-description>        → fix/jwt-expiry
chore/<short-description>      → chore/update-deps
```
