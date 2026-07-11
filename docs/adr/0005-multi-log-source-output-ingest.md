# ADR 0005: 複数 output_log の Log source 並行 ingest

## Status

Accepted（grill-with-docs セッションで合意、Issue #158）。  
Amendment Accepted（grill-with-docs、Issue #163–#165）— 下記 Decision 11–13。  
Amendment Accepted（grill-with-docs + grill-review-ready、単一ファイル監視撤廃）— 下記 Decision 14–16（Decision 1 / 11 / 12 の該当部分を改訂）。

## Context

- VRChat を複数起動すると、ログフォルダ内に複数の `output_log*.txt` が同時に増加する
- 従来の `OutputLogWatcher`（ディレクトリモード）は `ResolveLatestOutputLogFile` で **ModTime 最新の 1 ファイルだけ** tail し、より新しいログが現れると旧 tail を捨てる（Issue #158）
- `ActivityIngestAdapter` は **SessionCorrelator が 1 つ**、`EndPlaySession` / `CloseOpenEncountersAt` は **グローバル**（他クライアントの open 行を誤って閉じうる）
- `ActivityLogCheckpoint` は単一ファイル + offset のみ
- ADR 0002 で SessionCorrelator はファイルパスを知らない純粋相関マシンと決めている
- PR #160 以降、本番のディレクトリ監視は `MultiOutputLogWatcher` のみだが、`OutputLogWatcher` に旧 latest-only ディレクトリモードが残り、Log replay と bootstrap の handler／finalize 経路が分岐していた（Issue #163–#165）
- Decision 1 で単一ファイル直接指定を残したが、VRChat は起動ごとに `output_log_….txt` を新規作成するため実務上ほぼ害であり、本番経路の分岐と UI（ファイル参照）が読み手負荷になっている。grill-with-docs で **ディレクトリのみ**へ改訂する

用語は [`CONTEXT.md`](../../CONTEXT.md) の **Activity** セクション（**Log source**、**Log replay**、**VRChat instance key**、**Log rotation**、**Log rotation handoff**、**Log source stall**）を正本とする。

## Decision

1. **監視モード**（~~単一ファイル直接指定は従来どおり~~ → **Decision 14 で改訂**）: 設定が **ログフォルダ（ディレクトリ）** のとき `MultiOutputLogWatcher` を使い、サイズ増加中の全 `output_log*.txt` を並行 tail する
2. **Log source**: ingest・finalize のスコープ単位。識別子は正規化したログファイルの絶対パス。VRChat instance key（`wrld_…:room~type`）とは別概念
3. **パイプライン分離**: Log source ごとに `ActivityIngestAdapter` + `SessionCorrelator` を 1 組ずつ。`log_source_path` は adapter が保持し usecase 呼び出し時に注入する（correlator / domain command には載せない。ADR 0002 維持）
4. **永続化**: `user_encounters` と `play_sessions` の両方に `log_source_path` を追加。`play_sessions` には VRChat instance key 用の `instance_id` も追加し `StartPlaySession` で永続化する。スコープ付き `EndPlaySession` / `CloseOpenEncountersAt` は **一致する `log_source_path` の open 行のみ**対象
5. **既存行**: `log_source_path IS NULL` は backfill しない（誤帰属回避）。スコープ付き操作は NULL 行を触らない。**VRChat 全終了時**は NULL 含む全 open 行を finalize（現行グローバル finalize の延長）
6. **アクティブ判定**: ポーリング間でファイルサイズが増えたら tail 開始・継続
7. **Log source stall**: 60 秒サイズ増加なし → tail goroutine 停止 + checkpoint 保存。**finalize はしない**（ワールド滞在中のログ沈黙による誤退室回避）
8. **Log rotation handoff**: 旧ファイル増加停止かつ **別の `output_log*.txt` が増加開始**したとき、旧 Log source を **即 finalize**（60 秒 stall を待たない）。新ファイルは新 Log source として offset 0（または checkpoint）から replay → correlator Reset → tail。複数ファイルが **同時に増加**している間は handoff しない
9. **Checkpoint**: `ファイルパス → {byteOffset, vrChatLineTime}` の map。既存単一 JSON は初回読み込みで map に移行。Bootstrap は map 内各ファイルを保存 offset から再開、map 外の増加中ファイルは offset 0 から ingest 後 tail
10. **VRChat 全終了時**: tail 対象だった **全 Log source**（stall 停止済み含む）をそれぞれ finalize。最新 1 ログのみは廃止
11. **`OutputLogWatcher` は単一ファイル専用**（Issue #165、~~ユーティリティとして `ResolveLatest` を残す~~ → **Decision 14–16 で改訂**）: ディレクトリ向け latest-only モードは削除する。ディレクトリ監視は `MultiOutputLogWatcher` のみ
12. **Log replay は Activity のみ**（Issue #163）: bootstrap（および旧単一ファイル経路のローテーション／truncate 後 replay）は `ActivityIngestAdapter` のみ。Friend joined などの automation は **live tail** に限る（過去行の二重発火回避）。replay と bootstrap は同じコードパス（または明示分岐）にする。単一ファイル経路廃止後は bootstrap が主（**Decision 14**）
13. **全終了 finalize は 2 段**（Issue #164）: (a) checkpoint の path ∪ ディレクトリ内 `output_log*.txt` を 1 集合として per-source finalize（ファイル末尾時刻優先）、(b) 続けてグローバル `FinalizeAllOpenActivity` で `log_source_path IS NULL` を含む残りを掃除。checkpoint 列挙とディレクトリ列挙の二重ループは統合する
14. **監視モードはディレクトリのみ**（Decision 1 / 11 を改訂）: 設定・validate・保存はログフォルダ（または空＝既定フォルダ）のみ。単一ファイル直接指定の監視経路は廃止する。`OutputLogWatcher` は削除する。UI からファイル参照を外し、ファイルパスの保存は error（書き込まない）。空ディレクトリは設定として有効（ログ出現待ち）
15. **既存ファイル設定の移行**: watcher 起動時、設定値が通常ファイルなら親ディレクトリへ一度永続化する。親が存在しない／読めない場合は設定を空にし、既定ログフォルダへフォールバックする
16. **ユーティリティ整理とスコープ外**: `ResolveLatestOutputLogFile` は削除する（latest-only 監視の残骸）。in-place truncate の replay／finalize 強化と、パス変更後の watcher 再起動はスコープ外（Multi の truncate は offset 合わせ＋tail 停止のみを正とし、コードに `ponytail:` で意図を残す）

## Considered Options

| 論点 | 採用 | 却下した案 |
|------|------|------------|
| スコープキー | Log source（ファイル絶対パス） | VRChat instance key のみ（同一部屋の複数クライアントで衝突） |
| `log_source_path` 注入 | adapter → usecase | correlator command に埋め込み（ADR 0002 汚染） |
| 60s stall | tail 停止のみ | stall 時 finalize（AFK 誤退室リスク） |
| ローテーション finalize | 旧停止 + 新増加で即 handoff | 60s stall 待ちのみ |
| 既存 DB 行 | NULL のまま、全終了時のみグローバル掃除 | checkpoint パスで backfill |
| 監視モード（当初） | ディレクトリ Multi + 単一ファイル残置 | 単一ファイル指定も Multi に統一 |
| 監視モード（改訂） | ディレクトリのみ（ファイル経路削除） | ファイル→親へ正規化して許容、ソフト廃止 |
| 既存ファイル設定 | 起動時に親へマイグレーション | 無視して既定のみ、監視スキップ |
| ResolveLatest | 削除 | ユーティリティとして残す |
| ローテーション correlator | 新 Log source で Reset + replay | 旧状態を引き継ぎ |
| `OutputLogWatcher` dir モード | 削除（Multi のみ） | 現状維持（読み手負荷） |
| `OutputLogWatcher` 全体 | 削除 | 型温存・非公開ヘルパー化 |
| Log replay 中の automation | 発火しない | MultiHandler で replay（二重発火リスク） |
| 全終了 finalize | path 集合 1 本 + グローバル | checkpoint／dir／グローバルの 3 段並存 |
| 保存拒否の UI | Settings で明示エラー | サイレント失敗 |
| 保存 error 形 | sentinel（`errors.Is` 可能） | 文言のみ |
| パス変更後の watcher 再起動 | スコープ外（既存どおり再起動依存） | SetPathSettings で付け直し |

## Consequences

### 正

- 複数 VRChat 起動時に各クライアントの遭遇・プレイ時間の取りこぼしがなくなる
- クライアント A の Joining がクライアント B の遭遇を閉じない
- ADR 0002 の correlator 純粋性を維持したままインフラ側で並行化できる
- AFK 時の誤 finalize を Log source stall 設計で抑えられる
- 監視・replay・finalize の本番経路が読み手に一意になる（#163–#165）
- 単一ファイル経路削除後、監視モードがディレクトリ一本になり設定・UI・コードが一致する

### 負

- DB マイグレーション + checkpoint 形式変更 + watcher 差し替えのまとめての実装コスト
- `log_source_path` NULL のレガシー行は、VRChat 全終了まで open のまま残りうる（単一クライアント終了のみでは stall 停止で掃除されない。Issue 原文の 60s finalize より猶予が長い）
- ディレクトリ内の全 `output_log*.txt` をポーリングするため、古いログファイルが大量にある環境ではポーリングコストが増える（ponytail: 必要なら「増加中のみ」キャッシュで O(active) に抑える）
- Issue #158 原文（60s stall で finalize）からの仕様変更を ADR と Issue 更新で追従する必要がある
- 単一ファイル switch replay から automation を外すと、truncate 直後の過去 Join では automation が走らない（意図的。live 追記のみ）。単一ファイル経路削除後は Multi の truncate も replay しない（Decision 16）
- 既存ファイル設定のマイグレーションや保存拒否は実装コスト（Settings UI・sentinel・テスト）を増やす

## Implementation

Issue #158 参照。後続整理は Issue #165 → #163 → #164。  
単一ファイル監視撤廃は Decision 14–16（未 Issue 化なら実装前に Issue を切る）。

- `internal/infrastructure/logwatcher/multi_output_log_watcher.go`（新規）
- `internal/infrastructure/logwatcher/activity_ingest_adapter.go` — Log source バインド
- `internal/usecase/activity_usecase.go` — checkpoint map、スコープ付き End/Close
- `internal/infrastructure/sqlite/` — `log_source_path` / `play_sessions.instance_id` マイグレーション
- `app_activity_log_watch.go` — 全終了時の全 Log source finalize
- `internal/infrastructure/logwatcher/output_log_watcher.go` — 単一ファイル専用化（#165）→ **Decision 14 で削除**
- `app.go` — Log replay と bootstrap の共通化、replay は Activity のみ（#163）→ ファイル分支削除、起動時マイグレーション
- `app_activity_log_watch.go` — 全終了 finalize の path 集合統合（#164）
- `ResolveLatestOutputLogFile` — **Decision 16 で削除**
- Settings UI — ファイル参照削除、保存 error 表示、ヒント文更新
- `multi_output_log_watcher.go` truncate 分岐 — `ponytail:`（replay／finalize なし）
