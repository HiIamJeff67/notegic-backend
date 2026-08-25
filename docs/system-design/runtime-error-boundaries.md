# Runtime Error Boundaries

各 runtime 的錯誤語意由該 runtime ownership；跨 runtime 傳輸使用
`contracts/types/exceptions` 的通用 envelope。runtime-local `exceptions/` 只提供
domain helper，helper methods 直接回傳 shared contract exception，不把 local helper
type 暴露給 service 或 transport。

```text
runtime service / data / worker
        │ ordinary error 或 local helper 產生的 contract Exception
        ▼
HTTP / Kafka / external transport boundary
        │ public/event boundary handling
        ▼
contracts/types/exceptions.Exception
```

| Runtime | Local error ownership | Boundary mapping |
| --- | --- | --- |
| Core | `internal/core/exceptions/` domain factories | Core Gateway endpoint／transport response |
| DurableJob | `internal/durablejob/exceptions/` execution helpers | RoutineTask executor 讀取 shared exception 的 reason/retryable，再映射成穩定的本地結果與資料庫狀態 |
| Email | `internal/email/exceptions/` renderer、queue、delivery helpers | Core email Kafka consumer 讀取 shared exception 的 transient/schema 語意 |
| RealtimeGateway | `internal/realtimegateway/exceptions/` cache/data helpers | API/protocol boundary 使用 shared exception；rate limiter 可將 cache failure 視為 unavailable |
| Notification | `internal/notification/exceptions/` event、payload、repository、request helpers | Gateway endpoint 使用 shared exception 並轉成 public contract |
| YjsWorker | TypeScript native `Error` 與 versioned protocol error code | Kafka／WebSocket protocol boundary |

## 原則

- Local error 不得依賴另一個 runtime 的 source package。
- 非 HTTP 層不應依賴 `HTTPStatusCode`、Gin response 或 public response writer。
- `contracts/types/exceptions` 只保留可跨 runtime 傳遞的欄位與安全 envelope，不放 domain factory。
- Gateway／RealtimeGateway 的 transport boundary 才負責 `ToPublic()`、HTTP status 與 response body。
- Kafka consumer／producer 以穩定的 schema error、transient error、retryable 欄位傳遞失敗，不將 Go error 或 HTTP status 序列化進 event。
- Runtime-local helper methods 使用 `WithOrigin(err)` 保留原始 cause，讓 owner runtime 的 logger 與測試可以取得診斷資訊。
- Runtime-local exception package 遵循 Core 的 domain factory pattern：一個 owned domain 一個檔案，使用 `New<Runtime>Exception(domain)` 建立可注入的 helper instance，再呼叫 named methods；每個 named method 回傳 shared `*exceptions.Exception`。不建立 package-level domain instance，也不建立通用的 `New(reason, ...)` 或 `errors.go`。
- 當同一 runtime domain 同時涵蓋 renderer、delivery、payload、request 等不同責任時，`exception.go` 只保留 helper 基底，其餘 named methods 應依責任拆到對應檔案。

## 測試

Local error package 測試分類、retryability 與 cause preservation；service 測試確認錯誤分類在 boundary 前保持不變；transport 測試確認 mapping 後的 public envelope 不洩漏 origin 或 details。
