# ランチャー画面の使い勝手向上：起動引数のGUI化

**出典**: [GitHub Issue #5](https://github.com/JO3QMA/vrctweaker/issues/5)  
**関連**: docs/ai_dlc/VRChat 起動引数調査ガイド.md

## 概要

現在のランチャーでは起動引数を文字列で直接入力する必要があり、一般ユーザーにはハードルが高い。よく使う設定をチェックボックス等のGUIで直感的に設定できるようにする。

## 背景

- ユーザーは起動引数の存在や記法を知らないことが多い
- 公式ドキュメントやWikiを調べて入力するのは負担が大きい
- チェックボックスやトグルの「ポチポチ」操作で設定できるUIが望ましい

## 機能要件

### 1. GUI化する設定項目

| 設定 | 型 | 引数 | 説明 |
|------|-----|------|------|
| デスクトップモードで起動 | チェックボックス | `-no-vr` | VRヘッドセットが接続されていてもデスクトップモードで起動（公式引数） |
| 起動前にキャッシュをクリア | チェックボックス | （特別扱い） | 起動前に Cache-WindowsPlayer 等を削除。※VRChatに `--clear-cache` は公式に存在しないため、本アプリで起動前にディレクトリ削除を実行する。 convention: 引数文字列に `--clear-cache` を含めるとこの動作をトリガーし、VRChatには渡さない |
| フルスクリーン | トグル/Select | `-screen-fullscreen 1` / `0` | Unityスタイルのフルスクリーン制御（オプション） |

### 2. カスタム引数（上級者向け）

- 既存の「起動引数」入力欄は「カスタム引数」として残す
- GUIでカバーしきれない特殊な引数を入力可能
- 最終的な起動引数 = GUIで生成した引数 + カスタム引数の結合

### 3. データ構造と連動

- **永続化**: 既存の `LaunchProfile.arguments` に結合済み文字列をそのまま保存（スキーマ変更なし）
- **UI初期化**: 既存の `arguments` をパースし、`-no-vr` / `--clear-cache` / `-screen-fullscreen` の有無を検出してGUIに反映。残りをカスタム引数に表示
- **保存時**: GUIの状態 + カスタム引数をマージし、`arguments` として保存

### 4. キャッシュクリアの実装詳細

- `--clear-cache` が引数に含まれる場合、起動前に以下を削除:
  - `%USERPROFILE%\AppData\LocalLow\VRChat\VRChat\Cache-WindowsPlayer`（Windows）
  - `output_log_path` が設定済みなら、その親ディレクトリからの相対で Cache-WindowsPlayer を解決
- VRChatには `--clear-cache` を渡さない（公式に存在しないため）
- Linux は現状サポート対象外（パスが異なるため、将来対応）

## 受け入れ条件（DoD）

- [ ] デスクトップモード、キャッシュクリア、フルスクリーンの各GUIが LauncherView に追加されている
- [ ] GUIの変更が `arguments` と双方向に連動する（ロード時にパース、保存時にマージ）
- [ ] カスタム引数欄が残り、GUI外の引数を入力可能
- [ ] 起動時に `--clear-cache` が指定されていれば、起動前にキャッシュディレクトリを削除し、VRChatには渡さない
- [ ] 既存の `arguments` との後方互換性（手動で `--no-vr` 等を入力したプロファイルも正しくパースされる）
- [ ] 単体テスト（Go: パース/マージロジック、Vue: コンポーネント）が追加されている

## テスト観点

1. **Go**: 引数パース（`arguments` → NoVR, ClearCache, Fullscreen, Custom）のユニットテスト
2. **Go**: マージ（NoVR, ClearCache, Fullscreen, Custom → `arguments`）のユニットテスト
3. **Go**: 起動時の `--clear-cache` 検出とキャッシュ削除ロジックのテスト（モック/テンポディレクトリ使用）
4. **Vue**: LauncherView のGUI表示と双方向バインディングのテスト
5. **Vue**: 既存 arguments のロード時のパース反映のテスト

## 依存関係

- 既存: LauncherView.vue, LaunchProfile entity, launcher_usecase
- settings_usecase: output_log_path 取得（キャッシュパス解決用、オプション）
