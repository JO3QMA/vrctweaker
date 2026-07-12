# ADR 0006: Instance rejoin と Last launch profile

## Status

Accepted（grill-with-docs セッションで合意、[Issue #31](https://github.com/JO3QMA/vrctweaker/issues/31)）

## Context

- VRChat クラッシュ後などに、直前の部屋へ素早く戻りたい（Issue #31）
- 既存 **World join**（Gallery）は `world_id` のみを `vrchat://launch?id=<worldID>` に渡し、**同じインスタンスとは限らない**
- **Quick launch** は Default launch profile 固定。Dashboard で profile を選べる導線はない
- Output log ingest により `play_sessions.instance_id` に VRChat instance key（例: `wrld_…:12345~public~region(jp)`）が入る
- Default launch profile が未設定の状態があり得る（Quick launch は disabled になる）

用語は [`CONTEXT.md`](../../CONTEXT.md) の **Launcher** セクション（**Rejoin target**、**Last launch profile**、**Instance rejoin launch profile**、**Instance rejoin section**、**Instance rejoin**）を正本とする。

## Decision

1. **Rejoin target**: `instance_id` が空でない Play session のうち `start_time` が最も新しい 1 件。Open / 終了済みは問わない。複数 Log source の Open play session が同時にあっても開始時刻最新 1 件
2. **Instance rejoin**: Dashboard から Rejoin target の VRChat instance key **丸ごと**を `vrchat://launch?id=<key>` に載せて起動する。**World join** とは別導線。満員・非公開などの成否は VRChat 側に委ねる
3. **Last launch profile**: `app_settings` に Launch profile ID を永続化。Profile launch または Instance rejoin で **起動プロセス開始に成功したときのみ**更新。Dashboard セレクタ変更だけでは更新しない
4. **Instance rejoin launch profile**: Dashboard 上で選択。初期値は Last launch profile → Default launch profile → 保存済み一覧の先頭。起動引数は profile 保存内容（`-no-vr` 含む）をそのまま使い、起動時 Desktop/VR オーバーライドは行わない。Launcher の Unsaved launch profile edits は反映しない
5. **Instance rejoin section**: Quick launch 直下。Rejoin target があり、かつ Launch profile が 1 件以上あるときだけ表示。表示名あり →「{ワールド名} に参加」、なし → 汎用ラベル。`wrld_*` 等の技術 ID は出さない
6. **Activity retention**: 対象 Play session 削除で Rejoin target が無くなった場合、説明なくセクション非表示
7. **Quick launch**: 今回は変更しない（Launch profile セレクタは別 Issue）

## Failure modes（review-ready）

Dashboard / Wails 境界でのユーザー可視挙動:

| 状況 | ユーザー | サーバ |
|------|----------|--------|
| Rejoin target なし | Instance rejoin section 非表示 | — |
| Launch profile 0 件 | 同上 | — |
| Rejoin target 取得の **DB / infra 失敗** | 同上（エラーメッセージなし） | ログに記録 |
| Instance rejoin 起動失敗（Steam/VRChat 未検出、profile 削除済み等） | `ElMessage.error` | 必要ならログ |
| 起動プロセス開始成功 → VRChat 側 Join 失敗 | VRChat 側 UI | Last launch profile は更新済み |

## Return-value contract（review-ready）

### 読み取り — `GetInstanceRejoinSection() (*InstanceRejoinSectionDTO, error)`

| 層 | target / profile 無し（業務上 unavailable） | infra 失敗 |
|----|---------------------------------------------|------------|
| **Usecase** | `nil, nil` | `nil, error` |
| **App (Wails)** | `nil, nil` | ログ後 **`nil, nil`**（frontend には error を出さない） |
| **Frontend** | DTO `null` → section 非表示 | 同上 |

DTO には Rejoin target（表示名含む）、Launch profile 一覧、初期選択 profile ID（Last → Default → 先頭）を含める。

### 起動 — `InstanceRejoin(profileID string) error`

- `profileID` 空 → **`error`**
- profile 未存在 → **`error`**
- `cmd.Start` 失敗 → **`error`**（frontend → `ElMessage.error`）
- 成功 → Last launch profile 更新
- 空 ID のサーバ側自動解決（Last → Default → 先頭）は **行わない**（セレクタ state と二重管理を避ける）

## Config（review-ready）

- **新 tunable なし** — env 追加・数値 cap 追加は行わない
- Rejoin target の保存期間は既存 **Activity retention**（`log_retention_days`）のみ
- **`last_launch_profile_id`**: `settings_usecase.go` にキー定数。Get/Set で profile ID 文字列をそのまま永続化（保存時の存在チェックなし）。読み取り時 `GetInstanceRejoinSection` で Last → Default → 一覧先頭へフォールバック

## Trust boundaries（review-ready）

| 境界 | 方針 |
|------|------|
| **Frontend** | `GetInstanceRejoinSection` DTO に **VRChat instance key を含めない**（表示名等のメタデータのみ）。Join は `InstanceRejoin(profileID)` がサーバ側で Rejoin target を再解決 |
| **起動 URL** | usecase 内で `play_sessions` から key を取得し `vrchat://launch?id=…` を組み立て |
| **ログ** | フル instance key（`usr_*` 埋め込みあり得る）を **ログに出さない**。`world_id` または play session ID 程度 |
| **公開成果物** | `docs/agents/redaction.md` 遵守。テストは合成 key のみ |

## Lifecycle（review-ready）

- **起動** — 既存 `exec.CommandContext` + `cmd.Start`（同期）。新 goroutine は増やさない
- **Section 更新** — `onMounted` に加え `activity:encounters-changed` を購読（ActivityView と同様 debounce）。ingest 後に表示/非表示・ラベルが追従。Dashboard 開き直しだけに依存しない

## Interim / v1 scope（review-ready）

### v1 でやる

- 既存 `LaunchToWorld` / `BuildJoinWorldArgs` を **そのまま再利用** — instance key を `worldID` 引数に渡す（launch location として有効）
- `ponytail:` コメントで param が world ID またはフル instance key である旨を明記。関数リネームは v1 外

### v1 スコープ外

- Quick launch の Launch profile セレクタ（別 Issue）
- `BuildJoinWorldArgs` / `LaunchToWorld` のリネーム
- VRChat Join 成否のポーリング・再試行（成否は VRChat 側）
- Dashboard 以外（Activity 等）からの Instance rejoin 導線

## Edge cases → tests（review-ready）

v1 最低ライン。合成 VRChat instance key のみ（`docs/agents/redaction.md`）。

### Backend（usecase / launcher）

| 入力 | 期待 | テスト名 |
|------|------|----------|
| Play session 0 件 | Rejoin target `nil` | `TestGetRejoinTarget_noSessions` |
| 最新 session の `instance_id` 空、古い行に key | 空でない最新を返す | `TestGetRejoinTarget_skipsEmptyInstanceID` |
| 複数 open session（複数 Log source） | `start_time` 最新 1 件 | `TestGetRejoinTarget_picksLatestStartTime` |
| `world_info` 表示名あり | DTO に表示名 | `TestGetInstanceRejoinSection_withWorldDisplayName` |
| 表示名なし | 表示名空 | `TestGetInstanceRejoinSection_withoutWorldDisplayName` |
| Last launch profile 有効 | 初期選択 = Last | `TestResolveInstanceRejoinProfile_lastLaunch` |
| Last 削除済み | Default → 先頭 | `TestResolveInstanceRejoinProfile_staleLastLaunch` |
| Last / Default 無し | 一覧先頭 | `TestResolveInstanceRejoinProfile_firstProfile` |
| Launch profile 0 件 | section state `nil` | `TestGetInstanceRejoinSection_noProfiles` |
| DB エラー | usecase `error` | `TestGetRejoinTarget_dbError` |

### Backend（起動）

| 入力 | 期待 | テスト名 |
|------|------|----------|
| フル instance key → `BuildJoinWorldArgs` | URL に key 丸ごと | `TestBuildJoinWorldArgs_fullInstanceKey` |
| `InstanceRejoin("")` | `error` | `TestInstanceRejoin_emptyProfileID` |
| profile 未存在 | `error` | `TestInstanceRejoin_profileNotFound` |
| `cmd.Start` 失敗 | `error`、Last 未更新 | `TestInstanceRejoin_startFailureNoLastUpdate` |
| `cmd.Start` 成功 | Last 更新 | `TestInstanceRejoin_successUpdatesLastLaunch` |

### App（Wails）

| 入力 | 期待 | テスト名 |
|------|------|----------|
| usecase infra `error` | `nil, nil` + ログ | `TestGetInstanceRejoinSection_degradesOnError` |

### Frontend

| 入力 | 期待 | テスト名 |
|------|------|----------|
| section DTO `null` | section 非表示 | `DashboardView hides instance rejoin section when unavailable` |
| 表示名あり | 「{name} に参加」 | `DashboardView shows world name on rejoin button` |
| 表示名なし | 汎用ラベル | `DashboardView shows generic rejoin label without world name` |
| 起動 error | `ElMessage.error` | `DashboardView shows error on instance rejoin failure` |
| `activity:encounters-changed` | debounce 再取得 | `DashboardView refreshes instance rejoin on activity event` |

E2E は v1 外（Wails mock + usecase 単体で足りる）。

## Considered Options

| 論点 | 採用 | 却下した案 |
|------|------|------------|
| Join URL | VRChat instance key 丸ごと | `world_id` のみ（World join と同じ・別 instance になりうる） |
| 起動 profile | Dashboard で選択、Last 初期値 | Default 固定のみ |
| Desktop/VR | 選んだ profile の `-no-vr` に従う | Dashboard トグルで起動時オーバーライド |
| Last launch profile 保存 | 起動成功時のみ | セレクタ変更時も保存 |
| profile 初期値 | Last → Default → 一覧先頭 | Default 必須（Default 未設定時に Rejoin 不可） |
| target / profile 無し UI | 条件を満たさなければセクション非表示 | 常時表示 + disabled + 説明文 |
| retention 後 | 黙って非表示 | 空状態メッセージや Settings 導線 |
| Quick launch | 今回スコープ外 | 同時に profile セレクタを追加 |

## Consequences

### 正

- クラッシュ復帰で「同じ部屋」を狙える（World join より意図が明確）
- Desktop/VR は Launcher で profile を分けておけば Dashboard でも選べる
- Quick launch の「Default でサッと起動」という役割を維持できる

### 負

- `BuildJoinWorldArgs` 等の命名・呼び出しは instance key も受け付けるが、World join との混同に注意
- Last launch profile 参照先が削除されたら Default → 一覧先頭へフォールバックが必要
- Activity retention が短いと Rejoin 導線が消える（Activity 画面の retention ヒントに委ねる）

## Implementation（想定）

| 層 | 内容 |
|----|------|
| Settings / usecase | `last_launch_profile_id` を `app_settings` に get/set |
| Activity usecase | Rejoin target 解決（最新 Play session + `world_info` 表示名） |
| Launcher usecase | `LaunchToWorld`（または同等）に instance key を渡す。Profile launch / Instance rejoin 成功時に Last launch profile 更新 |
| Wails | `GetRejoinTarget`（または DTO 付き）、`InstanceRejoin(profileID)` |
| Frontend | `DashboardView` — Instance rejoin section（profile セレクタ + ボタン） |
| Follow-up | Quick launch への profile セレクタ（別 Issue） |

## Related

- Issue [#31](https://github.com/JO3QMA/vrctweaker/issues/31)
- [`docs/features/media-world-join-from-screenshot.md`](../features/media-world-join-from-screenshot.md) — World join（`world_id` のみ）
- ADR 0005 — Play session の `instance_id` / Log source
