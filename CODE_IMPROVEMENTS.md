# کد بررسی جامع و نقاط بهبود Malice Codebase

**تاریخ:** 31 دسامبر 2025

## خلاصه اجمالی
تحلیل عمیق کل Malice codebase نشان می‌دهد که علی‌رغم مدرن‌سازی اخیر، چندین حوزه برای بهبود وجود دارد.

---

## 1️⃣ مشکلات Error Handling (**اولویت بالا**)

### 🔴 مشکل 1.1: استفاده نادرست از `log.Fatal`
**فایل‌های تأثیر‌گذار:**
- `utils/utils.go` (line 36-38): `Assert()` function
- `commands/watch.go` (line 41, 55)
- `malice/docker/client/utils.go` (line 26)
- `malice/docker/machine.go` (commented, اما نمونه بد)

**مشکل:**
```go
// BAD - برنامه بدون graceful shutdown خاموش می‌شود
func Assert(err error) {
	if err != nil {
		log.Fatal(err)  // ❌ غیرقابل بازیابی
	}
}
```

**بهبود:**
```go
// GOOD - error را return می‌کنیم
func Assert(err error) error {
	return err  // ✅ قابل کنترل
}
// یا
func Must(err error) {
	if err != nil {
		panic(fmt.Sprintf("unexpected error: %v", err))  // ✅ قابل recovery
	}
}
```

### 🔴 مشکل 1.2: بدون custom error types
**فایل: `malice/errors/errors.go`**

مشکل: functions مثل `CheckError()` و `CheckErrorWithMessage()` استفاده می‌کنند اما:
- بدون ساختار مناسب
- به جای return، لاگ می‌کنند
- **Return value semantics غلط**: `true` یعنی NO ERROR (confusing!)

**بهبود:**
```go
// Custom error type
type MaliceError struct {
    Code    string
    Message string
    Err     error
    Context map[string]interface{}
}

func (e *MaliceError) Error() string {
    return fmt.Sprintf("[%s] %s: %v", e.Code, e.Message, e.Err)
}

// استفاده
if err != nil {
    return &MaliceError{
        Code:    "SCAN_FAILED",
        Message: "failed to scan file",
        Err:     err,
        Context: map[string]interface{}{"file": path},
    }
}
```

### 🔴 مشکل 1.3: Error wrapping نامناسب
**فایل: `commands/scan.go` (line 30)**

```go
// BAD - معلومات کم
err := container.List(docker, true)
if err != nil {
    return errors.Wrap(err, "failed to list containers")
}

// GOOD - با context بیشتر
if err != nil {
    return fmt.Errorf("list containers for cleanup: %w", err)
}
```

**پیشنهاد:** استفاده از `fmt.Errorf(...%w)` یا `errors.WithContext()`

---

## 2️⃣ مشکلات Goroutine Management (اولویت: **بالا**)

### 🔴 مشکل 2.1: Infinite loop بدون exit mechanism
**فایل: `commands/watch.go` (line 46-56)**

```go
done := make(chan bool)
go func() {
    for {  // ❌ بی‌نهایت - نمی‌تونه خاموش بشه
        select {
        case event := <-watcher.Events:
            // ...
        case err := <-watcher.Errors:
            // ...
        }
    }
}()
<-done  // ❌ این برای ابد چ ولی می‌شه! Dead code
```

**بهبود:**
```go
done := make(chan struct{})
ctx, cancel := context.WithCancel(context.Background())
defer cancel()

go func() {
    defer close(done)
    for {
        select {
        case event := <-watcher.Events:
            if event.Op&fsnotify.Create == fsnotify.Create {
                if err := cmdScan(event.Name, false); err != nil {
                    log.WithError(err).Error("scan failed")
                }
            }
        case err := <-watcher.Errors:
            log.WithError(err).Error("watcher error")
            return  // ✅ Exit on error
        case <-ctx.Done():
            return  // ✅ Exit on context cancellation
        }
    }
}()
return nil
```

### 🔴 مشکل 2.2: WaitGroup بدون timeout
**فایل: `commands/scan.go` (line 125-135)**

```go
var wg sync.WaitGroup
wg.Add(len(pluginsForMime))

for _, plugin := range pluginsForMime {
    go plugin.StartPlugin(docker, file.SHA256, scanID, true, elasticsearchInDocker, &wg)
}

wg.Wait()  // ❌ اگر plugin بدون reason crash کنه؟ → تا ابد منتظر!
```

**بهبود:**
```go
var wg sync.WaitGroup
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
defer cancel()

for _, plugin := range pluginsForMime {
    wg.Add(1)
    go func(p plugins.Plugin) {
        defer wg.Done()
        if err := p.StartPluginWithContext(ctx, docker, file.SHA256, scanID); err != nil {
            log.WithError(err).Warnf("plugin %s failed", p.Name)
            // Continue with other plugins - don't fail entire scan
        }
    }(plugin)
}

// Wait with timeout
done := make(chan struct{})
go func() {
    wg.Wait()
    close(done)
}()

select {
case <-done:
    log.Debug("all plugins completed")
case <-ctx.Done():
    return fmt.Errorf("scanning timeout after %v", 5*time.Minute)
}
```

### 🔴 مشکل 2.3: Goroutine leak در API server
**فایل: `api/server/server.go` (line 72-90)**

```go
for _, srv := range s.servers {
    srv.srv.Handler = s.routerSwapper
    go func(srv *HTTPServer) {  // ❌ اگر server بدون خاموش نشه؟
        var err error
        logrus.Infof("API listen on %s", srv.l.Addr())
        if err = srv.Serve(); err != nil && strings.Contains(err.Error(), "use of closed network connection") {
            err = nil
        }
        chErrors <- err
    }(srv)
}
```

**بهبود:**
```go
ctx, cancel := context.WithCancel(context.Background())
s.cancel = cancel  // Store for later cleanup

for _, srv := range s.servers {
    srv.srv.Handler = s.routerSwapper
    srv.srv.BaseContext = func(net.Listener) context.Context { return ctx }
    
    go func(srv *HTTPServer) {
        defer func() {
            if r := recover(); r != nil {
                logrus.Errorf("server panic: %v", r)
                chErrors <- fmt.Errorf("server panicked: %v", r)
            }
        }()
        
        logrus.Infof("API listen on %s", srv.l.Addr())
        if err := srv.Serve(); err != nil && err != http.ErrServerClosed {
            logrus.WithError(err).Error("serve error")
            chErrors <- err
        }
    }(srv)
}

// Add graceful shutdown
s.shutdownOnce = &sync.Once{}
```

---

## 3️⃣ مشکلات Resource Management (اولویت: **بالا**)

### 🔴 مشکل 3.1: بدون defer cleanup
**فایل: `commands/watch.go`**

```go
watcher, err := fsnotify.NewWatcher()
if err != nil {
    log.Fatal(err)
}
defer watcher.Close()  // ✅ خوب است، اما...

done := make(chan bool)
go func() {
    // ...
}()
<-done  // ❌ این هرگز اجرا نمی‌شود!
```

### 🔴 مشکل 3.2: Connection pooling نیست
**فایل: `malice/database/database.go` و `commands/scan.go`**

مشکل: هر بار که `es.Init()` call می‌شود، نیا connection ساخته می‌شود

```go
// BAD - بدون caching
for _, plugin := range pluginsForMime {
    go plugin.StartPlugin(docker, file.SHA256, scanID, true, elasticsearchInDocker, &wg)
    // Each plugin creates new ES connection!
}
```

**بهبود:** Connection pool singleton

```go
var (
    esClient *elasticsearch.Client
    esOnce   sync.Once
)

func GetESClient(url string) (*elasticsearch.Client, error) {
    var err error
    esOnce.Do(func() {
        cfg := elasticsearch.Config{
            Addresses: []string{url},
            MaxRetries: 3,
            RetryBackoff: func(attempt int) time.Duration {
                return time.Duration(math.Pow(2, float64(attempt))) * time.Second
            },
        }
        esClient, err = elasticsearch.NewClient(cfg)
    })
    return esClient, err
}
```

### 🔴 مشکل 3.3: بدون proper channel cleanup
**فایل: `api/server/server.go` (line 77)**

```go
var chErrors = make(chan error, len(s.servers))  // ❌ بعد از استفاده نبسته می‌شود
// ...
for i := 0; i < len(s.servers); i++ {
    err := <-chErrors
    if err != nil {
        return err
    }
}
// chErrors leak می‌شود!
```

**بهبود:**
```go
chErrors := make(chan error, len(s.servers))
defer close(chErrors)

for _, srv := range s.servers {
    go func(srv *HTTPServer) {
        chErrors <- srv.Serve()
    }(srv)
}

for i := 0; i < len(s.servers); i++ {
    if err := <-chErrors; err != nil {
        return err
    }
}
```

---

## 4️⃣ مشکلات Code Quality (اولویت: **متوسط**)

### 🔴 مشکل 4.1: Duplicate functions
**فایل‌های:**
- `utils/utils.go`: `Getopt()`, `GetOpt()` (duplicate)
- `malice/malutils/utils.go`: دوباره `Getopt()`
- `malice/malutils/utils.go`: دوباره `CopyFile()`

**بهبود:** Consolidate:
```go
// In internal/util/env.go
package util

func GetEnv(key, defaultVal string) string {
    if val := os.Getenv(key); val != "" {
        return val
    }
    return defaultVal
}

// Remove duplicates, use this everywhere
```

### 🔴 مشکل 4.2: بدون interfaces
مثال: `docker.Client` مستقیم pass می‌شود، mock نمی‌شود

```go
// BAD
func StartPlugin(docker *client.Docker, ...) {}

// GOOD
type DockerClient interface {
    ContainerStart(ctx context.Context, containerID string, options types.ContainerStartOptions) error
    ContainerCreate(ctx context.Context, config *container.Config, ...) (container.CreateResponse, error)
    // ...
}

func StartPlugin(ctx context.Context, client DockerClient, ...) error {}
```

### 🔴 مشکل 4.3: Magic numbers بدون constants
**فایل: `commands/scan.go`**

```go
// BAD
for i := 0; i < maxAttempts; i++ {
    time.Sleep(3 * time.Second)
}

// GOOD
const (
    maxRetries    = 60
    retryInterval = 3 * time.Second
    scanTimeout   = 5 * time.Minute
)
```

### 🔴 مشکل 4.4: Global variables نامناسب
**فایل: `main.go`**

```go
var (
    version = "dev"
    commit  = "none"
    date    = "unknown"
)
```

**بهبود:** Use build flags:
```bash
go build -ldflags "-X main.version=$VERSION -X main.commit=$COMMIT"
```

### 🔴 مشکل 4.5: بدون structured logging
**فایل‌های متعدد**

```go
// BAD
log.Println("event:", event)
log.Debugf(">>>>> RUNNING Plugin: %s >>>>>>>>>>>>>>>>>>>>>>>>>>>>>>", plugin.Name)

// GOOD
log.WithFields(log.Fields{
    "event": event,
    "type":  "file_created",
}).Debug("file system event")

log.WithField("plugin", plugin.Name).
    WithField("file", file.SHA256).
    Debug("running plugin scan")
```

---

## 5️⃣ مشکلات Performance (اولویت: **متوسط**)

### 🔴 مشکل 5.1: بی‌جا concurrent operations
**فایل: `commands/scan.go` (line 46-51)**

```go
// BAD - intel plugins serial
plugins.RunIntelPlugins(docker, file.SHA1, scanID, true, elasticsearchInDocker)

// Then AV plugins parallel
// But intel result impact nیست!

// GOOD - parallel where possible
type ScanStage struct {
    Name    string
    Timeout time.Duration
    Run     func(ctx context.Context) error
}

stages := []ScanStage{
    {Name: "intel", Timeout: 2*time.Minute, Run: func(ctx context.Context) error {
        return plugins.RunIntelPlugins(ctx, docker, ...)
    }},
    {Name: "mime", Timeout: 30*time.Second, Run: func(ctx context.Context) error {
        return persist.GetMimeType(ctx, docker, ...)
    }},
}

// Execute stages and AV plugins in parallel
```

### 🔴 مشکل 5.2: بی‌جا logging
**فایل: `api/server/middleware.go` (line 26-30)**

```go
// BAD - همه request لاگ می‌شود حتی 404 و health checks
if s.cfg.Logging && logrus.GetLevel() == logrus.DebugLevel {
    log.Printf("DEBUG: %s %s", r.Method, r.URL.Path)
}

// GOOD - strategic logging
if !isHealthCheck(r.URL.Path) && !is404(r.URL.Path) {
    log.WithFields(log.Fields{
        "method":   r.Method,
        "path":     r.URL.Path,
        "duration": elapsed,
    }).Debug("request handled")
}
```

### 🔴 مشکل 5.3: بدون profiling
**بهبود:**
```go
// Add pprof endpoints
import _ "net/http/pprof"

// In main
if debug {
    go func() {
        log.Println(http.ListenAndServe("localhost:6060", nil))
    }()
}
```

---

## 6️⃣ مشکلات Security (اولویت: **بالا**)

### 🟡 مشکل 6.1: بدون input validation
**فایل: `commands/scan.go` (line 22-27)**

```go
// BAD
if _, err := os.Stat(path); os.IsNotExist(err) {
    log.Fatal(path + ": no such file or directory")
}

// GOOD
func ValidateFilePath(path string) error {
    // Check path traversal
    absPath, err := filepath.Abs(path)
    if err != nil {
        return fmt.Errorf("invalid path: %w", err)
    }
    
    // Prevent path traversal
    if !strings.HasPrefix(absPath, allowedDir) {
        return fmt.Errorf("path outside allowed directory")
    }
    
    // Check permissions
    info, err := os.Stat(absPath)
    if err != nil {
        return fmt.Errorf("cannot access file: %w", err)
    }
    
    if !info.Mode().IsRegular() {
        return fmt.Errorf("not a regular file")
    }
    
    return nil
}
```

### 🟡 مشکل 6.2: بدون rate limiting
**فایل: `api/server/server.go`**

```go
// GOOD - add rate limiting middleware
import "golang.org/x/time/rate"

func (s *Server) rateLimitMiddleware(limit rate.Limit) http.HandlerFunc {
    limiter := rate.NewLimiter(limit, 10)
    return func(w http.ResponseWriter, r *http.Request) {
        if !limiter.Allow() {
            http.Error(w, "rate limit exceeded", http.StatusTooManyRequests)
            return
        }
        // Continue...
    }
}
```

### 🟡 مشکل 6.3: بدون timeout برای external services
**فایل: `malice/persist/file.go`**

```go
// BAD - بی‌نهایت منتظری
reader, err := docker.Client.ContainerLogs(ctx, contResponse.ID, options)

// GOOD
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()
reader, err := docker.Client.ContainerLogs(ctx, contResponse.ID, options)
```

---

## 7️⃣ نقاط بهبود معماری (اولویت: **بالا**)

### 📊 Refactoring پیشنهادی

#### 1. **Scan Pipeline Abstraction**
```go
type ScanPipeline interface {
    Scan(ctx context.Context, file *persist.File) (*ScanResult, error)
}

type DefaultPipeline struct {
    db            elasticsearch.Database
    docker        DockerClient
    pluginManager PluginManager
    timeout       time.Duration
}

func (p *DefaultPipeline) Scan(ctx context.Context, file *persist.File) (*ScanResult, error) {
    // Stage 1: Intel
    // Stage 2: MIME type
    // Stage 3: AV plugins (parallel)
    // Stage 4: Store results
}
```

#### 2. **Configuration Management**
```go
type AppConfig struct {
    Database    DatabaseConfig
    Docker      DockerConfig
    Logger      LoggerConfig
    Plugins     PluginConfig
    Security    SecurityConfig  // NEW
}

// Validation
func (c *AppConfig) Validate() error {
    if c.Database.Timeout <= 0 {
        return errors.New("database timeout must be positive")
    }
    // More validation...
}
```

#### 3. **Dependency Injection**
```go
type App struct {
    logger  *logrus.Logger
    docker  DockerClient
    db      elasticsearch.Database
    plugins PluginManager
}

// Use constructor
func NewApp(cfg *AppConfig) (*App, error) {
    logger := setupLogger(cfg.Logger)
    docker := setupDocker(cfg.Docker)
    db := setupDatabase(cfg.Database)
    
    return &App{
        logger:  logger,
        docker:  docker,
        db:      db,
        plugins: setupPlugins(docker),
    }, nil
}
```

---

## 8️⃣ Testing Strategy (اولویت: **متوسط**)

### مشکل: بدون unit tests
```go
// Add tests/unit/commands/scan_test.go
func TestScanValidation(t *testing.T) {
    tests := []struct {
        name  string
        path  string
        want  error
    }{
        {"valid file", "/tmp/test.bin", nil},
        {"nonexistent", "/tmp/nonexist", os.ErrNotExist},
        {"directory", "/tmp", ErrIsDirectory},
        {"path traversal", "../../etc/passwd", ErrPathTraversal},
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := ValidateFilePath(tt.path)
            if !errors.Is(err, tt.want) {
                t.Errorf("got %v, want %v", err, tt.want)
            }
        })
    }
}
```

---

## خلاصه اولویت‌ها

| اولویت | موضوع | تعداد فایل | تأثیر |
|-------|-------|----------|------|
| 🔴 بسیار بالا | Error Handling | 5+ | برنامه ممکن است ناگهان خاموش شود |
| 🔴 بسیار بالا | Goroutine Management | 3+ | Memory leak، deadlock |
| 🔴 بسیار بالا | Resource Cleanup | 4+ | File descriptor leak |
| 🟡 بالا | Security | 3+ | Path traversal، DoS |
| 🟡 متوسط | Code Quality | 8+ | Maintainability |
| 🟡 متوسط | Performance | 3+ | Scalability |
| 🔵 پایین | Architecture | - | Long-term maintainability |

---

## فایل‌های اول برای تصحیح (Order of Implementation)

1. ✅ `malice/errors/errors.go` - Refactor error types
2. ✅ `commands/watch.go` - Fix infinite loop + goroutine leak
3. ✅ `commands/scan.go` - Add timeouts + proper error handling
4. ✅ `api/server/server.go` - Add graceful shutdown
5. ✅ `utils/utils.go` - Remove duplicates
6. ✅ `commands/` - Add input validation

---

## فایل‌های بعدی (Nice to Have)

- Database connection pooling
- Structured logging throughout
- Performance profiling
- Unit test suite
- Integration test suite

