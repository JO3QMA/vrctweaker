# ADR 0005: 複数 output_log の Log source 並行 ingest

## Status

Accepted（grill-with-docs セッションで合意、Issue #158）。  
Amendment Accepted（grill-with-docs、Issue #163–#165）— 下記 Decision 11–13。

## Context

- VRChat を複数起動すると、ログフォルダ内に複数の `output_log*.txt` が同時に増加する
- 従来の `OutputLogWatcher`（ディレクトリモード）は `ResolveLatestOutputLogFile` で **ModTime 最新の 1 ファイルだけ** tail し、より新しいログが現れると旧 tail を捨てる（Issue #158）
- `ActivityIngestAdapter` は **SessionCorrelator が 1 つ**、`EndPlaySession` / `CloseOpenEncountersAt` は **グローバル**（他クライアントの open 行を誤って閉じうる）
- `ActivityLogCheckpoint` は単一ファイル + offset のみ
- ADR 0002 で SessionCorrelator はファイルパスを知らない純粋相関マシンと決めている
- PR #160 以降、本番のディレクトリ監視は `MultiOutputLogWatcher` のみだが、`OutputLogWatcher` に旧 latest-only ディレクトリモードが残り、Log replay と bootstrap の handler／finalize 経路が分岐していた（Issue #163–#165）

用語は [`CONTEXT.md`](../../CONTEXT.md) の **Activity** セクション（**Log source**、**Log replay**、**VRChat instance key**、**Log rotation**、**Log rotation handoff**、**Log source stall**）を正本とする。

## Decision

1. **監視モード**: 設定が **ログフォルダ（ディレクトリ）** のときだけ `MultiOutputLogWatcher` を使い、サイズ増加中の全 `output_log*.txt` を並行 tail する。**単一ファイル直接指定**は従来どおり 1 Log source のみ（複数クライアント対象外）
2. **Log source**: ingest・finalize のスコープ単位。識別子は正規化したログファイルの絶対パス。VRChat instance key（`wrld_…:room~type`）とは別概念
3. **パイプライン分離**: Log source ごとに `ActivityIngestAdapter` + `SessionCorrelator` を 1 組ずつ。`log_source_path` は adapter が保持し usecase 呼び出し時に注入する（correlator / domain command には載せない。ADR 0002 維持）
4. **永続化**: `user_encounters` と `play_sessions` の両方に `log_source_path` を追加。`play_sessions` には VRChat instance key 用の `instance_id` も追加し `StartPlaySession` で永続化する。スコープ付き `EndPlaySession` / `CloseOpenEncountersAt` は **一致する `log_source_path` の open 行のみ**対象
5. **既存行**: `log_source_path IS NULL` は backfill しない（誤帰属回避）。スコープ付き操作は NULL 行を触らない。**VRChat 全終了時**は NULL 含む全 open 行を finalize（現行グローバル finalize の延長）
6. **アクティブ判定**: ポーリング間でファイルサイズが増えたら tail 開始・継続
7. **Log source stall**: 60 秒サイズ増加なし → tail goroutine 停止 + checkpoint 保存。**finalize はしない**（ワールド滞在中のログ沈黙による誤退室回避）
8. **Log rotation handoff**: 旧ファイル増加停止かつ **別の `output_log*.txt` が増加開始**したとき、旧 Log source を **即 finalize**（60 秒 stall を待たない）。新ファイルは新 Log source として offset 0（または checkpoint）から replay → correlator Reset → tail。複数ファイルが **同時に増加**している間は handoff しない
9. **Checkpoint**: `ファイルパス → {byteOffset, vrChatLineTime}` の map。既存単一 JSON は初回読み込みで map に移行。Bootstrap は map 内各ファイルを保存 offset から再開、map 外の増加中ファイルは offset 0 から ingest 後 tail
10. **VRChat 全終了時**: tail 対象だった **全 Log source**（stall 停止済み含む）をそれぞれ finalize。最新 1 ログのみは廃止
11. **`OutputLogWatcher` は単一ファイル専用**（Issue #165）: ディレクトリ向け latest-only モードは削除する。ディレクトリ監視は `MultiOutputLogWatcher` のみ。`ResolveLatestOutputLogFile` はユーティリティとして残してよい
12. **Log replay は Activity のみ**（Issue #163）: bootstrap および単一ファイルのローテーション／truncate 後 replay は `ActivityIngestAdapter` のみ。Friend joined などの automation は **live tail** に限る（過去行の二重発火回避）。replay と bootstrap は同じコードパス（または明示分岐）にする
13. **全終了 finalize は 2 段**（Issue #164）: (a) checkpoint の path ∪ ディレクトリ内 `output_log*.txt` を 1 集合として per-source finalize（ファイル末尾時刻優先）、(b) 続けてグローバル `FinalizeAllOpenActivity` で `log_source_path IS NULL` を含む残りを掃除。checkpoint 列挙とディレクトリ列挙の二重ループは統合する

## Considered Options

| 論点 | 採用 | 却下した案 |
|------|------|------------|
| スコープキー | Log source（ファイル絶対パス） | VRChat instance key のみ（同一部屋の複数クライアントで衝突） |
| `log_source_path` 注入 | adapter → usecase | correlator command に埋め込み（ADR 0002 汚染） |
| 60s stall | tail 停止のみ | stall 時 finalize（AFK 誤退室リスク） |
| ローテーション finalize | 旧停止 + 新増加で即 handoff | 60s stall 待ちのみ |
| 既存 DB 行 | NULL のまま、全終了時のみグローバル掃除 | checkpoint パスで backfill |
| 監視モード | ディレクトリのみ Multi | 単一ファイル指定も Multi に統一 |
| ローテーション correlator | 新 Log source で Reset + replay | 旧状態を引き継ぎ |
| `OutputLogWatcher` dir モード | 削除（Multi のみ） | 現状維持（読み手負荷） |
| Log replay 中の automation | 発火しない | MultiHandler で replay（二重発火リスク） |
| 全終了 finalize | path 集合 1 本 + グローバル | checkpoint／dir／グローバルの 3 段並存 |

## Consequences

### 正

- 複数 VRChat 起動時に各クライアントの遭遇・プレイ時間の取りこぼしがなくなる
- クライアント A の Joining がクライアント B の遭遇を閉じない
- ADR 0002 の correlator 純粋性を維持したままインフラ側で並行化できる
- AFK 時の誤 finalize を Log source stall 設計で抑えられる
- 監視・replay・finalize の本番経路が読み手に一意になる（#163–#165）

### 負

- DB マイグレーション + checkpoint 形式変更 + watcher 差し替えのまとめての実装コスト
- `log_source_path` NULL のレガシー行は、VRChat 全終了まで open のまま残りうる（単一クライアント終了のみでは stall 停止で掃除されない。Issue 原文の 60s finalize より猶予が長い）
- ディレクトリ内の全 `output_log*.txt` をポーリングするため、古いログファイルが大量にある環境ではポーリングコストが増える（ponytail: 必要なら「増加中のみ」キャッシュで O(active) に抑える）
- Issue #158 原文（60s stall で finalize）からの仕様変更を ADR と Issue 更新で追従する必要がある
- 単一ファイル switch replay から automation を外すと、truncate 直後の過去 Join では automation が走らない（意図的。live 追記のみ）

## Implementation

Issue #158 参照。後続整理は Issue #165 → #163 → #164。

- `internal/infrastructure/logwatcher/multi_output_log_watcher.go`（新規）
- `internal/infrastructure/logwatcher/activity_ingest_adapter.go` — Log source バインド
- `internal/usecase/activity_usecase.go` — checkpoint map、スコープ付き End/Close
- `internal/infrastructure/sqlite/` — `log_source_path` / `play_sessions.instance_id` マイグレーション
- `app_activity_log_watch.go` — 全終了時の全 Log source finalize
- `internal/infrastructure/logwatcher/output_log_watcher.go` — 単一ファイル専用化（#165）
- `app.go` — Log replay と bootstrap の共通化、replay は Activity のみ（#163）
- `app_activity_log_watch.go` — 全終了 finalize の path 集合統合（#164）
