# ADR 0010: Dashboard launch block

## Status

Accepted（grill-with-docs で合意、[Issue #185](https://github.com/JO3QMA/vrctweaker/issues/185)）

## Context

- Dashboard の **Quick launch**（Default launch profile 固定の単体ボタン）と **Instance rejoin section**（専用セレクタ + Rejoin ボタン）が分離しており、起動前に profile と参加先を一度に把握しづらい
- [ADR 0006](0006-instance-rejoin-and-last-launch-profile.md) では Quick launch の profile セレクタはスコープ外とし、follow-up Issue に委ねていた
- `GetInstanceRejoinSection()` は Rejoin target が無いと `null` を返す。統合後は Rejoin 無しでもセレクタ + Quick launch を示す必要がある
- Server status（ADR 0009）と同様、常時表示ブロックの infra 失敗は `ElMessage` で汚さない方がよい

用語は [`CONTEXT.md`](../../CONTEXT.md) の **Launcher** / **Dashboard** セクション（**Dashboard launch block**、**Dashboard launch profile**、**Quick launch**、**Instance rejoin**、**Last launch profile**）を正本とする。

## Decision

1. **UI**: **Dashboard launch block** — Server status section 直下。共有 Launch profile セレクタ + Quick launch（汎用ラベル）+ Instance rejoin（Rejoin target があるときのみ）
2. **Quick launch**: Default 固定をやめ、**Dashboard launch profile**（共有セレクタの選択）で部屋指定なし起動。成功時 **Last launch profile** を更新する（Profile launch / Instance rejoin と同様）
3. **表示**: ブロックは **常に表示**。profile ≥ 1 → セレクタ・Quick launch 有効。profile 0 → disabled + Launcher 作成案内 + **Launcher へのリンク（またはボタン）**。Rejoin target 無し → Instance rejoin ボタンのみ非表示
4. **Wails**: 新規 **`GetDashboardLaunchBlock()`**。同一 PR で **`GetInstanceRejoinSection()` を削除**
5. **起動**: `LaunchVRChat(profileID)` 成功時に `setLastLaunchProfileOnSuccess` を呼ぶ（Quick launch 用）。Instance rejoin は既存 `InstanceRejoin` を維持
6. **失敗**: `GetDashboardLaunchBlock` の infra 失敗はブロック内インラインエラー。**`ElMessage` は出さない**（ADR 0009 と同方針）。`activity:encounters-changed` の debounce 再取得で復帰しうる
7. **起動ボタン失敗**（Quick launch / Instance rejoin のクリック後）: **`ElMessage.error`**。読み込み失敗（バックグラウンド）とユーザー操作失敗は別契約。Instance rejoin 失敗時は既存どおりセクション再取得。Last launch profile の保存失敗（起動成功後）は best-effort（ログのみ、呼び出し元には成功）
8. **i18n**: キーは **`dashboard.launchBlock.*`** に統一（例: `quickLaunch`, `profilePlaceholder`, `rejoinWithWorld`, `rejoinGeneric`, `loadError`, `launchError`, `rejoinError`, `emptyState`, `goToLauncher`）。旧 `dashboard.instanceRejoin*` と `dashboard.launchWithProfile` は削除。全 5 ロケールを同期
9. **PR  scope**: **1 PR**（Go + Vue + i18n×5 + テスト一括）。旧 API 削除まで同一 PR で完結

## Return-value contract

### 読み取り — `GetDashboardLaunchBlock() (*DashboardLaunchBlockDTO, error)`

| 状況 | App / Usecase | Frontend |
|------|---------------|----------|
| 成功 | DTO（profiles, selectedProfileId, rejoin または nil） | 通常表示 |
| 読み取り infra 失敗（DB 等） | **`error`** を返す | ブロック内 i18n エラー。toast なし |
| Quick launch / Instance rejoin 起動失敗 | `LaunchVRChat` / `InstanceRejoin` が **`error`** | **`ElMessage.error`**。Rejoin 時はブロック再取得 |
| Last 保存失敗（起動成功後） | `setLastLaunchProfileOnSuccess` がログのみ | ユーザーには成功扱い（起動済み） |
| profile 0 件 | DTO（profiles 空, rejoin は状況に応じて） | 空状態 + Launcher 導線 |

```go
type DashboardLaunchBlockDTO struct {
    Profiles          []LaunchProfileDTO      `json:"profiles"`
    SelectedProfileID string                  `json:"selectedProfileId"`
    Rejoin            *DashboardRejoinDTO     `json:"rejoin"` // nil = Instance rejoin ボタン非表示
}

type DashboardRejoinDTO struct {
    PlaySessionID    string `json:"playSessionId"`
    WorldDisplayName string `json:"worldDisplayName"`
}
```

- **SelectedProfileID** の解決は既存 `ResolveInstanceRejoinProfileID`（Last → Default → 先頭）を流用
- DTO に **VRChat instance key を含めない**（0006 と同様）。Join は `InstanceRejoin(profileID, playSessionID)`

## Considered options

| 案 | 却下理由 |
|----|----------|
| Quick launch は Default 固定のまま見た目だけ並べる | 「両方とも profile を選びたい」に合わない |
| セレクタをボタンごとに 2 つ | 重複操作 |
| `GetInstanceRejoinSection` の意味を拡張 | 名前と `null` 契約が新仕様と矛盾 |
| 読み取り失敗を `(nil, error)` + `ElMessage`（0006 同様） | 常時表示ブロックでは toast 汚染。0009 と不整合 |
| profile 0 件でブロック非表示 | 起動できない理由と Launcher 導線が消える |

## Consequences

- ADR 0006 の「Quick launch は変更しない」は本 ADR で supersede（0006 は Instance rejoin 単独導入の歴史として残す）
- `LaunchVRChat` は Dashboard Quick launch からのみ呼ばれるため、Last launch profile 更新の影響範囲は Dashboard 起動に限定される
- Frontend: `DashboardView` の起動 UI を `DashboardLaunchBlock.vue`（または同等）に抽出推奨。E2E の `data-testid` は `dashboard-launch-block` 等に追随
- **Async UI:** `ServerStatusSection` と同型の **`generation` カウンタ** + `inFlight`。`onUnmounted` で increment し、load / debounce refresh の `await` 後に古い世代なら `ref` 更新をスキップ。起動失敗の `ElMessage` はアンマウント後も表示してよい

## Test plan（review-ready）

| 層 | 入力・状況 | 期待 | テスト名（案） |
|----|------------|------|----------------|
| Go | profile ≥ 1、Rejoin あり | DTO に profiles + rejoin + selectedProfileId | `TestGetDashboardLaunchBlock_withRejoin` |
| Go | profile ≥ 1、Rejoin なし | DTO、`rejoin: nil` | `TestGetDashboardLaunchBlock_withoutRejoin` |
| Go | profile 0 件 | DTO、profiles 空 | `TestGetDashboardLaunchBlock_noProfiles` |
| Go | Last 有効 | selected = Last | `TestGetDashboardLaunchBlock_selectedLast` |
| Go | Last 削除済み | Default → 先頭 | 既存 `ResolveInstanceRejoinProfile_*` |
| Go | ListProfiles infra 失敗 | `error` | `TestGetDashboardLaunchBlock_listProfilesError` |
| Go | Quick launch 成功 | Last 更新 | `TestLaunchVRChat_successUpdatesLastLaunch` |
| Go | Quick launch 失敗 | Last 未更新 | `TestLaunchVRChat_failureNoLastUpdate` |
| Go | Rejoin stale playSessionID | `error` | 既存 `TestInstanceRejoin_stalePlaySessionID` |
| Vitest | load 成功 + rejoin | セレクタ + 両ボタン | `shows launch block with rejoin button` |
| Vitest | load 成功、rejoin nil | Quick launch のみ | `shows launch block without rejoin button` |
| Vitest | profile 0 | disabled + 空状態 + Launcher リンク | `shows empty state with launcher link` |
| Vitest | load error | インラインエラー、toast なし | `shows inline error on load failure` |
| Vitest | Quick launch 失敗 | `ElMessage.error` | `shows error on quick launch failure` |
| Vitest | Rejoin 失敗 | `ElMessage.error` + reload | `shows error on rejoin failure` |
| Vitest | activity event | debounce 再 load | 既存テストを新 API に追随 |
| Vitest | unmount 後 load 完了 | ref 更新なし | `skips state update after unmount` |
| E2E | mock 正常 | `dashboard-launch-block` 表示 | mock `GetDashboardLaunchBlock` 追随 |

## References

- [Issue #185](https://github.com/JO3QMA/vrctweaker/issues/185)
- [ADR 0006](0006-instance-rejoin-and-last-launch-profile.md)
- [ADR 0009](0009-dashboard-server-status.md)
