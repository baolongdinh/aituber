# Git Workflow Skill

## ⚠️ MANDATORY CHECK BEFORE COMMIT
Before finalizing any commit, you MUST run the `scripts/commit-check.py` script to ensure the commit message follows Conventional Commits and there is no trailing whitespace in staged files.

**Command:**
```bash
python3 scripts/commit-check.py "<your commit message>"
```

---

## TRIGGER
Read this file and execute git commit **AFTER COMPLETING** any feature, fix, or task.
Antigravity must automatically add + commit without user reminder.

---

## CORRECT COMMIT PROCESS

> **After each completed task, Antigravity WILL**:
> 1. `git add` relevant files
> 2. `git commit` with message following Conventional Commits
> 3. Report commit hash + message to user

---

## CONVENTIONAL COMMITS FORMAT

```
<type>(<scope>): <subject>

subject: short, English, no period at end, max 72 characters
```

### Allowed Types

| Type       | When to use                                      |
|------------|---------------------------------------------------|
| `feat`     | Adding new feature                                |
| `fix`      | Fix bug                                           |
| `refactor` | Refactor code, không thêm feature / sửa bug       |
| `chore`    | Maintenance: config, deps, tooling                |
| `docs`     | Chỉ thay đổi documentation / comments            |
| `test`     | Thêm / sửa test                                   |
| `style`    | CSS, formatting (không ảnh hưởng logic)           |
| `perf`     | Cải thiện performance                             |

### Scope examples

| Where to change   | Suggested scope              |
|-----------------|--------------------------|
| Backend handler    | `auth`, `nft`, `user`    |
| Service / repo     | `auth`, `nft`, `user`    |
| Frontend view      | `nft-list`, `login`      |
| Pinia store        | `auth-store`, `nft-store`|
| Composable         | `use-wallet`, `use-nft`  |
| DB migration       | `migration`              |
| Config / env       | `config`                 |
| Skills / docs      | `skills`, `docs`         |

---

## COMMIT WORKFLOW (STEP BY STEP)

After completing code, follow this sequence:

```bash
# 1. Check what has changed
git status

# 2. Stage relevant files for the task just completed
# Use specific files instead of `git add .` if possible
git add <file1> <file2> ...
# Or if certain all changes belong to same scope:
git add .

# 3. Commit
git commit -m "feat(nft): add mint NFT endpoint with AIOZ payment"
```

---

## COMMIT MESSAGE EXAMPLES

```bash
# New feature
git commit -m "feat(nft): add buy NFT endpoint with AIOZ token support"
git commit -m "feat(auth): implement SIWE login with Reown AppKit"
git commit -m "feat(wallet): add AIOZ mainnet/testnet network switcher"

# Bug fix
git commit -m "fix(auth): handle expired JWT token in middleware"
git commit -m "fix(nft): correct tokenId lookup in purchase flow"

# Refactor
git commit -m "refactor(user): extract profile update logic to service layer"

# Migration
git commit -m "chore(migration): add user_profiles table with email index"

# Config / env
git commit -m "chore(config): add AIOZ testnet chain config"

# Skills / docs
git commit -m "docs(skills): add Git workflow skill"
git commit -m "docs(skills): update dapp skill for Reown AppKit and AIOZ Network"
```

---

## MULTI-FILE COMMIT GROUPING

Nhóm commit theo **logic thay đổi**, không theo số lượng file:

```bash
# ✅ Một commit cho toàn bộ một feature (handler + service + repo + dto)
git add internal/handler/nft_handler.go \
        internal/service/nft_service.go \
        internal/repository/nft_repository.go \
        internal/dto/nft_dto.go
git commit -m "feat(nft): implement NFT listing with pagination"

# ✅ Commit riêng nếu migration độc lập
git add migrations/002_create_nfts.sql
git commit -m "chore(migration): add nfts table"
```

---

## DO / DON'T

✅ **DO**
- Commit ngay sau khi task hoàn thành và code chạy được
- Dùng scope để rõ ràng phạm vi thay đổi
- Message ngắn gọn, đủ hiểu, tiếng Anh
- Commit theo từng logical unit (không commit cả project 1 lần)

❌ **DON'T**
- KHÔNG commit code bị broken / chưa compile
- KHÔNG dùng message chung chung: `update`, `fix`, `changes`, `WIP`
- KHÔNG commit file không liên quan vào cùng commit
- KHÔNG `git add .` nếu thư mục có file nhạy cảm chưa được `.gitignore`
- KHÔNG commit file: `.env`, `*.log`, `go.sum` (nếu chưa có lý do rõ ràng)

---

## .GITIGNORE CHECKLIST

Đảm bảo `.gitignore` có các entries sau trước khi `git add .`:

```gitignore
# Backend
.env
*.log
/backend/tmp/
/backend/vendor/   # nếu không vendor

# Frontend
node_modules/
dist/
.env
.env.local
.env.*.local

# IDE
.vscode/
.idea/
*.swp

# OS
.DS_Store
Thumbs.db
```
