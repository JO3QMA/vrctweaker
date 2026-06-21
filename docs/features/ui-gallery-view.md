# Gallery（スクリーンショット一覧 UI）

## 概要

`CONTEXT.md` の **Gallery** ドメイン用語に沿ったスクリーンショット閲覧画面。実装は `frontend/src/views/GalleryView.vue`。

用語の正本: リポジトリルート [`CONTEXT.md`](../../CONTEXT.md)

## ゴール

- **Date grouping**（年 → 月 → 日）で Screenshot を一覧表示する
- **World search** と **Date range filter** で絞り込みつつ、グルーピングを維持する
- 詳細プレビューで taken-at / ワールド名 / `world_id` / ファイルパスを表示する
- **World join**（`world_id` がある場合のみ）— 別機能 [`media-world-join-from-screenshot.md`](./media-world-join-from-screenshot.md)
- **Manual sync**（Picture folder sync）でフォルダとインデックスを揃える

## 仕様

### 一覧の主軸

- **Date grouping** を既定とする（`galleryDateGroups.ts` の仮想行）
- Taken-at はメタデータ優先、無ければファイル更新日時（`CONTEXT.md` **Taken-at**）

### Gallery scope と欠損

- 一覧 API は **Gallery scope**（現行 Picture folder 配下）のみ返す（[#99](https://github.com/JO3QMA/vrctweaker/issues/99)）
- **Missing screenshot file** は一覧取得のたび `stat` で除外する（DB 行は削除しない）
- Picture folder 外の **Out-of-scope screenshot** も Gallery には出さない

### World search

検索ボックス 1 つ（`CONTEXT.md` **World search**）:

| 入力 | バックエンド |
|------|----------------|
| `wrld_` で始まる | `worldId` 完全一致 |
| それ以外 | `worldName` 部分一致 |

- UI の自動判定: [#100](https://github.com/JO3QMA/vrctweaker/pull/104)（マージ前は `worldId` のみ送信）
- debounce + Enter で `SearchScreenshots` を呼ぶ

### Date range filter

- `dateFrom` / `dateTo`（Taken-at 基準）で `SearchScreenshots` に渡す
- 有効時も Date grouping を維持
- UI: [#100](https://github.com/JO3QMA/vrctweaker/pull/104)

### グリッド・詳細

- サムネイルは `ScreenshotThumbnailDataURL`（WebView 向け data URL）
- 詳細パネル: takenAt / worldName / worldId / filePath、外部アプリで開く・フォルダ表示
- `world_id` が空なら World join ボタン無効

### 同期

| 操作 | 用語 | 実装 |
|------|------|------|
| ウォッチャー | **Automatic ingest** | 新規ファイルのみ取込、`gallery:screenshots-changed` |
| 「フォルダを同期」ボタン | **Manual sync** → **Picture folder sync** | `ScanScreenshotDir` |

**Picture folder sync** の内容（[#101](https://github.com/JO3QMA/vrctweaker/pull/105)）:

1. 新規画像の取込
2. 選択的メタ再抽出（新規、`world_id` 空、ファイルサイズ変更、サムネイル stale）
3. 欠損の Gallery 非表示は一覧 API 側（sync で DB 削除はしない）

進捗イベント: `gallery:scan-progress` / `gallery:scan-done`

## Wails API（実装済み）

| メソッド | 用途 |
|----------|------|
| `Screenshots(worldId?)` | フィルタなし or worldId のみ（後方互換） |
| `SearchScreenshots(filter)` | worldId / worldName / dateFrom / dateTo + Gallery scope |
| `GetScreenshot(id)` | 詳細 |
| `ScreenshotThumbnailDataURL(id)` | グリッド用サムネ |
| `ScanScreenshotDir(path)` | Manual sync（[#101](https://github.com/JO3QMA/vrctweaker/pull/105) で `SyncPictureFolder` に委譲予定） |
| `IsGalleryScanning()` | 同期中フラグ |
| `OpenScreenshotExternally` / `RevealScreenshotInFileManager` | OS 連携 |

`ReindexScreenshotDir` は開発・メンテ用（Gallery UI からは呼ばない）。

## 実装状況

| 項目 | 状態 |
|------|------|
| Date grouping UI | 実装済み |
| Gallery scope + 欠損除外 | 実装済み ([#99](https://github.com/JO3QMA/vrctweaker/issues/99)) |
| World search 自動判定 | PR [#104](https://github.com/JO3QMA/vrctweaker/pull/104) |
| Date range filter UI | PR [#104](https://github.com/JO3QMA/vrctweaker/pull/104) |
| Picture folder sync | PR [#105](https://github.com/JO3QMA/vrctweaker/pull/105) |
| World join | 実装済み |

## 受け入れ条件（DoD）

- Date grouping で一覧が表示され、クリックで詳細が開く
- Gallery scope 外・欠損ファイルが一覧に含まれない
- `world_id` が空のとき World join が無効
- World search / Date range / Manual sync は上表の Issue・PR と一致

## テスト

- Vitest: `galleryDateGroups.spec.ts`, `GalleryView.spec.ts`（[#100](https://github.com/JO3QMA/vrctweaker/pull/104) で `gallerySearchFilter` 追加）
- E2E: `frontend/e2e/gallery.spec.ts`
