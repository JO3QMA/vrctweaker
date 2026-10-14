# ADR 0025: アプリ自己更新（GitHub Releases）

## Status

**Accepted** — [Issue #29](https://github.com/JO3QMA/vrctweaker/issues/29)

## Context

- 製品バージョンは `wails.json` の `productVersion`（`internal/appversion`）が正本
- Windows 頒布物は Release ワークフローが `dist/vrchat-tweaker-v{version}-windows-amd64.zip` を公開し、リリース本文に zip の SHA256 を記載する
- 実行中の exe は Windows 上で直接上書きできないため、ヘルパー経由の置換と再起動が必要
- プレリリースは対象外。起動時の自動確認はユーザーがオフにできる

## Decision

1. **取得元**: 公開 GitHub API `GET /repos/JO3QMA/vrctweaker/releases/latest`（非 draft・非 prerelease のみ）
2. **検知**: 起動時（自動確認 ON 時）と設定画面の手動ボタン。semver 比較で現行より新しいときのみ通知
3. **同意**: フロントの更新ダイアログでリリースノート（プレーンテキスト・最大 500 文字）を表示し、承認後に更新
4. **Windows 更新**: zip をダウンロード → 本文の SHA256 で検証 → 展開 → PowerShell ヘルパーで exe 退避・置換・失敗時ロールバック → 新 exe 起動 → 現プロセス終了
5. **非 Windows**: 同じダイアログで通知し、「Release ページを開く」でブラウザを開く（自動置換なし）
6. **設定**: `app_update_auto_check`（既定 ON）、`dismissed_update_version` で「後で」スキップ
7. **レイヤー**: `internal/infrastructure/github`（API）、`internal/infrastructure/appupdate`（DL・検証・適用）、`internal/usecase`（オーケストレーション）、`internal/wailsapp`（バインディング・イベント）

## Consequences

- コード署名・MSIX 自動更新は本 ADR の範囲外（別 Issue）
- オフラインや API 失敗はログのみ（UI には出さない）
