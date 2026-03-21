# レビューコメント（Untitled-1）検証・訂正メモ

元ノートのうち、現行コードと食い違っていた点の訂正と、検証済みサマリです。

## 訂正（元レビューの誤り）

1. **`App.scanScreenshotDir` 失敗時のバナー**  
   元: `loadError` に入ると記載。  
   **訂正**: [`frontend/src/views/GalleryView.vue`](../frontend/src/views/GalleryView.vue) の `scanFolder` では `scanError`（`role="status"`）に設定している。

2. **Windows `explorer /select` の引用符**  
   元: `/select,"path"` のように二重引用符を足す実装と記載。  
   **訂正**: [`internal/infrastructure/desktop/shell_open.go`](../internal/infrastructure/desktop/shell_open.go) の `revealWindows` は `"/select," + abs` とし、**パス用の追加引用符は付けない**（コメントで意図を明記）。

3. **`ErrScreenshotThumbnailNotCached`**  
   元: 名前付きエラーとして言及。  
   **訂正**: リポジトリにそのシンボルはない。キャッシュ不整合時は [`internal/usecase/media_thumbnail.go`](../internal/usecase/media_thumbnail.go) の `ScreenshotThumbnailDataURL` が `fmt.Errorf("thumbnail still unavailable after ensure")` 等を返す。

4. **行番号**  
   元: `DefaultVRChatPictureFolder` が `app.go:522` 付近など。  
   **訂正**: ブランチにより変動するためファイル内検索を推奨。`DefaultVRChatPictureFolder` は `app.go` 内の該当コメント付近。`GalleryView.spec.ts` の「getVRChatConfig rejects」は **313 行付近**（元ノートの 187 行は別シナリオ）。

## 検証サマリ（妥当だった指摘）

- `cmd /c start` と cmd のメタ文字（実ファイル検証後の限定的リスク）
- `scanProgressEmitter` のスロットルと `flush`
- `bindings.go` の `FileSizeBytes` ポインタ共有（`TakenAt` はコピー）
- SQLite DSN の `foreign_keys` とコネクションプール
- ギャラリーのローカル日付グルーピング
- 更新ボタンが debounce タイマをクリアしない → **実装で修正済みの場合は本メモよりコードを正とする**
- `galleryHeaderAt` のテンプレ内複数回評価、`syncThumbnailsForVisible` の多重起動、`cursor++` の単一スレッド安全性
- `watch(list, { deep: true })` のオーバーヘッド（差し替えのみなら浅い watch で足りる）→ **実装で見直し済みの場合はコードを正とする**
- `media_thumbnail` の base64 オーバーヘッド、検証の重複、TOCTOU 懸念 → **リファクタで整理済みの場合はコードを正とする**
- `ScanDirectory` のシグネチャ変更、Reindex のサムネ無し行の扱い
- picturewatcher の `flush` とコンテキスト取消
- `callApp` の fallback が「バインディング欠如時のみ」であること
- `App.vue` の scoped と `:not(.gallery-view)`
- `ensureScreenshotsFileSizeColumn` のエラーメッセージ文字列依存
- JPEG SOI ガード（フィールド名は `JpegBlob`）

## 実装メモ（本リポジトリで行った対応）

- ギャラリー（[`GalleryView.vue`](../frontend/src/views/GalleryView.vue)）: 更新ボタンで debounce 解除、`list` の watch を浅くする、`getVRChatConfig` が失敗した場合も空パスと同様にデフォルト保存先へフォールバック。
- サムネユースケース（[`media_thumbnail.go`](../internal/usecase/media_thumbnail.go)）: `prepareScreenshotThumbnailJPEG` に集約し二重検証と TOCTOU を解消。
