# Layer 2: Data Handling — Advanced Details

## 1. JSON Serialization Benchmark

### Library Comparison (approximate, measured on typical structs)

| Library | Marshal ns/op | Unmarshal ns/op | allocs/op | Notes |
|---------|-------------|----------------|-----------|-------|
| encoding/json | 1200 | 2400 | 8 | Standard, uses reflection |
| json-iterator | 380 | 720 | 3 | Drop-in replacement, 100% compatible |
| easyjson | 220 | 450 | 1 | Codegen, minimal allocation |
| sonic | 160 | 380 | 2 | JIT, fastest, x86 only |

### When to Use Which
- **encoding/json**: Non-hot paths, simple usage, no extra dependencies.
- **json-iterator**: Hot paths, drop-in replacement without codegen steps.
- **easyjson**: Extreme performance needed, acceptable with a codegen step.
- **sonic**: Absolute maximum speed on amd64/arm64 production environments.

### Setting up easyjson
```bash
go install github.com/mailru/easyjson/...@latest

# Run codegen for DTOs:
easyjson -all internal/dto/user_dto.go
# Generates: internal/dto/user_dto_easyjson.go
# Automatically implements MarshalJSON / UnmarshalJSON
```

---

## 2. String Optimization

### strings.Builder vs Alternatives
```go
// Benchmark: 1000 strings, 10 chars each

// ❌ string += : 220,000 ns/op, 520,000 B/op
result := ""
for _, s := range items { result += s }

// ✅ strings.Builder: 4,200 ns/op, 8,192 B/op (-98%)
var b strings.Builder
b.Grow(estimatedLen) // Important: pre-allocate capacity
for _, s := range items { b.WriteString(s) }
result := b.String()

// ✅ strings.Join: 4,100 ns/op, comparable to Builder
result = strings.Join(items, "")

// ✅ bytes.Buffer: use when needing frequent Write([]byte)
var buf bytes.Buffer
buf.Grow(estimatedLen)
for _, b := range byteSlices { buf.Write(b) }
```

### Format String Optimization
```go
// ❌ Slow: fmt.Sprintf causes many allocations
key := fmt.Sprintf("user:%d", id)

// ✅ Faster: strconv
key := "user:" + strconv.FormatUint(uint64(id), 10)

// ✅ Even faster for repeated use: pre-build prefix
const keyPrefix = "user:"
var buf [20]byte // stack allocation
key := string(strconv.AppendUint(buf[:0], uint64(id), 10))
key = keyPrefix + key
```

---

## 3. Slice Optimization

```go
// ❌ Unknown capacity: causes multiple re-allocations
var results []UserDTO
for _, u := range users {
    results = append(results, toDTO(u))
}

// ✅ Known capacity: 1 allocation
results := make([]UserDTO, 0, len(users))
for _, u := range users {
    results = append(results, toDTO(u))
}

// ✅ Direct indexing: no append needed
results := make([]UserDTO, len(users))
for i, u := range users {
    results[i] = toDTO(u)
}

// When filtering (unknown final size):
results := make([]UserDTO, 0, len(users)/2) // estimate e.g. 50%
for _, u := range users {
    if u.Active {
        results = append(results, toDTO(u))
    }
}
```

---

## 4. Context Timeout Strategy

```go
// Timeout budget — divide total timeout across layers
func (h *Handler) CreateOrder(c *gin.Context) {
    // Total budget: 5s
    ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
    defer cancel()

    // Service called with existing context (carrying its deadline)
    // → DB call auto-timeouts if total exceeds 5s
    // → No need to set separate timeouts for every downstream call
    order, err := h.orderSvc.Create(ctx, req)
    ...
}

// Downstream layers do NOT create new timeouts, they just propagate ctx:
func (s *OrderService) Create(ctx context.Context, req CreateOrderReq) (*Order, error) {
    // Check if context is still valid before starting expensive work:
    if err := ctx.Err(); err != nil {
        return nil, err // Client already canceled, stop processing
    }

    user, err := s.userRepo.FindByID(ctx, req.UserID)
    ...
    order, err := s.orderRepo.Create(ctx, &Order{...})
    ...
}
```

---

## 5. Advanced Concurrency Patterns

### Pipeline with Backpressure
```go
func processItems(ctx context.Context, items []Item) ([]Result, error) {
    // Buffered channels provide natural backpressure
    itemCh := make(chan Item, 10)
    resultCh := make(chan Result, 10)
    errCh := make(chan error, 1)

    // Producer
    go func() {
        defer close(itemCh)
        for _, item := range items {
            select {
            case itemCh <- item:
            case <-ctx.Done():
                return
            }
        }
    }()

    // Workers (fan-out)
    var wg sync.WaitGroup
    for i := 0; i < 5; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            for item := range itemCh {
                result, err := process(ctx, item)
                if err != nil {
                    select {
                    case errCh <- err:
                    default:
                    }
                    return
                }
                select {
                case resultCh <- result:
                case <-ctx.Done():
                    return
                }
            }
        }()
    }

    // Closer
    go func() { wg.Wait(); close(resultCh) }()

    // Collector
    var results []Result
    for result := range resultCh {
        results = append(results, result)
    }

    select {
    case err := <-errCh:
        return nil, err
    default:
        return results, nil
    }
}
```

### Rate Limiter
```go
import "golang.org/x/time/rate"

// 100 req/s, burst max 10
limiter := rate.NewLimiter(100, 10)

func callExternalAPI(ctx context.Context) error {
    if err := limiter.Wait(ctx); err != nil {
        return fmt.Errorf("rate limit wait: %w", err)
    }
    return doCall()
}
```

### Retry with Exponential Backoff
```go
func withRetry(ctx context.Context, maxAttempts int, fn func() error) error {
    backoff := 100 * time.Millisecond
    for attempt := 0; attempt < maxAttempts; attempt++ {
        if err := fn(); err == nil {
            return nil
        }
        select {
        case <-ctx.Done(): return ctx.Err()
        case <-time.After(backoff):
            if backoff < 30*time.Second { backoff *= 2 }
        }
    }
    return fmt.Errorf("failed after %d attempts", maxAttempts)
}
```