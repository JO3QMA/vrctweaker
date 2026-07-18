# ADR 0011: Asset cache clear

## Status

Accepted（grill-with-docs で合意、[Issue #11](https://github.com/JO3QMA/vrctweaker/issues/11)）

## Context

- VRChat のアセットキャッシュが肥大化し、ディスクを空けたい（Issue #11。本文は当初空）
- Settings の DB メンテナンスには既に **Clear friends cache**（`users_cache`）等があるが、これは Tweaker DB であり **VRChat asset cache** ではない
- Config には `cache_directory` / `cache_size` / `cache_expiry_delay` の編集があるが、中身を消す操作は無い
- [Issue #208](https://github.com/JO3QMA/vrctweaker/issues/208) は同一ディレクトリの**使用量表示**。クリーンアップ操作は #208 の v1 スコープ外で、本 ADR（#11）が担う
- ユーザー任意パスの中身全削除は、誤設定時に Picture folder 等を消しうる

用語は [`CONTEXT.md`](../../CONTEXT.md) の **Config** セクション（**VRChat asset cache**、**Asset cache clear**、**Asset cache clear v1 scope**、**Config**、**Picture folder**）を正本とする。

## Decision

1. **対象**: 解決済み **VRChat asset cache** ディレクトリの**中身のみ**全削除。ディレクトリ自体は残す
2. **非対象**: Picture folder・Tools・`config.json`・Tweaker DB・Official yt-dlp cache。VRChat の `cache_size` / `cache_expiry_delay` による自動間引きとは別導線
3. **UI**: **Config** のキャッシュ設定カード内。Settings の DB メンテナンスには置かない
4. **パス正本**: ディスク上の**保存済み `config.json`**。Config 画面の未保存入力は使わない。`cache_directory` が空なら **`…/VRChat/VRChat/Cache-WindowsPlayer`**（データディレクトリ本体ではない。config.json / Tools を消さないため）
5. **確認**: 実行前に **1 回**の確認ダイアログ。解決済み絶対パス・再ダウンロードが必要になる旨・所要時間が長くなりうる旨を示す
6. **VRChat 起動中**: **拒否**（既存 `VRChatRunning`）。強制クリアはしない
7. **パスガード**: ボリュームルート・非ディレクトリ・空パスを拒否。解決パスが **Picture folder**（保存済み `picture_output_folder`、空なら既定）と同一なら拒否。解決パスが **VRChat データディレクトリ**（config.json の親）と同一なら拒否
8. **存在しないパス**: エラー（ディレクトリは作らない）。**空ディレクトリ**: 成功（削除 0）
9. **途中失敗**: エラーで止め、成功扱いにしない。ロールバックなし。残りは再実行で消せる
10. **実行 UI**: **同期 API + ボタン loading**。進捗％・キャンセルは v1 なし
11. **画面離脱**: バックエンドは完走。フロントは unmount 後に loading／トーストを更新しない
12. **成功フィードバック**: **削除エントリ数**（トップレベルエントリ。Settings クリア系と同型）。解放バイトは出さない（#208）
13. **API**: `ResolveVRChatAssetCachePath()` / `ClearVRChatAssetCache()`。フロントからパス文字列を渡して消す API は置かない
14. **v1 スコープ外**: [`CONTEXT.md`](../../CONTEXT.md) の **Asset cache clear v1 scope** を正本とする

## Failure modes（review-ready）

| 状況 | ユーザー | サーバ |
|------|----------|--------|
| VRChat 起動中 | 拒否メッセージ（終了を促す） | 削除しない。安定なエラー（i18n 可能なキーまたは英語フレーズ） |
| パスがボリュームルート／非ディレクトリ／空 | 拒否メッセージ | 削除しない |
| 解決パスが Picture folder と同一 | 拒否メッセージ（スクショ誤削除防止） | 削除しない |
| 解決パスが存在しない | エラー（パス不存在） | 作成しない |
| 空ディレクトリ | 成功トースト（0 件） | `(0, nil)` |
| 削除途中の権限／ロック失敗 | エラー表示。成功トーストなし | `error`。既削除分は残したまま |
| 削除成功 | 成功トースト（件数） | `(n, nil)` |
| 実行中に Config から離脱 | トースト／loading は更新しない | 削除は完走 |

## Return-value contract（review-ready）

### 書き込み — 例: `ClearVRChatAssetCache() (int64, error)`

| 層 | 業務上の拒否（起動中・危険パス・不存在） | 途中／infra 失敗 | 成功 |
|----|------------------------------------------|------------------|------|
| **Usecase / App** | **`error`**（削除しない） | **`error`** | **`(削除エントリ数, nil)`** |
| **Frontend** | 確認後または即時にエラー表示（`el-alert` または `ElMessage.error`。Config 既存トーンに合わせる） | 同上。成功トーストなし | `ElMessage.success`（件数） |

- パス解決・Picture folder 比較・`VRChatRunning` チェックは usecase に集約する
- フロントから「消したいパス文字列」を渡して消す API にはしない（保存済み config 解決が正本）
- 二重実行はボタン `loading` で抑止（v1）。バックエンドの単一フライトは必須としない

## Considered options

| 案 | 却下理由 |
|----|----------|
| Settings の DB メンテに置く | `users_cache` クリアと混同する |
| 画面上の未保存 `cache_directory` を消す | 意図しないフォルダを消しうる |
| `cache_directory` 空なら拒否 | 既定パス利用者が本命ユースケース |
| VRChat 起動中も警告つき許可 | ロック・部分削除が起きやすい |
| 進捗イベント + キャンセル | 契約が複雑。途中停止＝部分削除の説明が必要 |
| 途中失敗を部分成功トースト | 「クリアした」誤解が残る |
| ホーム／デスクトップ等のヒューリスティック拒否 | 正当なカスタム配置まで弾く |

## Consequences

- #208（使用量バー）とパス解決ヘルパを共有できるとよいが、**同一 PR 必須ではない**（#11 単独で既定パス解決を実装してよい）
- Windows 実機での大容量キャッシュ削除は PR の手動検証項目にする（CI では小規模 temp dir）
- 子エントリ削除はディレクトリごと `RemoveAll` 相当でよいが、**シンボリックリンクはターゲットへ辿って消さない**（リンク自体のみ削除）。実装時に OS 差分をテストまたはコメントで明示する

## Test plan（review-ready）

| 層 | 入力・状況 | 期待 | テスト名（案） |
|----|------------|------|----------------|
| Go | 子ファイルあり | 中身空、件数 > 0、ディレクトリ残存 | `TestClearVRChatAssetCache_clearsContents` |
| Go | 空ディレクトリ | `(0, nil)` | `TestClearVRChatAssetCache_emptyDir` |
| Go | パス不存在 | `error` | `TestClearVRChatAssetCache_missingDir` |
| Go | VRChat running | `error`、削除なし | `TestClearVRChatAssetCache_vrchatRunning` |
| Go | パスが Picture folder と同一 | `error`、削除なし | `TestClearVRChatAssetCache_sameAsPictureFolder` |
| Go | ボリュームルート | `error` | `TestClearVRChatAssetCache_volumeRoot` |
| Go | `cache_directory` 空 | 既定パスを解決して削除 | `TestClearVRChatAssetCache_defaultPath` |
| Go | 途中で子の削除失敗 | `error`、成功件数を成功扱いしない | `TestClearVRChatAssetCache_partialFailure` |
| Vue | 確認キャンセル | API 未呼出 | `skips asset cache clear when confirmation cancelled` |
| Vue | 成功 | 件数トースト | `shows asset cache clear success count` |
| Vue | 起動中エラー | エラー表示 | `shows asset cache clear blocked when vrchat running` |
