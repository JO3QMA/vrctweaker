# ADR 0001: Gallery scope と一覧時の欠損除外

## Status

Accepted（[#99](https://github.com/JO3QMA/vrctweaker/issues/99) で実装）

## Context

- スクリーンショットは DB に広くインデックスされるが、Gallery は「今の VRChat Picture folder の写真を思い出す」体験に限定したい
- Picture folder を変更したあとも、過去パスの行を DB に残し、フォルダを戻したときに再表示できるようにしたい
- ユーザーが手動で画像を削除した場合、壊れたサムネより「一覧に出さない」方が望ましい

用語は [`CONTEXT.md`](../../CONTEXT.md) を正本とする。

## Decision

1. **Gallery scope**: 一覧・検索 API は常に現行 Picture folder 配下（`FilePathPrefix`）に限定する
2. **Out-of-scope screenshot**: フォルダ外の DB 行は残すが Gallery には返さない
3. **Missing screenshot file**: 一覧取得のたび `os.Stat` で regular file を確認し、欠損行は返さない（DB 削除はしない）
4. **Taken-at**: メタデータの撮影日時を優先し、無ければファイル更新日時で Date grouping する

## Consequences

### 正

- Gallery とディスクの体感が一致しやすい
- フォルダ変更・復帰シナリオで再表示可能
- `ListScreenshotsInGalleryScope` に scope + stat を集約できる

### 負

- 一覧のたびに stat が走る（件数が極端に多い場合は将来キャッシュ検討）
- DB に残る out-of-scope / missing 行の整理は別操作（Manual sync では削除しない）

## Implementation

- `internal/domain/media/gallery_scope.go` — `PictureFolderPathPrefix`
- `internal/usecase/media_gallery_list.go` — `ListScreenshotsInGalleryScope`
- `app.go` — `listGalleryScreenshotDTOs` 経由で `Screenshots` / `SearchScreenshots`
