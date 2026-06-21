# World join（スクリーンショットからワールドへ入る）

## 概要

`CONTEXT.md` の **World join**: Gallery 詳細から、Screenshot の `world_id` を使って VRChat を起動し対象ワールドへ入る。

US 1.2「このワールドへ Join」に対応。**実装済み**。

## ゴール

- `world_id`（例: `wrld_...`）がある Screenshot でワンクリック起動
- `world_id` が無い場合はボタン無効（または API エラー）
- **常にデフォルト Launch profile** の引数に join 引数を連結（`CONTEXT.md`）

## 仕様

### 入力

- UI: `JoinWorldFromScreenshot(screenshotId)`
- 内部: `world_id` を DB から解決 → `JoinWorld(worldId)`

### 起動

- `LauncherUseCase.LaunchToWorld(ctx, profileID="", worldID, ...)`
  - `profileID` 空 → デフォルトプロファイル
- join 引数: `BuildJoinWorldArgs(baseArgs, worldID)` → 末尾に `vrchat://launch?id=<worldID>`

### プラットフォーム

- Windows / Linux: Steam 経由起動（`steam -applaunch 438100` + 引数）
- 詳細は `internal/usecase/launcher_usecase.go` と `docs/ai_dlc/VRChat 起動引数調査ガイド.md`（ローカル参照）

## コード配置

| 層 | ファイル |
|----|----------|
| Wails | `app.go` — `JoinWorldFromScreenshot`, `JoinWorld` |
| Use case | `internal/usecase/launcher_usecase.go` — `LaunchToWorld`, `BuildJoinWorldArgs` |
| UI | `frontend/src/views/GalleryView.vue` — 詳細パネルの Join ボタン |

## Wails API

```text
JoinWorldFromScreenshot(screenshotId: string) -> error
JoinWorld(worldId: string) -> error   // 直接 worldId 指定（Launcher 連携）
```

エラー例:

- screenshot 未存在
- `screenshot has no world_id`

## 受け入れ条件（DoD）

- [x] `world_id` ありで Join 押下 → VRChat 起動プロセス開始
- [x] `world_id` なしでボタン無効
- [x] デフォルト Launch profile の引数が使われる

## テスト

- `internal/usecase/launcher_usecase_test.go` — `TestBuildJoinWorldArgs`, `TestLauncherUseCase_LaunchToWorld_*`
- `frontend/src/views/__tests__/GalleryView.spec.ts` — Join ボタン有効/無効

## 関連

- Gallery UI: [`ui-gallery-view.md`](./ui-gallery-view.md)
- 用語: [`CONTEXT.md`](../../CONTEXT.md) — **World join**
