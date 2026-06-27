# ADR 0002: Activity の output_log 相関を SessionCorrelator + Command に分離

## Status

Accepted（grill-with-docs セッションで合意）

## Context

- VRChat `output_log.txt` から User encounter・プレイ時間・ワールド表示名を得る **Output log ingest** がある（用語は [`CONTEXT.md`](../../CONTEXT.md)）
- ログ行のパース（`LogParser`）と、インスタンス／ワールド文脈の付与（`sessionWorldID` / `pendingDestination` / `lastLeft*` 等）は別の関心事だが、`ActivityEventHandler`（infrastructure）に混在していた
- 相関ルールは VRChat の行順序に依存する（例: `OnPlayerLeftRoom` を SessionEnd にしない、`Destination set` 直後の `OnPlayerLeft` は旧インスタンスに帰属、ファイル境界で状態をリセット）
- 相関のテストは `activity_handler_test.go` にあるが、ActivityUseCase + リポジトリスタブ経由でしか検証できず、永続化とルールが結合している
- `AvatarSwitch` / `VideoPlayback` はパースされるが Activity では未消費（今回のスコープ外）

## Decision

1. **SessionCorrelator** を `internal/domain/activity` に置く。I/O なし。`ParsedEvent` を受け取り **fine-grained Domain command** の列を返す
2. **Command 実行**は `ActivityUseCase.ApplyCommand(ctx, cmd)` に集約する（既存の `RecordEncounterAt` 等を内部で呼ぶ）
3. **ファイル境界**は infrastructure の ingest オーケストレーターが、新ファイルを offset 0 から読む直前に `correlator.Reset()` を呼ぶ。Correlator はファイルパスを知らない
4. **ファイル末尾の後処理**（`CloseOpenEncountersAtLastLogLine` / `CloseOpenPlaySessionAtLastLogLine`）は行単位 correlator の外とし、bootstrap 完了時にオーケストレーターが明示的に呼ぶ
5. **UI 通知**（遭遇ログ変更）は Wails 固有の関心とし、correlator は知らない。Adapter が encounter 系 command 適用後に `NotifyEncounterLogChanged` を実行する（bootstrap 中は抑制）
6. **未消費の ParsedEvent**（`AvatarSwitch` / `VideoPlayback`）はパーサーに残し、correlator は空 command を返す（スコープ外）
7. **移行**は 2 段階 PR: PR1 で correlator・command・`ApplyCommand`・domain テスト（挙動不変）、PR2 で adapter 配線と `ActivityEventHandler` 相関ロジック削除
8. **ponytail #138（2026-06）**: sealed `ActivityCommand` interface は廃止。command は具象 struct のまま、correlator は `[]any` を返し `ApplyCommand(ctx, any)` の type switch で永続化する

### Command 型（fine-grained、想定）

- `EndPlaySession`, `StartPlaySession`
- `CloseOpenEncountersAt`
- `RecordEncounterJoin`, `RecordEncounterLeave`（`RecordEncounterAt` の join/leave 分割）
- `UpsertWorldVisit`, `UpsertWorldRoomName`
- `NotifyEncounterLogChanged`（adapter 付与または Apply 内 no-op + adapter が EventsEmit）

## Consequences

### 正

- 相関ルールをリポジトリなしの table-driven テストで検証できる
- `ActivityEventHandler` が薄い adapter になり、infrastructure と domain の境界が明確になる
- 新しいログ行パターン追加時、パースと相関の変更箇所が分離される

### 負

- command 型と `ApplyCommand` の switch が増える（fine-grained の代償）
- PR2 完了まで、一時的に旧 handler と新 correlator が並存する（PR1 期間）
- bootstrap のオーケストレーション（checkpoint・Reset・末尾クローズ）は `app.go` に残る

## Implementation

（PR1 / PR2 で実施）

- `internal/domain/activity/session_correlator.go` — 相関状態マシン
- `internal/domain/activity/commands.go` — command 型
- `internal/domain/activity/session_correlator_test.go` — `activity_handler_test.go` シナリオの移植
- `internal/usecase/activity_usecase.go` — `ApplyCommand`
- `internal/infrastructure/logwatcher/activity_ingest_adapter.go`（名は実装時に確定）— tail / bootstrap から correlator + usecase
- `app.go` — `correlator.Reset()`、`NotifyEncounterLogChanged` 抑制、末尾後処理
