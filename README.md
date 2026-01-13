# Golang OpenTelemetry Logging

一個整合 OpenTelemetry 的 Golang 日誌記錄程式，使用 Zap 框架並透過 **OpenTelemetry Logs API (實驗性)** 與 OTLP/gRPC 輸出日誌數據到 OpenTelemetry Collector。**支援 Trace Context (Trace ID & Span ID)**，實現日誌與追蹤的完整關聯。

## 功能特色

- 使用 **Uber Zap** 高效能結構化日誌框架
- 整合 **OpenTelemetry Logs API** (實驗性功能)
- 整合 **OpenTelemetry Tracing** 並自動注入 trace context
- 透過 **otelzap bridge** 連接 Zap 與 OpenTelemetry
- 使用 **OTLP/gRPC** 協議輸出日誌和追蹤到 Collector
- **每條日誌自動包含 Trace ID 和 Span ID**，實現日誌與追蹤關聯
- 支援多種日誌級別 (Debug, Info, Warn, Error)
- 豐富的結構化欄位，方便日誌分析
- 雙重輸出：人類易讀的彩色控制台 + 完整結構化 OTLP 日誌

## 系統需求

- **Go 1.21** 或更高版本
- **OpenTelemetry Collector** 運行於 `localhost:4317`，並配置 logs 和 traces pipeline

## 安裝步驟

1. 複製或下載此專案
2. 安裝依賴套件：

```bash
go mod download
```

## OpenTelemetry Collector 配置

確保您的 OTLP Collector 配置檔包含 logs 和 traces pipeline，例如：

```yaml
receivers:
  otlp:
    protocols:
      grpc:
        endpoint: 0.0.0.0:4317

processors:
  batch:

exporters:
  logging:
    loglevel: debug

service:
  pipelines:
    traces:
      receivers: [otlp]
      processors: [batch]
      exporters: [logging]
    logs:
      receivers: [otlp]
      processors: [batch]
      exporters: [logging]
```

## 配置設定

編輯 `config.yaml` 設定 OTLP 端點：

```yaml
otlp:
  endpoint: "localhost:4317"  # OTLP collector 端點
  insecure: true              # 設為 false 啟用 TLS
  
service:
  name: "golang-logging-demo"
  version: "1.0.0"
```

## 使用方式

### 編譯程式

```bash
go build -o golang-logging.exe
```

### 執行程式

```bash
# Windows
.\golang-logging.exe

# Linux/Mac
./golang-logging
```

### 直接運行（開發模式）

```bash
go run main.go
```

## 程式碼範例

```go
log := logger.Logger

// Info 級別 + 結構化欄位
log.Info("User logged in",
    zap.Int("user_id", 12345),
    zap.String("action", "login"))

// Warning 級別
log.Warn("High resource usage",
    zap.String("resource", "database"),
    zap.Float64("usage_percent", 85.5))

// Error 級別
log.Error("Operation failed",
    zap.String("error_code", "DB_CONNECTION_FAILED"),
    zap.Int("retry_count", 3))

// Debug 級別
log.Debug("Debug information",
    zap.String("request_id", "req-123"),
    zap.Int("duration_ms", 245))
```

## 專案結構

```
.
├── main.go                      # 程式進入點與日誌範例
├── config.yaml                  # 配置檔案
├── go.mod                       # Go 模組定義
├── go.sum                       # 依賴校驗檔
├── internal/
│   ├── config/
│   │   └── config.go           # 配置載入器
│   └── logger/
│       └── otlp.go             # OTLP 日誌初始化與 otelzap bridge
└── README.md                    # 本文件
```

## 運作原理

1. **配置載入**：從 `config.yaml` 讀取 OTLP 端點設定
2. **OTLP Trace Exporter 初始化**：建立 OpenTelemetry Trace Exporter (gRPC)
3. **OTLP Log Exporter 初始化**：建立 OpenTelemetry Log Exporter (gRPC)
4. **TracerProvider 設定**：建立 TracerProvider 並設為全域提供者
5. **LoggerProvider 設定**：建立 LoggerProvider 並設為全域提供者
6. **Propagator 設定**：配置 TraceContext 和 Baggage propagator
7. **otelzap Bridge**：使用 `otelzap.NewCore()` 建立橋接 Core，自動處理 trace context
8. **Zap Logger 配置**：將 otelzap Core 與彩色控制台 Core 結合 (Tee)
9. **日誌輸出**：
   - 日誌寫入 **控制台** (彩色人類易讀格式)
   - 同時透過 **OTLP/gRPC** 發送到 Collector (包含完整 trace context)
   - 追蹤資訊透過 **OTLP/gRPC** 發送到 Collector
10. **Trace Context 自動注入**：otelzap 會自動從 context 中提取 trace ID 和 span ID 注入日誌
11. **優雅關閉**：確保所有日誌和追蹤都已同步並發送後才結束程式

## 主要依賴套件

- `go.uber.org/zap` - 高效能結構化日誌框架
- `go.opentelemetry.io/contrib/bridges/otelzap` - Zap 到 OpenTelemetry 的橋接
- `go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploggrpc` - OTLP Logs gRPC exporter
- `go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc` - OTLP Traces gRPC exporter
- `go.opentelemetry.io/otel/sdk/log` - OpenTelemetry Logs SDK (實驗性)
- `go.opentelemetry.io/otel/trace` - OpenTelemetry Tracing API
- `gopkg.in/yaml.v3` - YAML 配置解析

## 驗證日誌輸出

### 1. 控制台輸出

程式運行時會在控制台看到：

```
2026/01/13 10:21:25 ✅ OTLP Logger initialized successfully, sending logs to: localhost:4317
2026/01/13 10:21:25 ✅ OTLP Tracer initialized successfully, trace context will be included in logs
INFO   Application started     {"service": "golang-logging-demo", "version": "1.0.0", "trace_id": "72fc562b386c78dd42b2268ddf012c6b", "span_id": "fa97387a09f8f6fc"}
WARN   Warning message         {"service": "golang-logging-demo", "version": "1.0.0", "trace_id": "72fc562b386c78dd42b2268ddf012c6b", "span_id": "b5776e6b7e5e21e0", "resource": "database", "usage_percent": 85.5}
```

注意：
- 相同的 **trace_id** (`72fc562b386c78dd42b2268ddf012c6b`) 表示這些日誌屬於同一個請求鏈路
- 不同的 **span_id** 表示不同的操作階段（父 span 和子 span）

### 2. OTLP Collector 輸出

在您的 OTLP Collector console 應該會看到：

**Traces:**
```
2026-01-13T10:21:25.123+0800    info    Traces  {"kind": "exporter", "data_type": "traces", "name": "logging"}
Span #0
    Trace ID       : 72fc562b386c78dd42b2268ddf012c6b
    Parent ID      : fa97387a09f8f6fc
    ID             : b5776e6b7e5e21e0
    Name           : database-query
    Kind           : Internal
```

**Logs with Trace Context:**
```
2026-01-13T10:21:25.124+0800    info    Logs    {"kind": "exporter", "data_type": "logs", "name": "logging"}
LogRecord #0
Timestamp: 2026-01-13 02:21:25 +0000 UTC
Severity: Info
Body: Application started
Attributes:
     -> service: Str(golang-logging-demo)
     -> version: Str(1.0.0)
     -> trace_id: Str(72fc562b386c78dd42b2268ddf012c6b)
     -> span_id: Str(fa97387a09f8f6fc)
TraceID: 72fc562b386c78dd42b2268ddf012c6b
SpanID: fa97387a09f8f6fc
```

關鍵特性：
- ✅ 日誌中包含 **TraceID** 和 **SpanID** 作為頂層欄位
- ✅ 也同時包含在 **Attributes** 中
- ✅ 可以在 Jaeger/Zipkin 等追蹤系統中點擊 trace，查看關聯的所有日誌

## 實驗性功能說明

⚠️ **重要提醒**：OpenTelemetry Logs API 目前還在**實驗階段 (Experimental)**，這意味著：

- API 介面可能會在未來版本中變更
- 部分功能可能還不夠穩定
- 建議在生產環境使用前進行充分測試

但這也是 OpenTelemetry 官方推薦的日誌整合方案，未來會成為穩定版本。

## 故障排除

### ❌ Connection refused 錯誤

```
failed to initialize OTLP logger: context deadline exceeded
```

**解決方案：**
- 確認 OTLP collector 正在 `localhost:4317` 運行
- 檢查防火牆設定
- 驗證 `config.yaml` 中的端點配置
- 測試連線：`telnet localhost 4317`

### ❌ Collector 沒有收到日誌

**檢查項目：**
1. Collector 配置中是否啟用了 `logs` pipeline
2. Collector 的 `receivers.otlp.protocols.grpc` 是否正確配置
3. 查看 Collector 的日誌是否有錯誤訊息
4. 確認 `exporters` 中有日誌輸出配置（如 `logging` exporter）

### ❌ 編譯錯誤

```bash
# 清除快取並重新下載依賴
go clean -modcache
go mod tidy
go mod download
```

### ❌ 版本衝突

OpenTelemetry 版本要求：
- `go.opentelemetry.io/otel` >= 1.33.0
- `go.opentelemetry.io/otel/sdk/log` >= 0.7.0
- `go.opentelemetry.io/contrib/bridges/otelzap` >= 0.7.0

## 測試環境

本程式預設連接到 `localhost:4317`。您已有測試環境運行在此端點，執行程式即可開始發送日誌到 Collector。

建議配置 Collector 使用 `logging` exporter 以在 console 查看收到的日誌。

## 授權條款

MIT License

## 安裝步驟

1. 複製或下載此專案
2. 安裝依賴套件：

```bash
go mod download
```

## 配置設定

編輯 `config.yaml` 設定 OTLP 端點：

```yaml
otlp:
  endpoint: "localhost:4317"  # OTLP collector 端點
  insecure: true              # 設為 false 啟用 TLS
  
service:
  name: "golang-logging-demo"
  version: "1.0.0"
```

## 使用方式

### 編譯程式

```bash
go build -o golang-logging.exe
```

### 執行程式

```bash
# Windows
.\golang-logging.exe

# Linux/Mac
./golang-logging
```

### 直接運行（開發模式）

```bash
go run main.go
```

## 程式碼範例

```go
log := logger.Logger

// Info 級別 + 結構化欄位
log.Info("User logged in",
    zap.Int("user_id", 12345),
    zap.String("action", "login"))

// Warning 級別
log.Warn("High resource usage",
    zap.String("resource", "database"),
    zap.Float64("usage_percent", 85.5))

// Error 級別
log.Error("Operation failed",
    zap.String("error_code", "DB_CONNECTION_FAILED"),
    zap.Int("retry_count", 3))

// Debug 級別
log.Debug("Debug information",
    zap.String("request_id", "req-123"),
    zap.Int("duration_ms", 245))
```

## 專案結構

```
.
├── main.go                      # 程式進入點與日誌範例
├── config.yaml                  # 配置檔案
├── go.mod                       # Go 模組定義
├── go.sum                       # 依賴校驗檔
├── internal/
│   ├── config/
│   │   └── config.go           # 配置載入器
│   └── logger/
│       └── otlp.go             # OTLP 初始化與 Zap 設定
└── README.md                    # 本文件
```

## 運作原理

1. **配置載入**：從 `config.yaml` 讀取 OTLP 端點設定
2. **OTLP 初始化**：建立 OpenTelemetry SDK 與 OTLP exporter
3. **Zap Logger 設定**：配置 Zap 作為結構化日誌框架
4. **日誌輸出**：
   - 日誌寫入 **標準輸出** (JSON 格式)
   - 追蹤資訊透過 **OTLP/gRPC** 送到 Collector
5. **優雅關閉**：確保所有日誌都已同步後才結束程式

## 主要依賴套件

- `go.uber.org/zap` - 高效能結構化日誌框架
- `go.opentelemetry.io/otel` - OpenTelemetry SDK
- `go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc` - OTLP gRPC exporter
- `gopkg.in/yaml.v3` - YAML 配置解析

## 驗證日誌輸出

您的日誌和追蹤資料應該會出現在 `localhost:4317` 的 OpenTelemetry Collector 中。驗證方式：

1. **檢查 Collector 日誌** - 查看是否有連線記錄
2. **檢查後端系統** - 例如 Jaeger、Prometheus、Grafana 等
3. **查看控制台輸出** - 日誌也會以 JSON 格式輸出到標準輸出

### 輸出範例

```json
{"level":"info","timestamp":"2024-01-13T09:40:00.123Z","caller":"main.go:29","msg":"Application started","service":"golang-logging-demo","version":"1.0.0"}
{"level":"debug","timestamp":"2024-01-13T09:40:00.124Z","caller":"main.go:33","msg":"Debug message","user_id":12345,"action":"login","service":"golang-logging-demo","version":"1.0.0"}
```

## 故障排除

### ❌ Connection refused 錯誤

- 確認 OTLP collector 正在 `localhost:4317` 運行
- 檢查防火牆設定
- 驗證 `config.yaml` 中的端點配置

### ❌ 沒有日誌出現

- 確認 Collector 配置中啟用了 logs pipeline
- 增加日誌級別到 Debug 以查看更多輸出
- 在關閉前加入延遲，確保批次處理器有時間刷新

### ❌ 編譯錯誤

```bash
# 清除快取並重新下載依賴
go clean -modcache
go mod tidy
go mod download
```

## 測試環境

本程式預設連接到 `localhost:4317`。您已有測試環境運行在此端點，直接執行程式即可開始記錄日誌。

## 授權條款

MIT License

