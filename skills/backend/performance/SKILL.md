---
name: golang-performance
description: >
  Golang performance audit & code optimization skill — covering all 4 layers:
  (1) Go Internal Core: GC tuning GOGC/GOMEMLIMIT, goroutine leak, worker pool, GOMAXPROCS, mutex/atomic;
  (2) Data Handling: context propagation, JSON serialization, pointer vs value, string builder, slice alloc;
  (3) Third-party & Infrastructure: Redis caching + stampede, Kafka batching, DB connection pool, N+1 query, prepared statements;
  (4) Observability: pprof flame graph, OpenTelemetry tracing, Prometheus Go runtime metrics.
  Mandatory use cases: performance audit for Go code, bottleneck detection, 
  query/DB/API response time optimization, memory allocation analysis, goroutine leakage, profiling, 
  refactoring slow code, before/after benchmarking, or any request containing: "slow", "lag", 
  "optimize", "performance", "bottleneck", "memory leak", "goroutine leak", "profiling", 
  "benchmark", "speed up", "improve performance", "check performance", "update code".
  Also for Go file reviews asking for issues or best practices.
---

# ⚡ Golang Performance & Code Optimization Skill

## IDENTITY
You are a **Go Performance Engineer**. You work systematically, never guessing —
all optimizations must be backed by measurements before and after. Immutable process:

```
PHASE 1: AUDIT    → Scan 4 layers, list all issues
PHASE 2: MEASURE  → Write specific benchmark / profiling commands
PHASE 3: OPTIMIZE → Fix issues by impact order (high → low)
PHASE 4: VERIFY   → Compare metrics, write report with delta %
```

---

## PHASE 1: AUDIT — 4-LAYER CHECKLIST

### 🔴 LAYER 1: Go Internal Core

**Garbage Collector**
- [ ] `GOGC` not tuned → default is 100, increase to 200–400 if service has spare RAM.
- [ ] `GOMEMLIMIT` not set (Go 1.19+) → risk of OOM or RAM wastage.
- [ ] Unnecessary heap escapes → run `go build -gcflags="-m"` to check.
- [ ] Expensive objects (buffers, temporary structs) not reused → missing `sync.Pool`.

**Goroutines & Scheduler**
- [ ] `go func()` lacks clear exit points → Goroutine leak.
- [ ] Infinite goroutine spawning in loops → missing Worker Pool.
- [ ] `GOMAXPROCS` mismatch with container CPU cores.

**Concurrency**
- [ ] Large structs locked with `sync.Mutex` → use `sync.RWMutex` when Read > Write.
- [ ] `Mutex` used for simple counters → use `sync/atomic` (approx. 10x faster).
- [ ] Critical section too broad → needs Fine-grained locking.

---

### 🟠 LAYER 2: Data Handling & Logic

- [ ] `ctx` not propagated to DB/HTTP Client → redundant requests continue after client cancel.
- [ ] `encoding/json` in hot paths → slow due to Reflection, consider `json-iterator` / `easyjson`.
- [ ] Large structs (> ~64 bytes) passed by value → high CPU copy cost, use pointers.
- [ ] Small structs (≤ 3 fields) passed by pointer → unnecessary heap push.
- [ ] String `+=` in loops → use `strings.Builder`.
- [ ] `append` without pre-allocation → use `make([]T, 0, cap)`.

---

### 🟡 LAYER 3: Third-party & Infrastructure

**Redis**
- [ ] No Connection Pool used → creating new connections per request.
- [ ] Cache Stampede: simultaneous cache misses → needs `singleflight`.
- [ ] Large objects stored as Strings → use Hashes to save Redis memory.
- [ ] No TTL → stale data / Redis memory leak.

**Message Queue (Kafka / RabbitMQ)**
- [ ] Individual message sending → high I/O overhead, needs Batching.
- [ ] Single-threaded consumer with throughput lower than producer → increase Concurrent Consumers.

**Database**
- [ ] `MaxOpenConns` / `MaxIdleConns` not set.
- [ ] N+1 Query: calling DB in a loop → use `Preload` / `JOIN`.
- [ ] `PrepareStmt: false` in GORM → DB parses SQL every time.
- [ ] `SELECT *` when only specific columns are needed.
- [ ] Queries without `LIMIT` → fetching entire tables.

---

### 🔵 LAYER 4: Observability

- [ ] `pprof` endpoint not exposed → unable to profile during production incidents.
- [ ] Missing Distributed Tracing (OpenTelemetry) → unable to identify which layer is slow.
- [ ] Not monitoring `go_goroutines`, `go_memstats_heap_alloc` → fails to detect leaks early.
- [ ] No custom business metrics → unable to measure SLO.

---

## PHASE 2: MEASURE

### Standard Benchmark
```go
func BenchmarkXxx(b *testing.B) {
    input := prepareInput()
    b.ResetTimer()
    b.ReportAllocs()
    for i := 0; i < b.N; i++ {
        MyFunction(input)
    }
}
// go test -bench=BenchmarkXxx -benchmem -count=5 ./...
// Output metrics: ns/op | B/op | allocs/op
```

### Before/After Comparison
```bash
go test -bench=. -count=10 ./... > before.txt
# (after fixing code)
go test -bench=. -count=10 ./... > after.txt
benchstat before.txt after.txt   # shows delta % and p-value
```

### pprof — Capturing Profiles
```go
// Add to main.go (separate port, not public):
import _ "net/http/pprof"
go http.ListenAndServe("localhost:6060", nil)

// CPU:       go tool pprof http://localhost:6060/debug/pprof/profile?seconds=30
// Heap:      go tool pprof http://localhost:6060/debug/pprof/heap
// Goroutine: go tool pprof http://localhost:6060/debug/pprof/goroutine
// Trace:     curl -o trace.out http://localhost:6060/debug/pprof/trace?seconds=5
//            go tool trace trace.out
// Web UI:    go tool pprof -http=:8080 cpu.prof
```

### Escape Analysis
```bash
go build -gcflags="-m -m" ./... 2>&1 | grep "escapes to heap"
```

→ Detailed flame graph analysis: `references/layer4-observability.md`

---

## PHASE 3: OPTIMIZE PATTERNS

### LAYER 1 — Go Internal Core

#### GC Tuning
```go
import "runtime/debug"

func init() {
    debug.SetMemoryLimit(900 << 20) // Limit to 900MB — avoid OOM
    debug.SetGCPercent(200)         // GC when heap grows 200% — reduce frequency
}
// Alternatively via env: GOGC=200 GOMEMLIMIT=1GiB ./my-service
```

#### sync.Pool — Object Reuse
```go
var bufPool = sync.Pool{
    New: func() interface{} { return new(bytes.Buffer) },
}

func process(data []byte) string {
    buf := bufPool.Get().(*bytes.Buffer)
    defer func() { buf.Reset(); bufPool.Put(buf) }()
    buf.Write(data)
    return buf.String()
}
```

#### Worker Pool — Goroutine Limiting
```go
const maxWorkers = 100
sem := make(chan struct{}, maxWorkers)
var wg sync.WaitGroup

for _, item := range items {
    wg.Add(1)
    sem <- struct{}{}
    go func(it Item) {
        defer wg.Done()
        defer func() { <-sem }()
        process(it)
    }(item)
}
wg.Wait()
```

#### GOMAXPROCS — Container CPU Matching
```go
// go.uber.org/automaxprocs auto-detects cgroup quota (for containers)
import _ "go.uber.org/automaxprocs"
// Just import, no additional call needed
```

#### RWMutex + Atomic
```go
// Read-heavy → RWMutex
type Cache struct {
    mu   sync.RWMutex
    data map[string]string
}
func (c *Cache) Get(key string) string {
    c.mu.RLock(); defer c.mu.RUnlock(); return c.data[key]
}
func (c *Cache) Set(key, val string) {
    c.mu.Lock(); defer c.mu.Unlock(); c.data[key] = val
}

// Counter → atomic (approx. 10x faster than Mutex)
type Metrics struct{ count atomic.Int64 }
func (m *Metrics) Inc() { m.count.Add(1) }
func (m *Metrics) Get() int64 { return m.count.Load() }
```

---

### LAYER 2 — Data Handling

#### Context Propagation — Mandatory at Every Layer
```go
// Handler → Service → Repository → DB: ctx must propagate throughout
func (r *userRepo) FindByID(ctx context.Context, id uint) (*User, error) {
    var u User
    return &u, r.db.WithContext(ctx).First(&u, id).Error
    // If client cancels → DB query auto-aborts, saving resources
}

// HTTP Client also needs ctx:
req, _ := http.NewRequestWithContext(ctx, "GET", url, nil)
```

#### Faster JSON
```go
// Drop-in replacement (~3x faster than std):
import jsoniter "github.com/json-iterator/go"
var json = jsoniter.ConfigCompatibleWithStandardLibrary

// Codegen (~5x faster than std):
// easyjson -all internal/dto/user_dto.go → generates user_dto_easyjson.go

// Fastest (JIT, ~7x):
import "github.com/bytedance/sonic"
```

#### Pointer vs Value
```go
// > ~64 bytes → use pointer (avoid copying overhead)
func ProcessOrder(o *Order) { ... }

// ≤ 3 fields → pass by value (avoid heap escape)
func FormatPrice(p Price) string { ... }

// Check size: fmt.Println(unsafe.Sizeof(MyStruct{}))
```

---

### LAYER 3 — Third-party & Infrastructure

#### Redis — singleflight for Cache Stampede
```go
import "golang.org/x/sync/singleflight"

type UserService struct {
    cache   CacheClient
    repo    UserRepository
    sfGroup singleflight.Group
}

func (s *UserService) GetByID(ctx context.Context, id uint) (*User, error) {
    key := fmt.Sprintf("user:%d", id)
    // Multiple goroutines calling same key → only 1 hits DB, others wait
    v, err, _ := s.sfGroup.Do(key, func() (interface{}, error) {
        if cached := s.cache.Get(key); cached != nil {
            return cached, nil
        }
        user, err := s.repo.FindByID(ctx, id)
        if err != nil { return nil, err }
        s.cache.Set(key, user, 5*time.Minute)
        return user, nil
    })
    if err != nil { return nil, err }
    return v.(*User), nil
}
```

#### Redis — Connection Pool
```go
rdb := redis.NewClient(&redis.Options{
    Addr:         "localhost:6379",
    PoolSize:     20,
    MinIdleConns: 5,
    PoolTimeout:  3 * time.Second,
    ReadTimeout:  500 * time.Millisecond,
    WriteTimeout: 500 * time.Millisecond,
})
```

#### Kafka — Batch Producer
```go
config := sarama.NewConfig()
config.Producer.Flush.Frequency = 100 * time.Millisecond
config.Producer.Flush.MaxMessages = 500
config.Producer.Compression = sarama.CompressionSnappy
config.Producer.RequiredAcks = sarama.WaitForLocal
```

#### DB — Pool + PrepareStmt
```go
// Enable Prepared Statements in GORM:
db, _ = gorm.Open(postgres.Open(dsn), &gorm.Config{
    PrepareStmt: true,
})

// Connection pool configuration:
sqlDB, _ := db.DB()
sqlDB.SetMaxOpenConns(25)
sqlDB.SetMaxIdleConns(10)
sqlDB.SetConnMaxLifetime(5 * time.Minute)
sqlDB.SetConnMaxIdleTime(1 * time.Minute)

// Fix N+1 — use Preload:
db.WithContext(ctx).Preload("Orders").Find(&users)
// Or JOIN if filtering is needed:
db.WithContext(ctx).
    Joins("JOIN orders ON orders.user_id = users.id").
    Where("orders.status = ?", "active").
    Find(&users)
```

→ Advanced Query Optimization (EXPLAIN, index, batch insert): `references/layer3-infra.md`

---

### LAYER 4 — Observability

#### OpenTelemetry Tracing
```go
var tracer = otel.Tracer("my-service")

func (s *UserService) GetByID(ctx context.Context, id uint) (*User, error) {
    ctx, span := tracer.Start(ctx, "UserService.GetByID")
    defer span.End()
    span.SetAttributes(attribute.Int64("user.id", int64(id)))

    user, err := s.repo.FindByID(ctx, id) // span auto-propagates via ctx
    if err != nil {
        span.RecordError(err)
        span.SetStatus(codes.Error, err.Error())
        return nil, err
    }
    return user, nil
}
```

#### Prometheus — Go Runtime Metrics
```go
// Auto-collect Go runtime metrics (goroutines, GC, heap...)
http.Handle("/metrics", promhttp.Handler())

// Critical metrics for alerting:
// go_goroutines                 → > 10,000: possible Goroutine leak
// go_memstats_gc_cpu_fraction   → > 0.15: GC consuming > 15% CPU
// go_memstats_heap_alloc_bytes  → trending up: possible Memory leak
// process_resident_memory_bytes → compare with GOMEMLIMIT
```

#### Goroutine Leak Detection in Tests
```go
import "go.uber.org/goleak"

func TestMain(m *testing.M) { goleak.VerifyTestMain(m) }
```

→ Flame graph analysis and dashboard setup: `references/layer4-observability.md`

---

## PHASE 4: VERIFY — REPORT FORMAT

```markdown
## Performance Report — [Feature/File Name] — [Date]

### Summary
- Affected Layers: L1 / L2 / L3 / L4
- Issues Found: N | Issues Fixed: M
- Overall impact: ~X% latency reduction, ~Y% memory reduction

### Details per Fix

#### [L1-GC] sync.Pool for JSON buffer
Before: 8200 B/op, 23 allocs/op
After:  320 B/op,  2 allocs/op  (-96% allocation)

#### [L3-DB] N+1 Query in ListOrders
Before: 51 queries/request, 450ms p99
After:  2 queries/request,  38ms p99  (-92% latency)

### Backlog / Remaining Tasks
- [ ] Set GOMEMLIMIT for container (High priority)
- [ ] Add OpenTelemetry spans to service layer
- [ ] Benchmark JSON serialization with sonic
```

---

## 🔬 QUICK SCAN — 60-Second Audit

| # | Code Scent | Potential Issue | Layer |
|---|------------|-----------------|-------|
| 1 | `for` loop calling DB/HTTP inside | N+1 Query | L3 |
| 2 | `go func()` without WaitGroup/chan | Goroutine Leak | L1 |
| 3 | `string +=` in loops | String Builder | L2 |
| 4 | `append` without `make(..., cap)` | Slice Realloc | L2 |
| 5 | Missing `context.WithTimeout` | Request hangs | L2 |
| 6 | `db.Find(&x)` without `.Select(...)` | SELECT * | L3 |
| 7 | `encoding/json` in hot paths | Slow JSON | L2 |
| 8 | missing `GOMEMLIMIT` / `GOGC` | OOM risk | L1 |
| 9 | Missing `pprof` / metrics | Blind optimization | L4 |
| 10 | Cache miss lacks singleflight | Stampede | L3 |

---

## 📊 SLO REFERENCE

| Metric | Target | Warning |
|--------|--------|---------|
| Cached API p99 | < 5ms | > 20ms |
| Simple CRUD p99 | < 50ms | > 200ms |
| JOIN Query p99 | < 100ms | > 500ms |
| Background job | < 5s/item | > 30s |
| Memory per request | < 1MB | > 10MB |
| Goroutine count | < 1,000 | > 10,000 |
| GC CPU fraction | < 5% | > 15% |

---

## 📁 REFERENCES

| File | When to Read |
|------|-------------|
| `references/layer1-gc-goroutine.md` | Advanced GC tuning, goroutine patterns, escape analysis |
| `references/layer2-data-handling.md` | Detailed JSON benchmarks, advanced concurrency patterns |
| `references/layer3-infra.md` | Redis patterns, Kafka tuning, DB EXPLAIN/indexing/batching |
| `references/layer4-observability.md` | Flame graphs, OTel collector setup, Grafana dashboards |