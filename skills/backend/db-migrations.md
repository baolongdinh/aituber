# DB Migrations Skill

## TRIGGER
Read this file when you need to create or edit database migrations.

---

## MIGRATION CONVENTION

- Files placed in `migrations/` folder
- Naming: `NNN_description.sql` (NNN = 3 digits, increasing)
- Each migration file includes both **UP** and **DOWN** (if using migration tool)
- Only use **raw SQL** — don't use GORM AutoMigrate in production

```
migrations/
├── 001_create_users.sql
├── 002_create_nfts.sql
├── 003_create_transactions.sql
└── 004_add_nft_metadata.sql
```

---

## MIGRATION FILE TEMPLATE

```sql
-- migrations/001_create_users.sql
-- UP

CREATE TABLE IF NOT EXISTS users (
    id            BIGSERIAL PRIMARY KEY,
    wallet_address VARCHAR(42)  NOT NULL UNIQUE,
    username      VARCHAR(50),
    email         VARCHAR(255) UNIQUE,
    avatar_url    TEXT,
    created_at    TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    deleted_at    TIMESTAMPTZ              -- soft delete (GORM)
);

CREATE INDEX idx_users_wallet_address ON users(wallet_address);
CREATE INDEX idx_users_deleted_at     ON users(deleted_at);
```

---

## COMMON COLUMN PATTERNS

```sql
-- Standard GORM Model columns (LUÔN có ở mỗi bảng)
id         BIGSERIAL PRIMARY KEY,
created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
deleted_at TIMESTAMPTZ,                       -- NULL = not deleted

-- Foreign key
owner_id   BIGINT NOT NULL REFERENCES users(id),
nft_id     BIGINT NOT NULL REFERENCES nfts(id),

-- Money / crypto amounts: dùng NUMERIC, không dùng FLOAT
price      NUMERIC(36, 18) NOT NULL DEFAULT 0,

-- Blockchain data
token_id        VARCHAR(78),    -- uint256 as string
contract_address VARCHAR(42),
tx_hash         VARCHAR(66),

-- ENUM pattern
status  VARCHAR(20) NOT NULL DEFAULT 'pending'
        CHECK (status IN ('pending', 'active', 'sold', 'burned')),
```

---

## DO / DON'T

✅ **DO**
- Luôn dùng `IF NOT EXISTS` khi CREATE TABLE
- Tạo index cho các cột thường dùng trong WHERE / JOIN
- Dùng `TIMESTAMPTZ` (timezone-aware) cho tất cả timestamp
- Dùng `NUMERIC` thay `FLOAT` cho số tiền / token amount
- Dùng `soft delete` (deleted_at) thay `hard delete`

❌ **DON'T**
- KHÔNG dùng `GORM AutoMigrate` trong production
- KHÔNG DROP COLUMN / TABLE mà không có backup plan
- KHÔNG dùng `FLOAT` cho giá trị tài chính
- KHÔNG quên thêm foreign key index
