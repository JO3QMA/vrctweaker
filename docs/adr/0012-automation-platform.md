# ADR 0012: Automation platform（event catalog・rule builder・Lua script）

## Status

Accepted（grill-with-docs で合意、[Issue #225](https://github.com/JO3QMA/vrctweaker/issues/225)）

## Context

- Automation タブは v0 として `friend_joined` → `change_status` の単一トリガー／単一アクションのみ（`AutomationView.vue` + `AutomationRule` SQLite）
- 利用者要望: スケジュール（例: 毎日 0 時）と VRChat 起動中などの条件で OS 操作（電源プラン）やプレゼンス変更を自動化したい
- 一般利用者向け GUI ビルダーと、エンジニア向け Lua の両方を同一画面で扱いたい
- Tweaker が既に持つログ tail・プロセス検知・identity 等を **Automation event** としてフックしたい
- ponytail 過去判断で EventBus は friend join 直結に簡素化されたが、複数 event 種・Lua subscribe を載せるには **カタログ付き event 配信**が必要

用語は [`CONTEXT.md`](../../CONTEXT.md) の **Automation** セクションを正本とする。

## Decision

### 1. 製品形

1. **Automation** 画面に **Automation item** を一覧。種別は **Automation rule**（`kind: rule`）または **Automation script**（`kind: script`）のどちらか一方のみ（同居・相互変換なし）
2. 評価モデルは **IFTTT 型 + イベント駆動**: **Automation trigger**（評価起点）→ **Automation condition**（AND）→ **Automation action sequence**（直列）
3. スケジュールも **Automation event**（`schedule.tick`）として同一バスに載せる
4. **Automation runtime**: v1 は **VRCTweaker プロセス起動中のみ**評価・実行。OS バックグラウンドサービスは v1 スコープ外（UI で明示）

### 2. Event catalog（v1）

| Event | 発火元 | 備考 |
|-------|--------|------|
| `friend_joined` | output_log live tail | **Log replay** では発火しない（Activity 定義に従う） |
| `schedule.tick` | Tweaker 内スケジューラ | **Schedule rule**（曜日複数 + 時分、ローカル TZ）に一致した分 |
| `vrchat.process` | プロセス状態変化 | `running` / `stopped` 等の payload |

- カタログは **追加のみ**（破壊的変更しない）。event 名と payload 形状を契約として公開
- Pipeline・Play session・ログイン状態等は **カタログ拡張 Issue** で後続
- 既存 UI の `afk_detected` は v1 カタログ外（別 Issue）

**同一分に複数 item が該当**: 有効 item を **item ID の昇順**で評価（安定順）。並列実行はしない

### 3. Action catalog（v1）

| Action | Platform | 内容 |
|--------|----------|------|
| `change_status` | 全 OS | VRChat プレゼンス（busy / ask me / join me）。既存 allowlist 維持 |
| `set_power_plan` | Windows のみ | **Power plan preset** または **Detected power plan** GUID |

- 各 action はカタログで `platforms` を宣言。非対応 OS では GUI 非表示／無効、Lua は実行時エラー
- **Automation action sequence**: 1 item あたり複数 action をリスト順に直列実行
- **Continue on error**: 未指定時は失敗で sequence 停止。action ごとに `continue_on_error: true` で続行可
- **Power plan preset** が OS 上に解決できない場合: 当該 action を **失敗**とし run log に記録（サイレントスキップしない）

### 4. Rule builder（GUI v1）

- **Automation rule builder**: セクション型カード — 「いつ（トリガー）」「もし（条件）」「したら（アクション列）」
- ドラッグ式ノードエディタは v1 スコープ外
- **Power plan selection**: 既定は preset、詳細で detected 一覧

### 5. Automation script（Lua v1）

- **Automation script API**: (1) event `subscribe`、(2) `actions.run`（action catalog 経由のみ）、(3) Tweaker 状態の**読み取り専用** API
- ファイル IO・任意 HTTP・シェル・OS 直叩きは v1 禁止。追加能力は action / 読み取り API のカタログ拡張
- script item は Lua ソースを DB に保存。起動時にサンドボックス VM で load（具体ライブラリは実装 Issue）

### 6. Run log（v1）

- **Automation run log**: 直近 **N 件**（目安 20〜50）。成功／失敗・時刻・item 名・アクション完了数
- 表示は **Redacted reproduction** 同趣旨（表示名・`usr_*` 等を載せない）
- 永続ページング・OS 通知は v1 スコープ外

### 7. ストレージ・API（方針）

- SQLite の automation テーブルを **Automation item** 形に拡張（`kind`、複数 action JSON、schedule / conditions JSON、script 本文）
- 既存 `AutomationRule` 行は `kind: rule` へマイグレーション
- Wails API: item CRUD、run log 取得、detected power plans 一覧（Windows）、カタログメタ（event/action 定義）を必要に応じて追加

### 8. v1 スコープ外（別 Issue）

[`CONTEXT.md`](../../CONTEXT.md) 各 `v1 scope` 語に加え、次は follow-up Issue とする:

| トピック | Issue |
|----------|-------|
| cron 相当スケジュール | [#226](https://github.com/JO3QMA/vrctweaker/issues/226) |
| 視覚的ブロック／ノード rule builder | [#227](https://github.com/JO3QMA/vrctweaker/issues/227) |
| OS バックグラウンドサービス | [#228](https://github.com/JO3QMA/vrctweaker/issues/228) |
| run log 永続・通知 | [#229](https://github.com/JO3QMA/vrctweaker/issues/229) |
| Pipeline 等 event 拡張 | [#230](https://github.com/JO3QMA/vrctweaker/issues/230) |
| `afk_detected` event | [#231](https://github.com/JO3QMA/vrctweaker/issues/231) |
| macOS 電源プラン action | [#232](https://github.com/JO3QMA/vrctweaker/issues/232) |
| rule → Lua エクスポート | [#233](https://github.com/JO3QMA/vrctweaker/issues/233) |
| 追加 action（音量・Webhook 等） | [#234](https://github.com/JO3QMA/vrctweaker/issues/234) |

## Considered options

| 案 | 却下理由 |
|----|----------|
| rule と script の双方向変換 | v1 工数・曖昧な部分変換。script は rule で書けないときだけ新規作成 |
| スケジュールを condition のみで表現 | event 駆動モデルと Lua subscribe が分断される |
| Lua フル権限 | デスクトップアプリでも誤スクリプトリスク。action カタログ経由に限定 |
| Automation 全体を Windows のみ | `change_status` 等は他 OS でも価値がある |
| v1 で EventBus 再導入せず直結のみ | event 種増加と Lua subscribe で直結が N×M に膨らむ |

## Failure modes（review-ready）

| 状況 | ユーザー | バックエンド |
|------|----------|--------------|
| ログ tail 経由の automation 評価／action 失敗 | **Automation run log** に記録のみ。`ElMessage` なし | Activity ingest は **fail-open**（止めない） |
| action sequence 部分成功 | run log に完了数（例: `1/2`）。item 全体は失敗（`continue_on_error` なし時） | 後続 action は実行しない |
| Lua script 実行時エラー | run log + 診断ログ。クラッシュしない | 他 item の評価は続行 |
| スケジューラ／Lua VM **init 失敗** | Automation 画面にインライン「利用不可」（Server status fetch failure と同型） | アプリ起動は続行 |
| Tweaker 未起動 | 評価なし（`CONTEXT.md` **Automation runtime**） | — |
| OS 通知・失敗トースト | v1 なし（#229） | — |

バックグラウンド実行の正本は **Automation run log**。ユーザーが Automation 画面を開いているときは run log パネルをポーリングまたは event で更新してよい。

## Return-value contract（review-ready）

### 読み取り

| メソッド | 成功 | インフラ失敗 | 備考 |
|----------|------|--------------|------|
| `ListAutomationItems` | `([]DTO, nil)` | `nil, error` | 0 件は空 slice |
| `GetAutomationRunLog` | `([]DTO, nil)` | `nil, error` | 0 件は空 slice |
| `GetAutomationRuntimeStatus` | `(DTO, nil)` | `nil, error` | `available: false` + **stable i18n key**（例: `subsystemUnavailable`）。init 失敗でも `error` は返さない |
| `ListDetectedPowerPlans` | `([]DTO, nil)` | `nil, error` | **非 Windows** は `([], nil)`（未対応は正常） |

Server status と同型: **サブシステム未初期化は DTO 状態**、DB 等の infra は `error`。

### 書き込み

| メソッド | 契約 |
|----------|------|
| `SaveAutomationItem` / `DeleteAutomationItem` | 検証失敗・DB 失敗は `error` |
| `ToggleAutomationItem` | 未知 ID は **`error`**（新 API）。既存 `ToggleAutomationRule` の silent no-op は互換維持し、マイグレーション PR で明記 |

### エラーメッセージ

- ユーザー向け: **stable i18n key** をフロントが翻訳（`automation.reason.*`）
- 診断ログ: 英語メッセージ可。表示名・`usr_*` は載せない

## Config and limits（review-ready）

v1 は **コード定数のみ**（`app_settings` / env なし）。実装と ADR の数字を一致させ、テストで定数を参照する。

| 定数 | 値 | 用途 |
|------|-----|------|
| `runLogMaxEntries` | 50 | メモリ内 run log リング上限 |
| `scheduleTickResolution` | 1 分 | Schedule rule の最小粒度。同一分に `schedule.tick` は 1 回 |
| `luaExecTimeout` | 10s | script 1 回の評価／ハンドラ上限 |
| `maxActionsPerItem` | 10 | 1 item の action sequence 上限 |
| `maxScriptBytes` | 32 KiB | Automation script 本文上限 |

Settings での変更は v1 スコープ外。

## Trust boundaries（review-ready）

| 境界 | v1 方針 |
|------|---------|
| **Automation run log（UI）** | item 名・event 種・action 結果・**VRChat 表示名**（例: 参加したフレンド）を可。**`usr_*` は UI に出さない** |
| **診断ログ（ローカル）** | event payload 全文可（開発・サポート用）。本番ビルドでもユーザーマシン上のみ |
| **Lua 読み取り API** | Tweaker キャッシュのフル DTO（内部 ID 含む）。script はユーザー自身が明示作成 |
| **公開成果物** | PR/Issue/コミットは **Redacted reproduction**（表示名・`usr_*` 不使用） |

Lua script 本文は診断ログに**全文を出さない**（失敗時は行番号 + 要約）。

## Concurrency and lifecycle（review-ready）

1. **Event 配信**: 発火元（ログ tail・スケジューラ・プロセス監視）は **buffered channel** に event を投げるのみ。評価・action 実行は行わない
2. **ワーカー**: **単一 goroutine** が channel から順に取り出し、該当 item を評価・実行（`ponytail:` キュー深度は定数化。溢れたら drop + 診断ログ）
3. **同一分の複数 item**: ワーカー内で **item ID 昇順**
4. **Shutdown**: `ctx` cancel → スケジューラ停止 → channel close → ワーカー `WaitGroup.Wait()`。進行中 Lua は `luaExecTimeout` で打ち切り
5. **Mutex**: `powercfg` / VRChat API 等の I/O は item 一覧ロックの外で実行
6. **vrchat.process** 連続発火: v1 は debounce なし（ワーカー直列で十分）

ログ ingest 経路を automation でブロックしない（fail-open）。

## Interim implementation（review-ready）

### 条件（Automation condition）v1

- GUI は **プリセット型のみ**（自由 JSON エディタなし）。保存は型付き JSON
- v1 プリセット:
  - `vrchat_running` — チェックボックス（全 trigger で利用可）
  - `friend_is` — `friend_joined` 用フレンド picker（表示名 UI、内部は VRC user ID）
- **v1 外**: OR / NOT、数値比較、Play session / Pipeline 条件、自由記述式 → **Automation script**
- 既存の生 `conditionJson` キー一致のみは **廃止方向**（マイグレーションでプリセットへ寄せるか、読み取り互換のみ）

`ponytail:` 一般条件式エンジンは v2 以降。複合ロジックは script。

## Wails / frontend contract（review-ready）

| トピック | v1 契約 |
|----------|---------|
| **Run log 更新** | バックエンドが run log 追記時に **`automation:run-log-changed`** を `EventsEmit`。`AutomationView` は `EventsOn` で再取得 |
| **Item 一覧** | 保存／削除／toggle 後に明示 `ListAutomationItems`（event 不要） |
| **エラー** | `automation.reason.*` stable key（質問 2）。5 ロケール同期 |
| **E2E mock** | `mock-wails.ts` に event 発火・新 API シグネチャを追加 |
| **Storybook** | Wails mock を story 側で上書き可能に（meta のみに依存しない） |

## Frontend async UI（review-ready）

| トピック | v1 契約 |
|----------|---------|
| **未保存編集** | Launcher 同型（**Unsaved automation edits**）。バナー + ルート離脱時確認。rule / script とも明示「保存」まで DB に書かない |
| **自動保存** | なし |
| **有効／無効 toggle** | 即時 API（編集バッファとは独立） |
| **保存失敗** | 編集内容を保持、`ElMessage.error`（i18n key） |
| **Unmount** | `EventsOn` / `GetAutomationRunLog` / 保存の `await` 後は generation ガードで `ref` 更新スキップ |
| **離脱ダイアログ** | `cancel` / `close` は破棄扱いにしない |

## Filesystem side effects（review-ready）

| 操作 | ディスク |
|------|----------|
| `set_power_plan` | **OS のアクティブ電源プランのみ**。Tweaker データディレクトリのファイルは変更しない |
| Automation item 保存 | SQLite のみ |
| Lua script v1 | ファイル IO なし |

**Sequence 部分成功**: 成功済み action（例: 電源プラン変更）は**自動ロールバックしない**。失敗は run log に記録し、後続 action は実行しない（`continue_on_error` なし時）。UI で「途中まで適用されうる」旨を短く示す。

`powercfg` 失敗時は OS プランは変更前のまま。

## Long-running worker loop（review-ready）

| トピック | v1 契約 |
|----------|---------|
| **Schedule tick** | 1 分に **1 回**評価。同分の二重 `schedule.tick` なし |
| **トレイ常駐** | 最小化中もスケジューラ・event ワーカーは継続（**Automation runtime**） |
| **同一 item の連続失敗** | **診断ログ**はレート制限（目安: 同一 item あたり **10 分に 1 回**）。**run log** は毎回記録 |
| **キュー溢れ** | drop + 診断ログ。ingest は継続 |
| **失敗 item の自動無効化** | v1 なし |

## PR scope（review-ready）

**2 PR 分割**（#225）:

| PR | 内容 |
|----|------|
| **PR1 Backend** | event channel + 単一ワーカー、スケジューラ、action catalog（`change_status` / `set_power_plan`）、DB マイグレーション、Wails API（item CRUD、run log、runtime status、detected power plans）、Lua sandbox + script 実行、Go テスト（Test plan 表） |
| **PR2 Frontend** | Automation 一覧、rule builder、script エディタ、run log パネル、`automation:run-log-changed`、未保存ガード、i18n×5、E2E・Storybook |

PR1 マージ時点で **rule のみ** end-to-end（既存 friend_joined 回帰含む）。script タブ UI は PR2。Draft PR 早期提出を推奨。

## Consequences

- event 配信層（in-process bus または同等）を再設計する。subscriber は AutomationUseCase（rule 評価）と script runtime
- 既存 `EvalRule` の payload キー一致は **条件エンジン v1** の一部として拡張（状態条件・プロセス条件）
- `AutomationView.vue` はセクション型ビルダー＋script エディタ＋run log へ置き換え規模
- E2E: スケジュール・電源プランは環境依存のため、単体／結合テストとモック中心。E2E は rule CRUD と表示を優先
- 公開成果物（PR/Issue）では実ユーザー ID・表示名を載せない

## Test plan（review-ready）

v1 の最小テスト契約（grill-review-ready 合意）。PR の Acceptance / チェックリストに載せる。

| 入力・状況 | 期待 | テスト名（案） |
|---|---|---|
| 無効 item | 発火しない | `TestAutomation_eval_disabledItemSkipped` |
| `schedule.tick` 曜日不一致 | 発火しない | `TestAutomation_schedule_wrongWeekday` |
| 条件 `vrchat_running` だが停止中 | action なし | `TestAutomation_condition_vrchatNotRunning` |
| sequence 2 件目失敗、`continue_on_error` なし | 1/2、run log 失敗 | `TestAutomation_sequence_stopsOnError` |
| 同上、`continue_on_error` あり | 2/2 試行 | `TestAutomation_sequence_continuesOnError` |
| Log replay 経路 | automation 0 回 | `TestAutomation_logReplay_noFire` |
| 非 Windows + `set_power_plan` | action エラー | `TestAutomation_setPowerPlan_unsupportedPlatform` |
| preset 未解決 | action 失敗 | `TestAutomation_setPowerPlan_presetUnresolved` |
| Lua 10s 超過 | run log 失敗、プロセス生存 | `TestAutomation_script_luaTimeout` |
| event キュー溢れ | drop + 診断ログ、ingest 継続 | `TestAutomation_worker_queueOverflow` |
| `ToggleAutomationItem` 未知 ID | `error` | `TestAutomation_toggle_unknownId` |
| スケジューラ init 失敗 | `GetAutomationRuntimeStatus.available=false` | `TestAutomation_runtime_subsystemUnavailable` |
| `friend_joined` + run log | 表示名あり、DTO に `usr_*` なし | `TestAutomation_runLog_displayNameNoUserId` |
| 同一分 item A,B | ID 昇順で評価 | `TestAutomation_schedule_stableItemOrder` |

**手動（Windows）**: `set_power_plan` 実プラン切替 1 回（Linux CI はモック）。

### 従来の層別スモーク

| 層 | 状況 | 期待 |
|----|------|------|
| Front | rule builder 保存・再読込 | JSON 往復 |
| Front | Windows 以外 | `set_power_plan` 選択肢なし |
| E2E | rule 作成・有効化 | 一覧・サマリー表示 |

## References

- [`CONTEXT.md`](../../CONTEXT.md) — Automation セクション
- `internal/domain/automation/` — 既存 rule engine
- `internal/infrastructure/logwatcher/automation_trigger_handler.go` — friend_joined
- ADR 0005 — Log replay と automation 非発火
