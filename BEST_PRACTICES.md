# Malice Go Best Practices & Coding Standards

## 📋 فهرست
1. [Error Handling](#error-handling)
2. [Concurrency](#concurrency)
3. [Resource Management](#resource-management)
4. [Logging](#logging)
5. [Testing](#testing)
6. [API Design](#api-design)
7. [Security](#security)
8. [Performance](#performance)

---

## Error Handling

### ✅ DO: Return errors, not log.Fatal in libraries

```go
// ✅ GOOD - Library function
func ParseConfig(path string) (*Config, error) {
    data, err := ioutil.ReadFile(path)
    if err != nil {
        return nil, fmt.Errorf("read config: %w", err)
    }
    // ...
    return cfg, nil
}

// ✅ GOOD - Main application
func main() {
    cfg, err := ParseConfig("config.toml")
    if err != nil {
        log.Fatalf("failed to load config: %v", err)
    }
    // ...
}
```

### ✅ DO: Wrap errors with context

```go
// ✅ GOOD - Context is clear
if err := database.Query(ctx, sql); err != nil {
    return fmt.Errorf("scan %s: query failed: %w", file.SHA256, err)
}

// ❌ BAD - Lost context
if err := database.Query(ctx, sql); err != nil {
    return err
}
```

### ✅ DO: Use custom error types for specific cases

```go
// ✅ GOOD
type ValidationError struct {
    Field   string
    Message string
}

func (e *ValidationError) Error() string {
    return fmt.Sprintf("validation failed on %q: %s", e.Field, e.Message)
}

// Usage
if len(path) == 0 {
    return &ValidationError{"path", "must not be empty"}
}

// Client can check type
if err != nil {
    if ve, ok := err.(*ValidationError); ok {
        // Handle validation error specifically
        log.Warnf("validation error: %s", ve.Field)
    }
}
```

### ✅ DO: Never ignore errors unless explicit

```go
// ✅ GOOD - Explicit ignore
err = os.Chmod(dst, fi.Mode())
if err != nil {
    return err  // Not ignoring
}

// ✅ GOOD - Explicitly documented ignore
_ = os.Remove(tmpFile)  // Ignore error, cleanup best-effort
```

### ❌ DON'T: Use boolean returns for errors

```go
// ❌ BAD - Confusing semantics
func CheckError(err error) bool {
    if err != nil {
        log.Error(err)
        return false  // Returns false on error? Confusing!
    }
    return true  // Returns true on success
}

// Usage is confusing
if !CheckError(err) {  // Double negative!
    return err
}

// ✅ GOOD - Clear semantics
func CheckError(err error) error {
    if err != nil {
        log.WithError(err).Error("operation failed")
        return err
    }
    return nil
}

// Usage is clear
if err = CheckError(err); err != nil {
    return err
}
```

---

## Concurrency

### ✅ DO: Always use context for goroutines

```go
// ✅ GOOD
func (s *Scanner) Scan(ctx context.Context, file string) error {
    ctx, cancel := context.WithTimeout(ctx, 5*time.Minute)
    defer cancel()
    
    var wg sync.WaitGroup
    errors := make(chan error, len(s.plugins))
    defer close(errors)
    
    for _, plugin := range s.plugins {
        wg.Add(1)
        go func(p Plugin) {
            defer wg.Done()
            if err := p.Run(ctx); err != nil {
                errors <- fmt.Errorf("plugin %s: %w", p.Name, err)
            }
        }(plugin)
    }
    
    // Wait with timeout handling
    done := make(chan struct{})
    go func() {
        wg.Wait()
        close(done)
    }()
    
    select {
    case <-done:
        return nil
    case <-ctx.Done():
        return ctx.Err()
    }
}
```

### ✅ DO: Defer cleanup in goroutine context

```go
// ✅ GOOD
go func() {
    defer func() {
        if r := recover(); r != nil {
            log.Errorf("goroutine panic: %v", r)
        }
    }()
    
    if err := doWork(); err != nil {
        log.WithError(err).Error("work failed")
    }
}()
```

### ✅ DO: Handle channel close properly

```go
// ✅ GOOD - Explicit close and drain
ch := make(chan Result)
defer close(ch)

go func() {
    defer close(ch)
    for _, item := range items {
        ch <- process(item)
    }
}()

for result := range ch {  // Safe - waits for close
    handle(result)
}
```

### ❌ DON'T: Ignore WaitGroup.Done()

```go
// ❌ BAD - May deadlock
var wg sync.WaitGroup
wg.Add(1)
go func() {
    // Missing: wg.Done()
}()
wg.Wait()  // Deadlock!

// ✅ GOOD
var wg sync.WaitGroup
wg.Add(1)
go func() {
    defer wg.Done()
    // work
}()
wg.Wait()
```

### ❌ DON'T: Create unbounded channels

```go
// ❌ BAD - Memory leak possible
ch := make(chan error)  // No buffer, sender blocks forever if receiver fails
go func() {
    ch <- err  // Blocks if nothing reading
}()

// ✅ GOOD
ch := make(chan error, 1)  // Buffered for single error
go func() {
    ch <- err  // Non-blocking
}()
err := <-ch  // Receive
```

---

## Resource Management

### ✅ DO: Always defer cleanup

```go
// ✅ GOOD
func processFile(path string) error {
    file, err := os.Open(path)
    if err != nil {
        return fmt.Errorf("open file: %w", err)
    }
    defer file.Close()
    
    scanner := bufio.NewScanner(file)
    for scanner.Scan() {
        // Process line
    }
    return scanner.Err()
}
```

### ✅ DO: Check defer cleanup errors

```go
// ✅ GOOD
func writeConfig(path string, cfg *Config) error {
    file, err := os.Create(path)
    if err != nil {
        return fmt.Errorf("create file: %w", err)
    }
    defer func() {
        if err := file.Close(); err != nil {
            log.WithError(err).Warn("close file")
        }
    }()
    
    return json.NewEncoder(file).Encode(cfg)
}
```

### ✅ DO: Use sync.Once for singleton initialization

```go
// ✅ GOOD
var (
    db     *Database
    dbOnce sync.Once
    dbErr  error
)

func GetDatabase() (*Database, error) {
    dbOnce.Do(func() {
        db, dbErr = initializeDatabase()
    })
    return db, dbErr
}
```

### ❌ DON'T: Leak file descriptors

```go
// ❌ BAD
for _, file := range files {
    f, _ := os.Open(file)
    data, _ := ioutil.ReadAll(f)  // Never closed!
    process(data)
}

// ✅ GOOD
for _, file := range files {
    func() {
        f, err := os.Open(file)
        if err != nil {
            return
        }
        defer f.Close()
        
        data, err := ioutil.ReadAll(f)
        if err != nil {
            return
        }
        process(data)
    }()
}
```

---

## Logging

### ✅ DO: Use structured logging

```go
// ✅ GOOD
log.WithFields(log.Fields{
    "file":   file.SHA256,
    "size":   file.Size,
    "plugins": len(plugins),
    "duration": elapsed,
}).Info("scan completed")

// ✅ GOOD - For errors
log.WithFields(log.Fields{
    "file": file,
    "error": err,
}).Error("scan failed")

// But better:
log.WithError(err).WithField("file", file).Error("scan failed")
```

### ✅ DO: Use appropriate log levels

```go
// ✅ GOOD
log.Debug("checking plugin configuration")      // Development only
log.Info("started database connection")          // Important info
log.Warn("elasticsearch slow (>1s)")             // Potential issue
log.Error("plugin startup failed")               // Error occurred
log.Fatal("configuration missing, cannot start") // Exit
```

### ✅ DO: Add context to log messages

```go
// ✅ GOOD - Clear what failed and why
log.WithFields(log.Fields{
    "stage": "scanning",
    "file": file,
    "scan_id": scanID,
}).WithError(err).Error("plugin failed")

// ❌ BAD - Vague
log.Error("error:", err)
```

### ❌ DON'T: Use log.Println for important messages

```go
// ❌ BAD
log.Println("event:", event)  // Lost context, no timestamp

// ✅ GOOD
log.WithField("type", "file_created").
    WithField("path", event.Name).
    Debug("file system event")
```

---

## Testing

### ✅ DO: Write table-driven tests

```go
// ✅ GOOD
func TestValidatePath(t *testing.T) {
    tests := []struct {
        name      string
        path      string
        wantError bool
        errType   string
    }{
        {"valid file", "/tmp/sample.bin", false, ""},
        {"nonexistent", "/tmp/nonexist", true, "NotExist"},
        {"empty path", "", true, "validation"},
        {"path traversal", "../../etc/passwd", true, "traversal"},
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := ValidatePath(tt.path)
            if (err != nil) != tt.wantError {
                t.Errorf("got error %v, want error %v", err != nil, tt.wantError)
            }
        })
    }
}
```

### ✅ DO: Mock external dependencies

```go
// ✅ GOOD - Use interface for mocking
type DockerClient interface {
    ContainerStart(ctx context.Context, id string, options types.ContainerStartOptions) error
}

type MockDocker struct {
    startCalls int
    startErr   error
}

func (m *MockDocker) ContainerStart(ctx context.Context, id string, options types.ContainerStartOptions) error {
    m.startCalls++
    return m.startErr
}

func TestPluginStartWithDockerError(t *testing.T) {
    mock := &MockDocker{startErr: errors.New("docker failed")}
    plugin := NewPlugin(mock)
    
    err := plugin.Start(context.Background())
    if err == nil {
        t.Error("expected error")
    }
    if mock.startCalls != 1 {
        t.Errorf("expected 1 docker call, got %d", mock.startCalls)
    }
}
```

---

## API Design

### ✅ DO: Accept context as first parameter

```go
// ✅ GOOD - Consistent with stdlib
func (s *Scanner) Scan(ctx context.Context, file string) error {
    // Can set timeout
}

func (db *Database) Query(ctx context.Context, sql string) ([]Row, error) {
    // Can cancel mid-query
}
```

### ✅ DO: Return error as last return value

```go
// ✅ GOOD
func (p *Plugin) Run(ctx context.Context, sample string) (*Result, error) {
    // ...
}

// Usage is natural
result, err := plugin.Run(ctx, sample)
if err != nil {
    return err
}
```

### ✅ DO: Validate inputs early

```go
// ✅ GOOD
func (s *Scanner) Scan(ctx context.Context, file string) error {
    // Validate immediately
    if file == "" {
        return fmt.Errorf("file path required")
    }
    
    info, err := os.Stat(file)
    if err != nil {
        return fmt.Errorf("cannot access file: %w", err)
    }
    
    if !info.Mode().IsRegular() {
        return fmt.Errorf("not a regular file")
    }
    
    // Proceed with logic
    return s.scan(ctx, file)
}
```

---

## Security

### ✅ DO: Validate file paths

```go
// ✅ GOOD
func ValidateFilePath(path string) error {
    // Normalize path
    absPath, err := filepath.Abs(path)
    if err != nil {
        return fmt.Errorf("invalid path: %w", err)
    }
    
    // Prevent path traversal
    if !strings.HasPrefix(absPath, allowedDir) {
        return fmt.Errorf("path %q outside allowed directory", path)
    }
    
    return nil
}
```

### ✅ DO: Use timeouts for external calls

```go
// ✅ GOOD
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

resp, err := dockerClient.ContainerCreate(ctx, ...)
if err != nil {
    return fmt.Errorf("create container: %w", err)
}
```

### ✅ DO: Limit concurrent operations

```go
// ✅ GOOD - Prevent resource exhaustion
sem := make(chan struct{}, maxConcurrent)
for _, plugin := range plugins {
    go func(p Plugin) {
        sem <- struct{}{}        // Acquire
        defer func() { <-sem }() // Release
        p.Run(ctx)
    }(plugin)
}
```

---

## Performance

### ✅ DO: Use sync.Pool for object reuse

```go
// ✅ GOOD - Reduce GC pressure
var bufferPool = sync.Pool{
    New: func() interface{} {
        return new(bytes.Buffer)
    },
}

func processData(data []byte) error {
    buf := bufferPool.Get().(*bytes.Buffer)
    defer bufferPool.Put(buf)
    buf.Reset()
    
    // Use buffer
    return nil
}
```

### ✅ DO: Batch operations when possible

```go
// ✅ GOOD - Fewer ES calls
var results []Result
for _, plugin := range plugins {
    if result, err := plugin.Run(ctx); err == nil {
        results = append(results, result)
    }
}

// Single bulk insert
if err := es.BulkIndex(ctx, results); err != nil {
    return fmt.Errorf("bulk insert: %w", err)
}
```

### ❌ DON'T: Create unnecessary allocations

```go
// ❌ BAD - Allocates string N times
for i := 0; i < n; i++ {
    s := fmt.Sprintf("item_%d", i)  // New allocation each time
    process(s)
}

// ✅ GOOD - Preallocate
sb := strings.Builder{}
for i := 0; i < n; i++ {
    sb.Reset()
    sb.WriteString("item_")
    sb.WriteString(strconv.Itoa(i))
    process(sb.String())
}
```

---

## Code Review Checklist

- [ ] Errors are wrapped with context using `%w`
- [ ] All goroutines receive context parameter
- [ ] All resources are deferred cleanup
- [ ] No `log.Fatal()` in libraries
- [ ] Logging is structured with fields
- [ ] Timeouts are set for external calls
- [ ] Input validation occurs early
- [ ] File paths are validated for traversal
- [ ] Channel buffers are considered
- [ ] WaitGroups use defer Done()
- [ ] Tests exist for happy path
- [ ] Tests exist for error cases
- [ ] Mock interfaces are used for dependencies
- [ ] No unnecessary allocations
- [ ] Concurrent operations are bounded

---

## Common Patterns

### Pattern 1: Request with Timeout

```go
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

result, err := client.Do(ctx, request)
```

### Pattern 2: Parallel Work with Error Handling

```go
type Result struct {
    Value interface{}
    Error error
}

results := make(chan Result, len(workers))
var wg sync.WaitGroup

for _, w := range workers {
    wg.Add(1)
    go func(w Worker) {
        defer wg.Done()
        value, err := w.Work(ctx)
        results <- Result{Value: value, Error: err}
    }(w)
}

go func() {
    wg.Wait()
    close(results)
}()

for r := range results {
    if r.Error != nil {
        // Handle error
        continue
    }
    // Process r.Value
}
```

### Pattern 3: Resource Cleanup

```go
resource, err := Acquire()
if err != nil {
    return err
}
defer func() {
    if err := resource.Close(); err != nil {
        log.WithError(err).Warn("close resource")
    }
}()

// Use resource
```

---

## References

- [Effective Go](https://golang.org/doc/effective_go)
- [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
- [Context Package](https://golang.org/pkg/context/)
- [The Go Blog: Error Handling](https://go.dev/blog/error-handling-and-go)

