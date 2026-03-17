# Layer 3: Third-party & Infrastructure — Advanced Details

## 1. Redis Performance Patterns

### Avoiding Cache Stampede (Thundering Herd)
When a hot cache key expires, multiple concurrent requests might hit the database at once.

```go
import "golang.org/x/sync/singleflight"

var g singleflight.Group

func fetchWithCache(ctx context.Context, key string) (string, error) {
    // 1. Try Cache
    if val, err := cache.Get(ctx, key); err == nil {
        return val, nil
    }

    // 2. Cache Miss - Use singleflight to ensure only ONE hits the DB
    v, err, _ := g.Do(key, func() (interface{}, error) {
        // Double check cache inside singleflight (Optional but recommended)
        if val, err := cache.Get(ctx, key); err == nil {
            return val, nil
        }
        
        // Fetch from DB
        val, err := db.Fetch(ctx, key)
        if err != nil { return nil, err }
        
        // Set Cache
        cache.Set(ctx, key, val, 5*time.Minute)
        return val, nil
    })
    
    return v.(string), err
}
```

### Saving Redis Memory with Hashes
Instead of storing thousands of small keys like `user:1:name`, `user:1:email`, store them as a Hash `user:1`.

```go
// ❌ Wasteful: many individual keys
rdb.Set(ctx, "user:1:name", "Alice", 0)
rdb.Set(ctx, "user:1:email", "alice@example.com", 0)

// ✅ Efficient: one hash key
rdb.HSet(ctx, "user:1", map[string]interface{}{
    "name":  "Alice",
    "email": "alice@example.com",
})
```

---

## 2. Kafka Tuning & Batching

### High Throughput Batching
Sending messages individually is slow due to I/O overhead.

```go
config := sarama.NewConfig()
// Wait up to 100ms for more messages before sending a batch
config.Producer.Flush.Frequency = 100 * time.Millisecond
// Buffer up to 500 messages per batch
config.Producer.Flush.MaxMessages = 500
// Use Snappy compression (fast, good balance)
config.Producer.Compression = sarama.CompressionSnappy
```

### Concurrent Consumers
Increase throughput by processing messages across multiple goroutines.

```go
func (h *ConsumerHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
    // Semaphore to limit concurrency
    sem := make(chan struct{}, 20)
    
    for message := range claim.Messages() {
        sem <- struct{}{}
        go func(msg *sarama.ConsumerMessage) {
            defer func() { <-sem }()
            process(msg)
            session.MarkMessage(msg, "")
        }(message)
    }
    return nil
}
```

---

## 3. Database: GORM & PostgreSQL Optimization

### Prepared Statements
Enable this to allow Postgres to reuse execution plans, saving parsing time.

```go
db, _ := gorm.Open(postgres.Open(dsn), &gorm.Config{
    PrepareStmt: true, // Speeds up repetitive queries
})
```

### Solving N+1 Queries
```go
// ❌ Slow: 1 query for users + N queries for orders in a loop
var users []User
db.Find(&users)
for i := range users {
    db.Model(&users[i]).Related(&users[i].Orders)
}

// ✅ Fast: 1 query for users + 1 query for ALL related orders
db.Preload("Orders").Find(&users)

// ✅ Fast: JOIN if you need to filter based on relations
db.Joins("JOIN orders ON orders.user_id = users.id").
   Where("orders.status = ?", "pending").
   Find(&users)
```

### Batch Inserts
```go
// ❌ Slow: inserting one by one
for _, user := range users {
    db.Create(&user)
}

// ✅ Fast: bulk insert
db.CreateInBatches(users, 100)
```

### Index Analysis with EXPLAIN
Always check if your query uses an index:

```sql
EXPLAIN ANALYZE SELECT * FROM users WHERE email = 'alice@example.com';

-- Look for "Index Scan" vs "Seq Scan" (Sequential Scan = BAD)
```

---

## 4. Connection Pool Engineering

Incorrect pool sizes can kill performance (too small: waiting; too large: DB overhead).

```go
sqlDB, _ := db.DB()

// Set to slightly more than your peak concurrent requests
sqlDB.SetMaxOpenConns(25) 

// Keep some connections warm
sqlDB.SetMaxIdleConns(10) 

// Recycle connections to prevent leaks and handle DB restarts
sqlDB.SetConnMaxLifetime(5 * time.Minute) 
```