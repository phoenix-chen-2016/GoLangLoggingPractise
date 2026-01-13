# Golang OpenTelemetry Logging

一個整合 OpenTelemetry 的 Golang 日誌記錄程式，使用 Zap 框架並透過 **OpenTelemetry Logs API (實驗性)** 與 OTLP/gRPC 輸出日誌數據到 OpenTelemetry Collector。**支援 Trace Context (Trace ID & Span ID) 自動注入**，實現日誌與追蹤的完整關聯。

## 功能特色

- 使用 **Uber Zap** 高效能結構化日誌框架
- 整合 **OpenTelemetry Logs API** (實驗性功能)
- 整合 **OpenTelemetry Tracing** 並自動注入 trace context
- 透過 **otelzap bridge** 連接 Zap 與 OpenTelemetry
- 使用 **OTLP/gRPC** 協議輸出日誌和追蹤到 Collector
- **✨ 每條日誌自動包含 Trace ID 和 Span ID，無需手動傳遞**
- **✨ Context-aware 日誌函數，自動從 context 提取 trace 資訊**
- **✨ 控制台輸出乾淨簡潔，OTLP 輸出包含完整 trace context**
- 支援多種日誌級別 (Debug, Info, Warn, Error, Fatal)
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

### 架構圖

```
┌─────────────────────────────────────────────────────────────┐
│                        Application                          │
│                                                             │
│  ctx, span := tracer.Start(ctx, "operation")               │
│  logger.InfoContext(ctx, "message", fields...)             │
│                           ↓                                 │
│  ┌──────────────────────────────────────────────────────┐  │
│  │  Context-aware Logger Function                       │  │
│  │  • 添加 zap.Any("context", ctx) 字段                 │  │
│  └──────────────────────────────────────────────────────┘  │
│                           ↓                                 │
│  ┌──────────────────────────────────────────────────────┐  │
│  │           Zap Logger (Tee Core)                      │  │
│  │  ┌────────────────────┐  ┌─────────────────────────┐│  │
│  │  │  Console Core      │  │   otelzap Core          ││  │
│  │  │  (過濾 context)    │  │  (保留 context)         ││  │
│  │  └────────────────────┘  └─────────────────────────┘│  │
│  └──────────────────────────────────────────────────────┘  │
│         ↓                            ↓                      │
└─────────────────────────────────────────────────────────────┘
          ↓                            ↓
    ┌──────────┐              ┌────────────────────┐
    │ Console  │              │  otelzap Bridge    │
    │ Output   │              │  • 檢測 context    │
    │          │              │  • 提取 trace ID   │
    │ (簡潔)   │              │  • 提取 span ID    │
    └──────────┘              └────────────────────┘
                                       ↓
                              ┌────────────────────┐
                              │  OTLP Log Exporter │
                              └────────────────────┘
                                       ↓
                              ┌────────────────────┐
                              │ OTLP Collector     │
                              │ (localhost:4317)   │
                              └────────────────────┘
```

### 詳細步驟

1. **配置載入**：從 `config.yaml` 讀取 OTLP 端點設定

2. **OTLP Trace Exporter 初始化**：建立 OpenTelemetry Trace Exporter (gRPC)

3. **OTLP Log Exporter 初始化**：建立 OpenTelemetry Log Exporter (gRPC)

4. **TracerProvider 設定**：建立 TracerProvider 並設為全域提供者

5. **LoggerProvider 設定**：建立 LoggerProvider 並設為全域提供者

6. **Propagator 設定**：配置 TraceContext 和 Baggage propagator

7. **otelzap Bridge**：
   - 使用 `otelzap.NewCore()` 建立橋接 Core
   - 自動檢測 `context.Context` 類型的字段
   - 從 context 中提取 trace ID 和 span ID

8. **Zap Logger 配置**：
   - 建立控制台 Core（過濾掉 context 字段）
   - 建立 otelzap Core（保留 context 字段）
   - 使用 Tee 結合兩個 Core（雙重輸出）

9. **Context-aware 日誌函數**：
   - `InfoContext(ctx, msg, fields...)` 等函數
   - 自動添加 `zap.Any("context", ctx)` 字段
   - 無需手動提取 trace ID 和 span ID

10. **日誌輸出流程**：
    - **控制台 Core**：過濾掉 context 字段，輸出簡潔的日誌
    - **otelzap Core**：保留 context 字段，自動提取 trace context
    - **OTLP Exporter**：將日誌（含 trace context）發送到 Collector

11. **Trace Context 自動注入**：
    - otelzap 檢測到 `context.Context` 字段
    - 調用 `trace.SpanFromContext(ctx)` 提取 span
    - 將 `TraceID` 和 `SpanID` 注入到 OpenTelemetry log record
    - 發送到 OTLP backend 時包含完整的 trace context

12. **優雅關閉**：確保所有日誌和追蹤都已同步並發送後才結束程式

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

程式運行時會在控制台看到簡潔的輸出（**不包含** trace_id 和 span_id）：

```
2026/01/13 13:31:19 ✅ OTLP Logger initialized successfully, sending logs to: localhost:4317
2026/01/13 13:31:19 ✅ OTLP Tracer initialized successfully, trace context will be included in logs
INFO    Application started     {"service": "golang-logging-demo", "version": "1.0.0"}
WARN    Warning message         {"service": "golang-logging-demo", "version": "1.0.0", "resource": "database", "usage_percent": 85.5}
ERROR   Error occurred          {"service": "golang-logging-demo", "version": "1.0.0", "error_code": "DB_CONNECTION_FAILED", "retry_count": 3}
INFO    Processing request      {"service": "golang-logging-demo", "version": "1.0.0", "request_id": "req-abc-123", "user_agent": "Mozilla/5.0"}
```

**特點：**
- ✅ 控制台輸出乾淨簡潔，只顯示業務相關欄位
- ✅ 不顯示 context 或 trace_id/span_id，避免視覺混亂
- ✅ 使用彩色編碼（INFO=綠色, WARN=黃色, ERROR=紅色）
- ✅ Trace context 仍然會**自動發送到 OTLP backend**

### 2. OTLP Collector 輸出

在您的 OTLP Collector console 會看到**完整的 trace context**：

**Traces:**
```
2026-01-13T13:31:19.123+0800    info    Traces  {"kind": "exporter", "data_type": "traces", "name": "logging"}
Span #0
    Trace ID       : a3f8d9e2b7c4a1b5c6d7e8f9a0b1c2d3
    Parent ID      : 
    ID             : f1e2d3c4b5a69788
    Name           : main-operation
    Kind           : Internal
    Start time     : 2026-01-13 05:31:19.120 +0000 UTC
    End time       : 2026-01-13 05:31:21.125 +0000 UTC

Span #1
    Trace ID       : a3f8d9e2b7c4a1b5c6d7e8f9a0b1c2d3
    Parent ID      : f1e2d3c4b5a69788
    ID             : b9c8d7e6f5a49382
    Name           : database-query
    Kind           : Internal
```

**Logs with Trace Context（自動注入）:**
```
2026-01-13T13:31:19.124+0800    info    Logs    {"kind": "exporter", "data_type": "logs", "name": "logging"}
LogRecord #0
ObservedTimestamp: 2026-01-13 05:31:19.124 +0000 UTC
Timestamp: 2026-01-13 05:31:19.124 +0000 UTC
SeverityText: INFO
SeverityNumber: Info(9)
Body: Str(Application started)
Attributes:
     -> service: Str(golang-logging-demo)
     -> version: Str(1.0.0)
TraceID: a3f8d9e2b7c4a1b5c6d7e8f9a0b1c2d3
SpanID: f1e2d3c4b5a69788
TraceFlags: 01

LogRecord #1
ObservedTimestamp: 2026-01-13 05:31:19.125 +0000 UTC
Timestamp: 2026-01-13 05:31:19.125 +0000 UTC
SeverityText: WARN
SeverityNumber: Warn(13)
Body: Str(Warning message)
Attributes:
     -> service: Str(golang-logging-demo)
     -> version: Str(1.0.0)
     -> resource: Str(database)

### ❌ 日誌中沒有 Trace ID 和 Span ID

**可能原因和解決方案：**

1. **沒有使用 Context-aware 函數**
   ```go
   // ❌ 錯誤：直接使用 logger.Info，不會包含 trace context
   logger.Logger.Info("message")
   
   // ✅ 正確：使用 InfoContext，自動注入 trace context
   logger.InfoContext(ctx, "message")
   ```

2. **Context 中沒有 Span**
   ```go
   // ❌ 錯誤：使用空的 context
   ctx := context.Background()
   logger.InfoContext(ctx, "message")  // 不會有 trace ID
   
   // ✅ 正確：從 tracer 創建 span
   ctx, span := tracer.Start(ctx, "operation")
   defer span.End()
   logger.InfoContext(ctx, "message")  // 會包含 trace ID
   ```

3. **otelzap Core 配置錯誤**
   - 確認使用 `otelzap.NewCore()` 創建 Core
   - 確認 `LoggerProvider` 已正確初始化
   - 確認使用 `zapcore.NewTee()` 結合 console 和 otelzap Core

4. **Span 沒有正在記錄**
   ```go
   span := trace.SpanFromContext(ctx)
   if !span.IsRecording() {
       // Span 沒有在記錄，可能沒有啟動或已結束
   }
   ```
     -> usage_percent: Float64(85.5)
TraceID: a3f8d9e2b7c4a1b5c6d7e8f9a0b1c2d3
SpanID: b9c8d7e6f5a49382
TraceFlags: 01
```

**關鍵特性：**
- ✅ 每條日誌**自動包含** `TraceID` 和 `SpanID` 欄位
- ✅ 相同的 `TraceID` 串聯整個請求鏈路
- ✅ 不同的 `SpanID` 標識不同操作階段
- ✅ 可在 Jaeger/Grafana 等工具中查看完整的 trace + logs
- ✅ **無需手動傳遞或提取 trace ID**，完全自動化

### 3. 驗證 Trace Context 關聯

您可以在 Jaeger UI 中驗證：

1. 打開 Jaeger UI（通常是 `http://localhost:16686`）
2. 搜尋 service name: `golang-logging-demo`
3. 選擇任一 trace，查看 spans
4. 點擊 "Logs" 標籤，會看到關聯的所有日誌
5. 日誌會按照時間順序顯示，並標註所屬的 span

**範例視覺化：**
```
Trace: a3f8d9e2b7c4a1b5c6d7e8f9a0b1c2d3
├─ Span: main-operation (f1e2d3c4b5a69788)
│  └─ Log: "Application started" @ 13:31:19.124
│
├─ Span: database-query (b9c8d7e6f5a49382)
│  ├─ Log: "Warning message" @ 13:31:19.125
│  └─ Log: "Error occurred" @ 13:31:19.126
│
└─ Span: http-request (c7d6e5f4a3b29271)
   └─ Log: "Processing request" @ 13:31:19.127
```

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

## 測試環境

本程式預設連接到 `localhost:4317`。您已有測試環境運行在此端點，執行程式即可開始發送日誌到 Collector。

建議配置 Collector 使用 `logging` exporter 以在 console 查看收到的日誌。

## 授權條款

MIT License

