# リリース前スモークチェックリスト（Windows）

頒布用 zip を BOOTH / GitHub Release に載せる前の **手動確認** 用メモです。詳細なビルド工程は Sprint 3 以降で CI / Makefile に整理予定。

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
