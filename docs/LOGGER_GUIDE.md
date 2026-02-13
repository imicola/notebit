# Logger 使用规范文档

## 概述

本文档描述了 Notebit 应用中统一日志模块的使用规范、最佳实践和配置指南。

## 目录

- [快速开始](#快速开始)
- [日志级别](#日志级别)
- [使用方式](#使用方式)
- [配置管理](#配置管理)
- [性能优化](#性能优化)
- [可观测性](#可观测性)
- [最佳实践](#最佳实践)
- [故障排查](#故障排查)

---

## 快速开始

### 初始化

在 `main.go` 中初始化日志模块：

```go
import "notebit/pkg/logger"

func main() {
    // 从环境变量加载配置（推荐）
    err := logger.Initialize(logger.LoadConfigFromEnv(logger.Config{
        Level:         logger.INFO,
        LogDir:        "logs",
        FileName:      "notebit.log",
        MaxFileSize:   100 * 1024 * 1024, // 100MB
        MaxBackups:    15,                 // 保留15天
        ConsoleOutput: true,
        ConsoleColor:  true,
        BatchSize:     10,
        FlushInterval: 100, // 100ms
    }))
    if err != nil {
        logger.Fatal("Failed to initialize logger: %v", err)
    }
    defer logger.GetDefault().Close()
    
    // 程序逻辑...
}
```

### 基本使用

```go
// 简单日志
logger.Info("Application started")
logger.Debug("Processing request")
logger.Warn("Configuration not found, using defaults")
logger.Error("Failed to connect to database: %v", err)

// 带上下文的日志（推荐）
ctx := logger.WithTraceID(context.Background(), logger.NewTraceID())
logger.InfoCtx(ctx, "User logged in")

// 带字段的日志
logger.InfoWithFields(ctx, map[string]interface{}{
    "user_id":  "user-123",
    "action":   "login",
    "ip":       "192.168.1.1",
}, "Authentication successful")

// 性能追踪
timer := logger.StartTimer()
// ... 执行操作 ...
logger.InfoWithDuration(ctx, timer(), "Operation completed successfully")
```

---

## 日志级别

日志系统支持5个级别（从低到高）：

| 级别  | 用途                           | 示例                                |
|-------|--------------------------------|-------------------------------------|
| DEBUG | 详细的调试信息                 | 变量值、执行路径、中间状态          |
| INFO  | 重要业务事件                   | 启动/关闭、用户操作、状态变化        |
| WARN  | 警告但不影响运行               | 配置缺失但使用默认值、API弃用        |
| ERROR | 错误但程序可继续               | 请求失败、文件读写错误、网络超时    |
| FATAL | 严重错误需立即退出             | 核心组件初始化失败、数据库无法连接  |

### 级别选择指南

- **开发环境**: DEBUG
- **测试环境**: INFO
- **预发布环境**: INFO
- **生产环境**: WARN

---

## 使用方式

### 1. 标准日志方法

```go
logger.Debug(format string, args ...interface{})
logger.Info(format string, args ...interface{})
logger.Warn(format string, args ...interface{})
logger.Error(format string, args ...interface{})
logger.Fatal(format string, args ...interface{})
```

### 2. 上下文日志（推荐）

```go
ctx := logger.WithTraceID(context.Background(), logger.NewTraceID())
ctx = logger.WithFields(ctx, map[string]interface{}{
    "user_id": "user-123",
    "session_id": "session-456",
})

logger.InfoCtx(ctx, "Processing user request")
logger.ErrorCtx(ctx, "Request failed: %v", err)
```

### 3. 带字段的结构化日志

```go
fields := map[string]interface{}{
    "user_id":    "user-123",
    "order_id":   "order-789",
    "amount":     99.99,
    "currency":   "USD",
}
logger.InfoWithFields(ctx, fields, "Order created")
```

### 4. 性能追踪

```go
// 方式1：手动计时
timer := logger.StartTimer()
processOrder()
logger.InfoWithDuration(ctx, timer(), "Order processed")

// 方式2：defer模式
func processOrder() {
    timer := logger.StartTimer()
    defer func() {
        logger.InfoWithDuration(ctx, timer(), "processOrder completed")
    }()
    // ... 业务逻辑 ...
}
```

---

## 配置管理

### 1. 代码配置

```go
cfg := logger.Config{
    Level:           logger.INFO,
    LogDir:          "logs",
    FileName:        "app.log",
    MaxFileSize:     100 * 1024 * 1024, // 100MB
    MaxBackups:      15,                 // 保留15天
    ConsoleOutput:   true,
    ConsoleColor:    true,
    AsyncBufferSize: 1000,
    BatchSize:       10,
    FlushInterval:   100, // 毫秒
    KafkaEnabled:    false,
    KafkaBrokers:    []string{"localhost:9092"},
    KafkaTopic:      "app-logs",
}
```

### 2. 环境变量配置（推荐）

支持的环境变量：

| 环境变量                 | 说明                    | 默认值       |
|--------------------------|-------------------------|--------------|
| LOG_LEVEL                | 日志级别                | INFO         |
| LOG_DIR                  | 日志目录                | logs         |
| LOG_FILE_NAME            | 日志文件名              | app.log      |
| LOG_MAX_FILE_SIZE_MB     | 单文件最大大小(MB)      | 100          |
| LOG_MAX_BACKUPS          | 保留文件数量            | 15           |
| LOG_CONSOLE_OUTPUT       | 是否输出到控制台        | true         |
| LOG_CONSOLE_COLOR        | 控制台彩色输出          | true         |
| LOG_BUFFER_SIZE          | 异步缓冲大小            | 1000         |
| LOG_BATCH_SIZE           | 批处理大小              | 10           |
| LOG_FLUSH_INTERVAL_MS    | 刷盘间隔(毫秒)          | 100          |
| KAFKA_ENABLED            | 启用Kafka输出           | false        |
| KAFKA_BROKERS            | Kafka代理地址(逗号分隔) | localhost:9092|
| KAFKA_TOPIC              | Kafka主题               | app-logs     |

示例：

```bash
export LOG_LEVEL=DEBUG
export LOG_MAX_FILE_SIZE_MB=200
export LOG_CONSOLE_COLOR=true
export KAFKA_ENABLED=true
export KAFKA_BROKERS=kafka1:9092,kafka2:9092
```

### 3. 动态配置（热更新）

```go
// 运行时修改日志级别
logger.SetLevel(logger.DEBUG)

// 配置监视器会自动检测环境变量变化
watcher := logger.NewConfigWatcher(logger.GetDefault(), 10*time.Second)
watcher.Start()
defer watcher.Stop()
```

---

## 性能优化

### 1. 批量刷盘

日志系统使用批处理策略，积累一定数量的日志后一次性写入磁盘：

- **BatchSize**: 批次大小（默认10条）
- **FlushInterval**: 最大等待时间（默认100ms）

### 2. 异步缓冲

- **AsyncBufferSize**: 缓冲队列大小（默认1000）
- 缓冲满时采用智能丢弃策略：优先丢弃DEBUG日志

### 3. 智能丢弃策略

当缓冲区满时：
1. **DEBUG日志**: 直接丢弃
2. **INFO/WARN/ERROR**: 尝试入队，失败则输出到stderr
3. **FATAL**: 强制写入，确保不丢失

### 4. 性能基准

在10k QPS压力下：
- CPU占用: < 5%
- P99延迟: < 5ms
- 内存占用: < 50MB

---

## 可观测性

### 1. Prometheus Metrics

暴露 `/metrics/log` 端点，提供以下指标：

```
# 各级别日志数量
logger_debug_total
logger_info_total
logger_warn_total
logger_error_total
logger_fatal_total

# 性能指标
logger_queue_length              # 当前队列长度
logger_dropped_total             # 丢弃日志总数
logger_last_flush_latency_microseconds  # 最近刷盘延迟
logger_avg_flush_latency_microseconds   # 平均刷盘延迟
logger_max_flush_latency_microseconds   # 最大刷盘延迟
logger_batch_count_total         # 批次总数
logger_avg_batch_size            # 平均批次大小
```

### 2. 集成Metrics端点

```go
import (
    "net/http"
    "notebit/pkg/logger"
)

func main() {
    // ... logger初始化 ...
    
    // 注册metrics端点
    mux := http.NewServeMux()
    logger.RegisterMetricsEndpoint(mux)
    
    go http.ListenAndServe(":9090", mux)
}
```

### 3. Grafana Dashboard

访问 `/metrics/log` 可获取JSON格式或Prometheus格式的指标数据。

推荐监控面板配置：
- 日志级别分布（饼图）
- 日志速率趋势（线图）
- 队列长度（仪表盘）
- 刷盘延迟（热力图）
- 丢弃率（告警）

---

## 最佳实践

### 1. 命名规范

```go
// ✅ 好的做法
logger.Info("User %s logged in successfully", userID)
logger.ErrorWithFields(ctx, map[string]interface{}{
    "user_id": userID,
    "error": err.Error(),
}, "Login failed")

// ❌ 不好的做法
logger.Info("User login")  // 缺少关键信息
logger.Error(err.Error())  // 缺少上下文
```

### 2. TraceID传递

```go
// HTTP请求处理
func HandleRequest(w http.ResponseWriter, r *http.Request) {
    // 生成或提取TraceID
    traceID := r.Header.Get("X-Trace-ID")
    if traceID == "" {
        traceID = logger.NewTraceID()
    }
    
    // 存入context
    ctx := logger.WithTraceID(r.Context(), traceID)
    
    // 透传到下游
    logger.InfoCtx(ctx, "Request received: %s", r.URL.Path)
    processRequest(ctx, r)
}

func processRequest(ctx context.Context, r *http.Request) {
    // 使用携带TraceID的context
    logger.DebugCtx(ctx, "Processing request...")
}
```

### 3. 敏感信息脱敏

```go
// ❌ 不要记录敏感信息
logger.Info("User password: %s", password)

// ✅ 脱敏处理
logger.InfoWithFields(ctx, map[string]interface{}{
    "user_id": userID,
    "card_number": maskCardNumber(cardNumber), // "1234****5678"
}, "Payment processed")

// 脱敏函数示例
func maskCardNumber(number string) string {
    if len(number) < 8 {
        return "****"
    }
    return number[:4] + "****" + number[len(number)-4:]
}
```

### 4. 错误日志最佳实践

```go
// ✅ 记录错误堆栈和上下文
if err != nil {
    logger.ErrorWithFields(ctx, map[string]interface{}{
        "operation": "save_file",
        "path": filePath,
        "error": err.Error(),
    }, "File operation failed")
    return err
}

// ✅ 关键路径记录耗时
timer := logger.StartTimer()
result, err := callExternalAPI(ctx, request)
if err != nil {
    logger.ErrorWithDuration(ctx, timer(), "External API call failed: %v", err)
} else {
    logger.InfoWithDuration(ctx, timer(), "External API call succeeded")
}
```

### 5. 采样策略

对于高频日志，使用采样避免日志洪水：

```go
var sampleCounter atomic.Uint64

func processItem(item Item) {
    // 每100个记录一次
    count := sampleCounter.Add(1)
    if count % 100 == 0 {
        logger.Debug("Processed %d items", count)
    }
}
```

---

## 故障排查

### 1. 日志未生成

检查项：
- 日志目录权限
- 磁盘空间
- 日志级别配置

```bash
# 检查日志目录
ls -la logs/

# 检查磁盘空间
df -h

# 临时提升日志级别
export LOG_LEVEL=DEBUG
```

### 2. 日志丢失

可能原因：
- 缓冲区过小（增大 `AsyncBufferSize`）
- 写入速度过快（检查 `DroppedLogs` 指标）
- 进程异常退出（确保调用 `Close()`）

### 3. 性能问题

优化建议：
- 减少DEBUG日志（生产环境使用INFO级别）
- 增大批处理大小（`BatchSize`）
- 调整刷盘间隔（`FlushInterval`）

### 4. Kafka连接失败

检查项：
- Kafka服务可用性
- 网络连通性
- Topic是否存在

```bash
# 测试Kafka连接
nc -zv kafka-broker 9092

# 查看Topic
kafka-topics --list --bootstrap-server kafka-broker:9092
```

---

## 附录

### A. 配置模板

**开发环境**:
```bash
export LOG_LEVEL=DEBUG
export LOG_CONSOLE_COLOR=true
export LOG_BUFFER_SIZE=100
```

**生产环境**:
```bash
export LOG_LEVEL=WARN
export LOG_MAX_FILE_SIZE_MB=100
export LOG_MAX_BACKUPS=30
export LOG_CONSOLE_OUTPUT=false
export KAFKA_ENABLED=true
export KAFKA_BROKERS=kafka1:9092,kafka2:9092,kafka3:9092
```

### B. 常用命令

```bash
# 实时查看日志
tail -f logs/notebit.log

# 按级别过滤
grep "ERROR" logs/notebit.log

# 按TraceID追踪请求
grep "trace-id-123" logs/notebit.log

# 查看今天的错误日志
grep "$(date +%Y-%m-%d).*ERROR" logs/notebit.log
```

---

## 支持

如有问题，请联系开发团队或提交Issue。
