# 実装計画: ランチャー画面の起動引数GUI化

## 概要

- **対象ドキュメント**: docs/features/launcher-launch-args-gui.md
- **ゴール**: 起動引数をチェックボックス・トグルで直感的に設定でき、カスタム引数と結合して永続化・起動時に反映する

## タスク一覧（TDD: 各タスクは「テスト→実装」の順）

### インフラ/ドメイン層（Go）

1. **LaunchArgsParsed 構造体の定義と ParseLaunchArgsForGUI** - ファイル: internal/domain/launcher/launch_args.go - テスト: internal/domain/launcher/launch_args_test.go
2. **MergeLaunchArgsForGUI の実装** - ファイル: internal/domain/launcher/launch_args.go - テスト: internal/domain/launcher/launch_args_test.go
3. **キャッシュディレクトリ解決ロジック** - ファイル: internal/usecase/launcher_usecase.go（または internal/infrastructure/vrchat/cache_clearer.go） - テスト: internal/usecase/launcher_usecase_test.go
4. **起動前に --clear-cache 検出・キャッシュ削除・引数フィルタ** - ファイル: internal/usecase/launcher_usecase.go - テスト: internal/usecase/launcher_usecase_test.go

### Wails バインディング

5. **ParseLaunchArgsForGUI / MergeLaunchArgsForGUI のバインディング追加** - ファイル: app.go, bindings.go, wails bindings - テスト: 既存 app テストまたは手動

### フロントエンド

6. **LaunchArgsParsedDTO 型と Parse/Merge 呼び出し** - ファイル: frontend/src/wails/app.ts - テスト: 型チェック (vue-tsc)
7. **LauncherView に GUI 項目追加（デスクトップモード・キャッシュクリア・フルスクリーン・カスタム引数）** - ファイル: frontend/src/views/LauncherView.vue - テスト: frontend/src/views/__tests__/LauncherView.spec.ts
8. **arguments との双方向連動（ロード時パース、保存時マージ）** - ファイル: frontend/src/views/LauncherView.vue - テスト: frontend/src/views/__tests__/LauncherView.spec.ts

## 変更ファイル

| 種別 | パス | 内容の要点 |
|------|------|------------|
| 新規 | internal/domain/launcher/launch_args.go | ParseLaunchArgsForGUI, MergeLaunchArgsForGUI。NoVR(-no-vr/--no-vr), ClearCache(--clear-cache), Fullscreen(-screen-fullscreen 0/1), Custom のパースとマージ |
| 新規 | internal/domain/launcher/launch_args_test.go | パース・マージのユニットテスト。後方互換（手動入力 --no-vr 等）のテスト |
| 変更 | internal/usecase/launcher_usecase.go | LaunchVRChat/LaunchToWorld で --clear-cache 検出→キャッシュ削除→VRChatには渡さない。outputLogPath を引数で受け取りキャッシュパス解決 |
| 変更 | internal/usecase/launcher_usecase_test.go | clear-cache 動作のテスト（テンポディレクトリ使用） |
| 変更 | app.go | LaunchVRChat で output_log_path 取得し launcher に渡す。ParseLaunchArgsForGUI, MergeLaunchArgsForGUI バインディング追加 |
| 変更 | bindings.go | LaunchArgsParsedDTO 型定義 |
| 変更 | frontend/src/wails/app.ts | ParseLaunchArgsForGUI, MergeLaunchArgsForGUI 呼び出し。LaunchArgsParsedDTO 型 |
| 変更 | frontend/src/views/LauncherView.vue | デスクトップモード・キャッシュクリア・フルスクリーンのチェック/トグル、カスタム引数欄。ロード時パース、保存時マージ |
| 新規 | frontend/src/views/__tests__/LauncherView.spec.ts | LauncherView の GUI 表示・双方向バインディング・パース反映のテスト |

## 依存関係

- **先行実装が必要**: なし
- **参照する既存コード**:
  - `internal/usecase/launcher_usecase.go` - parseLaunchArgs, launchWindowsWithArgs, launchLinuxWithArgs
  - `internal/usecase/settings_usecase.go` - GetOutputLogPath, GetPathSettings
  - `frontend/src/views/LauncherView.vue` - 既存のプロファイル編集UI
  - `app.go` - LaunchVRChat, SaveLaunchProfile
  - `bindings.go` - LaunchProfileDTO, toLaunchProfile

## 実装順序

1. **インフラ層（パース・マージロジック）**
   - internal/domain/launcher/launch_args.go 作成
   - Parse: `arguments` → `{ NoVR, ClearCache, Fullscreen, Custom }`
   - Merge: `{ NoVR, ClearCache, Fullscreen, Custom }` → `arguments`
   - 対応する引数: `-no-vr` / `--no-vr`, `--clear-cache`, `-screen-fullscreen 0/1`

2. **ドメイン/ユースケース**
   - LauncherUseCase に outputLogPath を渡す（または SettingsUseCase を注入）
   - LaunchVRChat / LaunchToWorld 内で:
     - args に `--clear-cache` が含まれるか判定
     - 含まれる場合: キャッシュパスを解決して削除、args から `--clear-cache` を除去
     - 除去後の args で VRChat 起動
   - キャッシュパス: `%USERPROFILE%\AppData\LocalLow\VRChat\VRChat\Cache-WindowsPlayer`（output_log_path の親ディレクトリ基準の相対も可、Linux は将来対応）

3. **Wails バインディング**
   - `ParseLaunchArgsForGUI(args string) → LaunchArgsParsedDTO`
   - `MergeLaunchArgsForGUI(dto) → string`
   - LaunchArgsParsedDTO: `{ noVr, clearCache, fullscreen, custom }`

4. **フロント UI**
   - LauncherView: 起動引数欄を「GUI 項目 + カスタム引数」に分割
   - プロファイル選択時: `App.ParseLaunchArgsForGUI(selected.arguments)` で GUI 状態を初期化
   - 保存時: `App.MergeLaunchArgsForGUI({noVr, clearCache, fullscreen, custom})` で arguments 文字列を生成して保存

## 補足: パース対象の引数と後方互換

| GUI項目 | 検出する引数 | 備考 |
|---------|-------------|------|
| デスクトップモード | `-no-vr`, `--no-vr` | 両方対応（後方互換） |
| キャッシュクリア | `--clear-cache` | 本アプリ独自。VRChat には渡さない |
| フルスクリーン | `-screen-fullscreen 1` | 1=有効, 0=無効。省略時はオフ扱い |
| カスタム | 上記以外 | そのまま残す |

## 検証ループ（完了時に実行）

- `make fmt` / `go fmt ./...` / `cd frontend && pnpm run format`
- `make test` / `go test -v -race -cover ./internal/...` / `cd frontend && pnpm run test`
- `make lint` / `golangci-lint run ./...` / `cd frontend && pnpm run lint && pnpm exec vue-tsc --noEmit`
