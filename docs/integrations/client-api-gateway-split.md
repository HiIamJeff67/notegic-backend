# ClientGateway / APIGateway 分流規格

這份文件記錄 Phase 5 的 gateway 邊界與 API key 認證決策，供 backend、contract generator 與前端文件頁面共同使用。Tutorial（例如 quick start、overview 或 UI 操作步驟）不放在這裡。

## 目標邊界

| Surface | 認證 | 用途 | 對外 contract |
| --- | --- | --- | --- |
| ClientGateway | `JWTMiddleware()`（access/refresh token） | Web/Client 現有流量與 client-only routes | user/client public |
| APIGateway | edge `KeyMiddleware()` + Core `APIKeyMiddleware()` | 外部整合與公開 API | integration public |
| RealtimeGateway | 現有 websocket ticket/connection 認證 | realtime transport | realtime public |

兩個 gateway 都只能透過 Core delegation credential 呼叫 Core。delegation claims 會攜帶：

- `gatewaySource`: `client` 或 `api`
- `authMethod`: `jwt` 或 `api-key`
- `apiKeyId`: API key 的資料庫 ID；只放 ID，不放明文 key

舊 delegation token 沒有上述欄位時，Core 會相容地視為 `client/jwt`。

## API key 儲存規則

`APIKeyTable` 只儲存 `keyHash`（SHA-256 hex digest），不儲存 secret。建立 API key 時只回傳一次完整 secret；列表 API 僅回傳 `publicId`、`name`、`keyPrefix`、狀態、`createdAt`、`lastUsedAt`、`expiresAt`，不回傳 hash 或 secret。

必要欄位：`id`、`publicId`、`userId`、`name`、`keyPrefix`、`keyHash`、`lastUsedAt`、`expiresAt`、`revokedAt`、`createdAt`、`updatedAt`。key 必須可撤銷、可設定到期時間，並以 user ownership 隔離；沒有 shared key 概念。

建議傳輸格式：

```http
X-API-Key: nzy_<base64url-secret>
```

raw key 僅允許出現在 request header 的短暫生命週期，禁止寫入 log、trace、metrics、Redis value 或 error response。

ClientGateway 以既有 JWT cookie session 提供 API key 管理：

| Method | Route | 用途 |
| --- | --- | --- |
| `POST` | `/api/v1/me/api-keys/create` | 建立 key；完整 secret 只在這次 response 出現 |
| `GET` | `/api/v1/me/api-keys/` | 列出目前使用者的 key metadata |
| `DELETE` | `/api/v1/me/api-keys/:api-key-id` | 撤銷目前使用者擁有的 key |

建立與撤銷只接受 ClientGateway 的 JWT flow；APIGateway 不提供 key management route。

## 驗證流程

1. APIGateway 先以 `KeyMiddleware()` 檢查 header 存在與基本格式。
2. Core `DelegationMiddleware()` 驗證 gateway signature、operation、request ID 與 permission claims。
3. Core `APIKeyMiddleware()` 對 header secret 做 hash，先查 cache、再 fallback 到 `APIKeyRepository`。
4. 命中且未撤銷/未過期後，載入 user，寫入 `UserId`、`UserPublicId`、`GatewaySource=api`、`AuthMethod=api-key`、`APIKeyId` context。
5. Core service 繼續使用既有 actor context；API request 不走 browser `AuthMiddleware()`。

共享 routes 以 request-level `EitherMiddleware(clientHandlers, apiHandlers, condition)` 選擇分支；condition 必須讀取已驗證的 `GatewaySource`，不能由 process startup 的 bool 決定。client-only route 不註冊到 APIGateway allowlist。

## Rate limit

APIGateway 沿用 `UnauthorizedRateLimitMiddleware()`。主要 partition 永遠是 `ctx.ClientIP()`，用來防止同一來源在大量建立/猜測 key 時繞過限制。API key ID 可以作為輔助觀測或第二層已授權 quota，但不能取代 IP bucket，也不能把 raw key 當作 Redis key 或 metrics label。

## Router 組合規則

目前 APIGateway v1 只註冊以下九個 resource domains：RootShelf、SubShelf、Material、BlockPack、Block、Station、Routine、RoutineTask、RoutineTag。其餘 auth、user/account、notification、realtime、GraphQL 與 static routes 只註冊在 ClientGateway。

總 router 只負責傳遞 repository/service dependencies。`AuthMiddleware()`、`APIKeyMiddleware()`、`EitherMiddleware()` 等應在各 `*_router.go` 就地組合，讓 route allowlist 與認證規則靠近 endpoint；目前每個可共享 domain router 都依已驗證的 delegation source 就地建立 selector。

ClientGateway 與 APIGateway 各自擁有自己的 route registration、Core adapter、Redis cache、rate limiter、configuration、transport 與 tests：ClientGateway 使用 JWT cookie flow；APIGateway 以 `KeyMiddleware()` 設定 API source，再由 Core API-key middleware 完成 cache/DB fallback。APIGateway 是獨立 module，command 位於 `runtimes/apigateway/commands/`，使用 `API_GATEWAY_LISTEN_ADDRESS`，不 import ClientGateway runtime source。

本地 Compose 會讓 ClientGateway 綁定 `7777`、APIGateway 綁定 `7780`；目前 Nginx 的預設 `/` upstream 仍指向 ClientGateway，API service 可透過獨立 port/host 配置接入。正式環境若要使用不同網域，應在 Nginx/Ingress 以 host rule 將該網域導向 `notegic-api-gateway:7780`。

現有 `Reposition` helper 已支援把 fronts、shared/default middlewares 與 route-specific backs 依序組合；兩個 gateway 的 route 已完成 `Reposition` 命名遷移。

## Contract 與文件輸出

最終目錄規劃：

```text
contracts/client-gateway/v1/public/
contracts/api-gateway/v1/public/
contracts/realtime-gateway/v1/public/
```

三個 runtime 都有自己的 `public/` contract；`api-gateway/v1/public/` 是給外部整合、CLI 與使用者自有 server 的 APIGateway integration contract，`client-gateway/v1/public/` 是給 Web/client 使用的 user contract，`realtime-gateway/v1/public/` 是給 realtime client 使用的 transport contract。generator 仍必須由 route allowlist 驅動，避免把 client-only 或 internal route 意外發布到 APIGateway。

## 驗收重點

- ClientGateway 的既有 cookie/JWT 行為不變。
- APIGateway 沒有有效 API key 或呼叫未公開 route 時一律拒絕。
- Core 能取得既有 actor context，且 API flow 不依賴 access/refresh cookie。
- 未授權限流仍以 client IP 為主。
- 明文 API key 不出現在任何持久化資料、log、trace、metrics 或 response。
- generator、Postman/examples、Makefile 與 package tests 都能由 CI 重現。
