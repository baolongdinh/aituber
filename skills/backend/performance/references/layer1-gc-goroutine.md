# Layer 1: Go Internal Core — Advanced Details

## 1. GC Tuning — GOGC & GOMEMLIMIT

### How it Works
- `GOGC=100` (default): GC triggers when heap grows by 100% since the last GC.
- `GOGC=200`: GC triggers when heap grows by 200% → less frequent GC, higher throughput, higher RAM usage.
- `GOMEMLIMIT`: Set a hard cap on process memory usage → GC will run more aggressively as memory usage approaches this limit.

### Tuning Recommendations
```
If service has RAM = 4GB, targeted usage 80% = 3.2GB
→ GOMEMLIMIT=3200MiB
→ GOGC=200 (if latency is prioritized over RAM)
→ GOGC=50  (if RAM is constrained, accepting more frequent GC)
```

### Measuring GC Overhead
```go
import "runtime"

var stats runtime.MemStats
runtime.ReadMemStats(&stats)

fmt.Printf("GC runs: %d\n", stats.NumGC)
fmt.Printf("GC pause total: %v\n", time.Duration(stats.PauseTotalNs))
fmt.Printf("Last GC pause: %v\n", time.Duration(stats.PauseNs[(stats.NumGC+255)%256]))
fmt.Printf("GC CPU fraction: %.2f%%\n", stats.GCCPUFraction*100)
// GCCPUFraction > 0.15 → GC is consuming > 15% CPU → needs tuning
```

### GC Tracing
```bash
GODEBUG=gctrace=1 ./my-service
# Output per GC cycle:
# gc 14 @5.855s 0%: 0.006+0.95+0.003 ms clock, ...
#                ↑             ↑
#             %CPU         STW pause
```

---

## 2. Escape Analysis — Keep Variables on the Stack

### Common Patterns causing Heap Escape

```go
// 1. Returning pointer to local variable → escape
func newUser() *User {
    u := User{Name: "Alice"} // u escapes to heap
    return &u
}
// → OK if intentional, but be aware of the cost.

// 2. Interface boxing → escape
var err error = MyError{msg: "fail"} // MyError escapes to heap
// → Use concrete types in hot paths if possible.

// 3. Closure capture → escape
x := 10
go func() { fmt.Println(x) }() // x escapes to heap
// → Pass x as an argument instead of capturing.

// 4. Slices larger than threshold → escape
buf := make([]byte, 65537) // > 64KB → escapes to heap
// → Use sync.Pool for reuse.

// 5. fmt.Sprintf → usually escapes
s := fmt.Sprintf("user:%d", id) // allocation!
// → Use strconv.AppendInt or strings.Builder.
```

### Reading -gcflags="-m" Output
```
./service.go:45:12: &User{...} escapes to heap
./service.go:78:14: x escapes to heap
./service.go:92:23: make([]byte, n) escapes to heap  ← dynamic size
./service.go:105:5: inlined call to ...
```

---

## 3. sync.Pool — Advanced Usage

```go
// Pools for different object types
type ObjectPools struct {
    smallBuf sync.Pool  // 4KB buffers
    largeBuf sync.Pool  // 64KB buffers
    reqCtx   sync.Pool  // RequestContext struct
}

var pools = &ObjectPools{
    smallBuf: sync.Pool{New: func() interface{} { return make([]byte, 4096) }},
    largeBuf: sync.Pool{New: func() interface{} { return make([]byte, 65536) }},
    reqCtx:   sync.Pool{New: func() interface{} { return &RequestContext{} }},
}

// Important: sync.Pool is cleared by GC at any time.
// → Do NOT use for storing vital state.
// → ONLY use for reusable objects (must be reset to zero state).
```

---

## 4. Advanced Goroutine Patterns

### Detecting Leaks with Runtime
```go
func goroutineCount() int {
    return runtime.NumGoroutine()
}

// In tests: check goroutine count before/after
func TestNoLeak(t *testing.T) {
    before := runtime.NumGoroutine()
    
    // Run operation
    doSomething()
    
    time.Sleep(100 * time.Millisecond) // wait for goroutine cleanup
    after := runtime.NumGoroutine()
    
    if after > before+2 {
        t.Errorf("goroutine leak: %d → %d", before, after)
    }
}
```

### Context Cancellation Pattern
```go
func startWorker(ctx context.Context) {
    go func() {
        ticker := time.NewTicker(1 * time.Second)
        defer ticker.Stop()
        
        for {
            select {
            case <-ctx.Done():
                return // ALWAYS have this case
            case <-ticker.C:
                doWork()
            }
        }
    }()
}
```

### ErrGroup — Goroutines with Error Handling
```go
import "golang.org/x/sync/errgroup"

g, ctx := errgroup.WithContext(context.Background())

for _, id := range userIDs {
    id := id // capture
    g.Go(func() error {
        return processUser(ctx, id)
    })
}

// Wait for all to finish, returns the first error encountered (and cancels ctx)
if err := g.Wait(); err != nil {
    return err
}
```

---

## 5. Fine-grained Locking

```go
// ❌ Locking an entire large map
type UserCache struct {
    mu    sync.RWMutex
    users map[uint]*User
}

// ✅ Sharded map — divided into N shards, each with its own lock
const shardCount = 32

type ShardedCache struct {
    shards [shardCount]struct {
        sync.RWMutex
        data map[uint]*User
    }
}

func (c *ShardedCache) shard(id uint) int {
    return int(id) % shardCount
}

func (c *ShardedCache) Get(id uint) *User {
    s := c.shard(id)
    c.shards[s].RLock()
    defer c.shards[s].RUnlock()
    return c.shards[s].data[id]
}
```

---

## 6. GOMAXPROCS in Containers

```go
// Issue: container cgroup limits CPU, but runtime.NumCPU() returns host CPU count.
// Solution: use automaxprocs.

import _ "go.uber.org/automaxprocs"
// Automatically detects cgroup quota and sets GOMAXPROCS correctly.

// Or manual:
import "runtime"
import "github.com/uber-go/automaxprocs/maxprocs"

func main() {
    undo, err := maxprocs.Set(maxprocs.Logger(log.Printf))
    defer undo()
    ...
}
```