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

### 一覧の存在確認キャッシュ（[#242](https://github.com/JO3QMA/vrctweaker/issues/242)）

欠損除外の「一覧 API が存在確認を担う」契約は維持したまま、パス単位の存在確認結果を
**短い TTL（30 秒）でキャッシュ**する（`galleryFileExistsCache`、プロセス内）。

- **イベント無効化**: Manual sync（`SyncPictureFolder`）と Automatic ingest
  （`IngestUnderPictureRootSince` / ウォッチャー経由の `IngestScreenshotFile`）が成功したら
  キャッシュを無効化する。新規・復帰ファイルは次回一覧から即反映される。
- **TTL**: ディスク上の外部削除はウォッチャーが検知しないため、TTL 経過後に再 stat して
  一覧へ反映する（最大 30 秒の stale を許容）。
- **並行性**: 無効化は世代番号を進め、`get` で観測した世代と一致しない `put` は破棄する。
  一覧の stat と ingest の無効化が並行しても、古い「missing」結果が無効化後に再登録されない。
  バルク処理（`ScanDirectory` / `IngestUnderPictureRootSince` / `SyncPictureFolder`）は最後に
  一括で無効化するため、ループ内のパス単位無効化（世代更新）は行わない。長時間の同期中に
  並行する一覧のキャッシュ登録が連続的に失われるのを防ぐ。
- **上限**: キャッシュは `galleryFileStatCacheMaxItems`（8192）で境界。一覧に現れなくなった
  パスが残り続けるのを防ぐ。ただし put を安価に保つため、全走査の掃除は倍容量（2×）到達時に
  のみ行い、それ以外は境界を超えた分だけ任意に追い出す。
- **エラー分類**: `os.IsNotExist` のみ「欠損」としてキャッシュする。権限エラーや一時 I/O エラーは
  キャッシュせず次回一覧で再試行する（一時障害で実在ファイルが TTL 分ギャラリーから隠れない）。
- 実装: `internal/usecase/gallery_file_exists_cache.go`、`internal/usecase/media_gallery_list.go`

## Implementation

- `internal/domain/media/gallery_scope.go` — `PictureFolderPathPrefix`
- `internal/usecase/media_gallery_list.go` — `ListScreenshotsInGalleryScope`
- `app.go` — `listGalleryScreenshotDTOs` 経由で `Screenshots` / `SearchScreenshots`
