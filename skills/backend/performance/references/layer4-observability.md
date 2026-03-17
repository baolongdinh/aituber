# Layer 4: Observability — Tracking & Profiling

## 1. Zero-Instrumentation Profiling with pprof

The `pprof` tool is essential for identifying bottlenecks in a live system.

### Endpoint Setup
```go
import _ "net/http/pprof"

func main() {
    // Expose on a private port
    go func() {
        log.Println(http.ListenAndServe("localhost:6060", nil))
    }()
    ...
}
```

### Capturing Data
```bash
# Capture 30s CPU profile
go tool pprof http://localhost:6060/debug/pprof/profile?seconds=30

# Capture Heap (memory) profile
go tool pprof http://localhost:6060/debug/pprof/heap

# Capture Goroutine stack traces (detect leaks/deadlocks)
go tool pprof http://localhost:6060/debug/pprof/goroutine
```

### Reading Flame Graphs
Use the `-http` flag to open a browser-based visualization:

```bash
go tool pprof -http=:8080 cpu.prof
```
- **Wide boxes**: Functions that take the most time.
- **Deep stacks**: Deep function call recursion.
- **Red/Hot areas**: Focus optimization here.

---

## 2. Distributed Tracing (OpenTelemetry)

Tracing allows you to see the path of a request across services and layers.

### Creating Spans
```go
var tracer = otel.Tracer("order-service")

func (s *Service) ProcessOrder(ctx context.Context, id uint) error {
    ctx, span := tracer.Start(ctx, "ProcessOrder")
    defer span.End()
    
    // Add metadata for debugging
    span.SetAttributes(attribute.Int("order.id", int(id)))

    if err := s.validate(ctx, id); err != nil {
        span.RecordError(err)
        span.SetStatus(codes.Error, "validation failed")
        return err
    }
    
    return nil
}
```

### Context Propagation
The `ctx` object MUST be passed to every function call, database query, and HTTP request to maintain the trace continuity.

---

## 3. Runtime Metrics (Prometheus)

Monitor the health of the Go runtime itself.

### Exposing Metrics
```go
import "github.com/prometheus/client_golang/prometheus/promhttp"

http.Handle("/metrics", promhttp.Handler())
```

### Key Metrics to Watch
- `go_goroutines`: If trending up without coming down, you have a **Goroutine Leak**.
- `go_memstats_heap_alloc_bytes`: Current memory in use.
- `go_memstats_gc_cpu_fraction`: Percentage of CPU used by the Garbage Collector. If > 10-15%, you are in "GC pressure" → increase GOGC or reduce allocations.
- `go_memstats_heap_sys_bytes`: Total memory requested from the OS.

---

## 4. Execution Tracing (`go tool trace`)

Best for debugging scheduler issues, lock contention, and latency spikes.

```bash
# Capture trace
curl -o trace.out http://localhost:6060/debug/pprof/trace?seconds=5

# View trace (requires Chrome)
go tool trace trace.out
```

---

## 5. Logging for Performance

Logging too much is a performance bottleneck.

### Structured Logging with Zero Allocations
Use `rs/zerolog` or `uber-go/zap` for high-performance logging.

```go
// ❌ Slow: reflection-based
log.Printf("User %d logged in", userID)

// ✅ Fast: structured, type-safe
logger.Info().Int("userID", userID).Msg("User logged in")
```
- **LogLevels**: Use `Debug` for verbose info, `Info` for significant events, `Error` for failures.
- **Avoid logging in hot loops**.