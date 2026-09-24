# リリース前スモークチェックリスト（Windows）

頒布用 zip を BOOTH / GitHub Release に載せる前の **手動確認** 用メモです。

## ビルド・公開の流れ（メンテナ向け）

1. `wails.json` の `info.productVersion` と `CHANGELOG.md` の版セクションを更新し、main にマージする。
2. タグ `v<productVersion>` を push する（例: `v0.1.0`）。`.github/workflows/release.yml` が走る。
3. CI が `make test` 相当（Go + フロント Vitest）のあと、**Windows のみ** ビルドし zip を作成する。
4. GitHub に **draft Release** が作られる（従来どおり）。Release ノートに **zip の SHA256** と CHANGELOG 抜粋が入る。
5. 下記スモークを **draft の zip** で実施し、問題なければ Release を **Publish** する。
6. **同じ zip ファイル** を BOOTH の商品ファイルにアップロードし、説明文に版番号と zip の SHA256 を記載する。

ローカルで zip だけ作る場合: `make release-zip`（成果物は `dist/vrchat-tweaker-v<version>-windows-amd64.zip`）。

### zip の中身

| ファイル | 内容 |
|----------|------|
| `vrchat-tweaker.exe` | アプリ本体 |
| `LICENSE` | MIT |
| `README.txt` | 利用者向け短文 + Getting Started / Releases リンク |
| `checksums.txt` | 同梱ファイル（exe / LICENSE / README.txt）の SHA256 |

## 環境

- [ ] クリーンな Windows 10/11 VM またはテスト PC
- [ ] 配布予定と **同一の zip**（SHA256 を Release ノートと照合）

## チェック項目

- [ ] zip 展開 → `vrchat-tweaker.exe` 起動
- [ ] Settings で **アプリのバージョン** が Release タグと一致
- [ ] UI 言語切替（少なくとも ja / en）
- [ ] VRChat ログイン → ログアウト（資格情報が残らないこと）
- [ ] Gallery / Activity / Friends のいずれかでデータ表示（環境に応じて）
- [ ] ウィンドウを閉じたときのタスクトレイ動作（既定設定）

## 記録

問題があれば GitHub Issue に **再現手順のみ**（個人を特定できる情報なし）で記録してください。

関連: [はじめに](./getting-started.md)
