# Yjs Collaboration and Persistence Design

## Scope

This document defines the backend-owned BlockPack collaboration model:
document identity, schema evolution, durable persistence, projection, and the
Gateway-to-worker boundary. Public connection framing is specified by
[Realtime Protocol Design](realtime-protocol.md); HTTP ticket behavior is
specified by [Realtime Editor API Design](../api-route-design/realtime-editor-api.md).

Notegic's BlockPack collaboration document is shared by BlockNote, the public
Realtime Gateway, the Node Yjs worker, Go persistence, and Block projection.
No implementation may independently change room, fragment, schema, or sequence
semantics.

## Document Identity

| 項目 | 固定值 |
| --- | --- |
| channel type | `BlockPack` |
| channel id | BlockPack UUID |
| room name | `block-pack:{blockPackId}` |
| Y.XmlFragment name | `document-store` |
| document schema id | `notegic.blocknote` |
| initial document schema version | `1` |

Go constants: `YjsBlockPackRoomPrefix`、`YjsBlockPackFragmentName`、`YjsBlockPackSchemaId`、`YjsBlockPackSchemaVersion`。

`document-store` 必須顯式傳給 BlockNote collaboration configuration，例如 `doc.getXmlFragment("document-store")`。不得依賴 BlockNote Yjs utility 的預設 fragment name。

一個 BlockPack 對應一份 logical Yjs document。`Y.Doc` 是 Node worker 在 active room 的記憶體 runtime object；它不是資料庫 entity，也不會直接提供給外部 editor 或 Go service。

## BlockNote Schema

schema version `1` 的 block type manifest 與目前後端 `BlockType` 對齊：

`paragraph`、`heading`、`quote`、`bulletListItem`、`numberedListItem`、`checkListItem`、`toggleListItem`、`image`、`video`、`audio`、`file`、`table`、`codeBlock`、`mathBlock`、`diagram`、`calendar`。

所有 editor runtime 必須以單一 `BlockNoteSchema` factory 建立 editor、Yjs import/export 與 server-side projector 使用的 schema。Node worker 使用相同的 block/inline/style manifest；Go 不解析 Yjs tree，也不自行重建 BlockNote document。

Node projector 使用 `@blocknote/core/yjs` 的 `yXmlFragmentToBlocks`，並明確讀取 `document-store` fragment。它的 schema 排除 `divider`，因為目前後端 `BlockType` 未支援此 block type；editor schema 也不得建立 `divider`。

新增、刪除或變更 block props、inline content、style schema 都是 schema migration，不是一般 feature flag。

## Version Policy

`schemaVersion` 是 logical document version，不等同 npm package version，也不等同 Realtime protocol version。

| 規則 | 語意 |
| --- | --- |
| 新 document | 以目前 supported schema version 建立，初始為 `1`。 |
| 讀取 | client 與 Node worker 僅可開啟自己明確支援的 version。 |
| 向後相容變更 | 保持 version，僅限舊 reader 能無損理解的新增 optional data。 |
| 不相容變更 | 建立新 version，Node worker 對完整 Y.Doc migration，產出新 snapshot 後才切換。 |
| 投影 | projector 的 schema version 必須與 source document version 相同。 |

目前所有 document 都使用 `YjsBlockPackSchemaVersion = 1`。未來若需要同時支援多個 document schema version，才新增 per-document `SchemaVersion`。

## Persistence And Sequence

durable Yjs truth 是 `BlockPackYjsDocument.Snapshot` 加上尚未 compact 的 `BlockPackYjsUpdate` tail。Snapshot 是 Yjs encoded state update，`StateVector` 是同一個 snapshot 的 encoded state vector；active `Y.Doc` 只是這份 durable truth 的 memory materialization。

每個 BlockPack 必須在建立它的同一筆 transaction 內建立唯一的 `BlockPackYjsDocument`；讀取、append 與 projection 路徑不得 lazy create document。若建立來源已具有 BlockNote blocks（例如 RoutineTask 的 CreateBlockPack），Go 必須先要求 Node worker 以相同 schema 和 `document-store` fragment 產生 initial Snapshot/StateVector，並在同一筆 transaction 寫入 document 與 `BlockTable` projection。不得只寫入 `BlockTable` 而留下空的 Yjs document。

`BlockTable` 是 Yjs document 的 materialized projection，Block 不支援 soft delete。projection 對不再存在於 document 的 block 使用實體 `DELETE`；BlockPack soft delete 時則保留它的 Blocks，還原 BlockPack 後可直接重用既有 projection。

Block REST read endpoints 與 GraphQL `searchBlocks` 都只讀 `BlockTable` projection；它們不得用於建立或回填 active `Y.Doc`。BlockPack REST read response 會帶 `lastUpdateSequence`、`compactedUntilSequence`、`projectedUntilSequence` 與 `isProjectionCurrent`。`isProjectionCurrent = false` 表示 read model 落後 durable document，協作狀態一律仍以 Yjs channel 為準。

| 欄位 | 語意 |
| --- | --- |
| `UpdateSequence` | 單一 BlockPack 內 append-only 的 update 序號，從 `1` 起，永不重用。 |
| `LastUpdateSequence` | 該 BlockPack 已接受的最高 update sequence；不得回退。 |
| `CompactedUntilSequence` | 已被目前 Snapshot 吸收的最高 sequence；不得回退。 |
| `ProjectedUntilSequence` | BlockTable 已成功投影的最高 sequence；document-level checkpoint，初始為 `-1`，且不得回退。 |

不變條件：`0 <= CompactedUntilSequence <= LastUpdateSequence`、`-1 <= ProjectedUntilSequence <= LastUpdateSequence`。空 document 的 durable update/compaction sequence 都是 `0`，未投影時 `ProjectedUntilSequence` 是 `-1`。

compaction 在 Node worker 重建完整 Y.Doc 後執行：它讀取 snapshot 與 update tail、合併到 runtime Y.Doc、寫入新的 Snapshot/StateVector，最後將 `CompactedUntilSequence` 推進到被吸收的最高 sequence。Go 不執行 CRDT merge。

room cold start 固定依序執行：建立空 `Y.Doc`、套用非空 Snapshot、再套用 `CompactedUntilSequence < update_sequence <= LastUpdateSequence` 的 tail。`LastUpdateSequence` 是最新 accepted update，不是 snapshot tail 的查詢起點。

## Public Connection And Capability

root WebSocket authentication 以 connection ticket 取代 access-token middleware；它只識別 user，不授權任何 BlockPack。每個 BlockPack channel 都在 subscribe 驗證 capability ticket 後才建立。

ticket claims 的最小集合：

```json
{
  "sub": "user public UUID",
  "jti": "ticket trace UUID",
  "channelType": "BlockPack",
  "channelId": "blockPack UUID",
  "permission": "read or write",
  "realtimeProtocolVersion": 1,
  "schemaVersion": 1,
  "roomAdmissionPolicyVersion": 1,
  "roomAdmissionEnforcementStrategy": "reject-new-subscriber",
  "maximumSubscribers": 0,
  "documentQuotaPolicyVersion": 1,
  "maximumBlockCount": 0,
  "iat": 0,
  "exp": 0
}
```

RealtimeGateway 負責驗證 connection/channel ticket，以及兩者的 user、channel type 與 BlockPack id 是否相符；Node worker 只信任已驗證後送出的 attach message。ticket 是短效 signed capability；RealtimeGateway 以 Redis `SET NX` 與到期 TTL 原子消費 `jti`，因此 ticket 在所有 RealtimeGateway instance 都只能使用一次。admission 與 document quota claims 是 Core 簽發的完整 policy snapshot；RealtimeGateway 以 verified admission claims 執行 atomic lease admission，並將 verified Block quota 交給 worker。worker 只有在 materialize authoritative Y.Doc 且註冊 subscriber 後才回傳 `attached`。每筆更新先套用 validation Y.Doc，超過 `maximumBlockCount` 時整筆拒絕，不進入 authoritative Y.Doc、persistence、projection 或 broadcast。Core 的權限與 resource lifecycle 變更透過 outbox、Kafka 與 RealtimeGateway fanout 主動撤銷已存在的 channel。

## Cross-Service Frames

Public WebSocket JSON control frames and binary headers are defined in
[Realtime Protocol Design](realtime-protocol.md). Go-to-worker internal binary
frames always include `connectionId`, `connectorChannelId`, `channelType`, and
`channelId`; raw Yjs updates are never Base64-encoded or rewritten as JSON
block events.

internal attach/detach 是 idempotent。worker reconnect 後，RealtimeGateway 為其所屬 active channels replay 帶有 quota policy 的 attach；worker 會先向 Core cold-load snapshot + tail，materialize `Y.Doc` 並回傳 `attached`，RealtimeGateway 才向 client 回覆 `subscribed`。合法 raw Yjs updates 會以同一個 BlockPack room 為單位暫存並使用 `Y.mergeUpdates()` 合併為一筆 persistence batch。batch 會以 Kafka command 非同步送往 Core；editor message path 只等待 Kafka producer 的 replicated broker ACK，不等待 Core transaction。broker ACK 後 worker 即可釋出 room 的下一個 batch 並廣播 merged raw Yjs update；Core 的 application reply 會在背景更新持久化 watermark，若 Core 最終拒絕或逾時則要求 room resync。

每個 persistence batch 有只供 Go/worker 使用的 UUID idempotency key。Go 以 `(block_pack_id, persistence_batch_id)` 保證 internal WebSocket retry 不會建立重複 update row；同一 batch 的多個來源 connection 不可任意挑選其中一個寫入 `OriginConnectionId`，必須保留為 `NULL`。broker ACK 只代表 command 已被 Kafka 接受，不代表 PostgreSQL 已提交；若 Core application reply 最終失敗，worker 對 room 所有 subscriber 發出 `resync_required`，並以 cold-load 的 authoritative watermark 修正本地 optimistic sequence。

batch flush 條件由 worker constants 控制：trailing debounce、maximum wait、raw update count、raw payload bytes、最後 subscriber detach 與 graceful worker shutdown。worker 在 broker ACK 後以每個 batch 一個 optimistic sequence 維持 room 順序；Core application reply 回來後會以 authoritative `UpdateSequence` 校正，重連 cold-load 時永遠以 Core watermark 為準。

## Projection

Node worker 是唯一的 Yjs CRDT merge owner，也是 Y.Doc -> BlockNote blocks conversion owner。它以 current `schemaVersion` 將 active Y.Doc 轉換為 canonical BlockNote block tree，再送出 projection payload 給 Go。

projection payload 最小欄位：

```json
{
  "schemaId": "notegic.blocknote",
  "schemaVersion": 1,
  "projectedSequence": 42,
  "blocks": []
}
```

projection 使用 private Go-to-worker internal frame `apply-block-projection`；BlockPack identity 取自 frame header 的 `channelId`，不是 payload。Go 僅在 payload 的 `schemaId`/`schemaVersion` 受支援、target sequence 不回退，且所屬 BlockPack 有效時寫入 Block projection，成功時回覆 `block-projection-applied`，否則回覆 `block-projection-failed`。bulk apply、anti-regression transaction 與 accounting 以 document-level sequence 為準；不新增 per-block ordering/search/hash metadata，除非實際 read requirement 證明需要。

worker 僅在 room 沒有尚未持久化的 update 時，將目前 `LastUpdateSequence` 的 document 投影。它對更新 burst 做 debounce，且每個 room 同時最多一筆 projection；只有收到 `block-projection-applied` 後才推進 in-memory `ProjectedUntilSequence`。失敗不會前進 checkpoint，並以 retry delay 重試。

外部 editor 不讀取 `BlockPackYjsDocument` 或 `BlockPackYjsUpdate` rows，也不自行合併 update tail；加入 room 時由 Node worker 從 snapshot + tail 恢復 Y.Doc，再以標準 Yjs sync protocol 完成同步。

## Core Maintenance Coordination

Core owns the Yjs maintenance worker。Core 在建立 BlockPack Yjs document，或
同一筆 transaction 接受新的 Yjs update 後，將只包含 document watermark、大小與
sequence 的 `YjsMaintenanceHint` 寫入 Core transactional outbox。hint 不攜帶
snapshot、state vector 或 raw Yjs binary。

Core worker 消費 hint 後以 BlockPack UUID 作為 Kafka partition key，在記憶體中
coalesce 同一個 BlockPack 的最新 hint，依 uncompacted update count、projection
lag 與 document age 排序，直接向 Yjs Worker 提出 compact/project command。Core
worker 直接消化 result 並 bounded retry；DurableJob 不參與這條 maintenance path，
也不持有 Core 的 Yjs repository 或 maintenance client。

這條 maintenance path 是非同步的：Core domain transaction 不等待 Yjs Worker，
也不把大型 document payload 放進 Kafka。所有 consumer
都必須以 request/event UUID 與 document sequence 做 idempotency，並接受 outbox
relay 的 at-least-once delivery。
