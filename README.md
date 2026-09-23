# VRChat Tweaker

VRChat 周辺のログ・スクリーンショット・フレンド情報などをまとめて扱う **Windows 向け** デスクトップアプリです。  
[Wails v2](https://wails.io/)（Go バックエンド + Vue 3 フロントエンド）で構築しています。

## 利用者向け

| 項目 | 内容 |
|------|------|
| 対応 OS | **Windows 10 / 11** のみ（Linux / macOS は未サポート） |
| 配布 | **BOOTH** および **[GitHub Releases](https://github.com/JO3QMA/vrctweaker/releases)** から Windows 用 zip |
| はじめ方 | **[Getting Started（利用ガイド）](./docs/user/getting-started.md)** — インストール、初回起動、ログイン、データ保存場所、トラブルシュート |

### 主な機能

| 画面 | 概要 |
|------|------|
| **Dashboard** | ホーム / 概要 |
| **Launcher** | VRChat の起動補助 |
| **Gallery** | スクリーンショットの閲覧・検索（日付グループ、ワールド検索、期間フィルタ） |
| **Activity** | Output log からの遭遇ログ・プレイ時間など |
| **Friends** | フレンド一覧・プロフィール |
| **Automation** | 自動化設定 |
| **Settings / Config** | アプリ・VRChat 連携の設定 |

### VRChat 非公式・データの取り扱い

- **VRChat Tweaker は VRChat 公式の製品ではありません。** VRChat 株式会社とは無関係の非公式ツールです。
- VRChat にログインする場合、セッション情報は **Windows の資格情報ストア等** に保存されます。アプリの SQLite データベースや設定は **`%AppData%\Roaming\vrchat-tweaker\`** に保存されます。詳細は [Getting Started](./docs/user/getting-started.md#データの保存場所windows) を参照してください。

### サポート

不具合・要望は **[GitHub Issues](https://github.com/JO3QMA/vrctweaker/issues)** から（テンプレート利用推奨）。Issue 本文に VRChat の表示名・ユーザー ID 等の個人情報を載せないでください。

---

## 開発者向け

ドメイン用語は [`CONTEXT.md`](./CONTEXT.md)、設計判断は [`docs/adr/`](./docs/adr/) を参照してください。

### 技術スタック

- **Backend**: Go 1.25, SQLite, Wails v2
- **Frontend**: Vue 3, TypeScript, Element Plus, Vite, Vitest, Playwright, Storybook

### 開発環境

#### 推奨: Dev Container

VS Code / Cursor の **Reopen in Container** で `.devcontainer/` を開くと、Go・pnpm・Playwright・`gh` が揃います。

WSL 上でホストの VRChat フォルダ（`/mnt/c/...`）をマウントする場合は、`.devcontainer/wsl-host-vrchat/` の構成を選んでください。

#### 手動セットアップ

| ツール | 用途 |
|--------|------|
| Go 1.25+ | バックエンド |
| [Wails CLI v2](https://wails.io/docs/gettingstarted/installation) | デスクトップアプリのビルド・開発 |
| Node.js 20+ / pnpm | フロントエンド |
| golangci-lint | Go の Lint |
| [Lefthook](https://github.com/evilmartians/lefthook) | Git pre-commit hooks（任意・推奨） |

```bash
# 依存関係
make install-front

# pre-commit（golangci-lint / go test compile / vue-tsc）— clone 後に一度
make setup-hooks

# 開発サーバー（ヘッドレス環境では xvfb 付き）
make dev-wails

# 品質チェック（コミット前の目安）
make fmt
make test
make lint

# フロントの src を変更したときは E2E も
make setup-e2e   # 初回のみ
make test-e2e
```

`make help` で Makefile の全ターゲットを確認できます。

### リポジトリ構成（抜粋）

```
.
├── app.go                 # Wails エントリ・バインディング
├── internal/              # Go（domain / usecase / infrastructure）
├── frontend/              # Vue アプリ
├── docs/
│   ├── adr/               # Architecture Decision Records
│   ├── agents/            # Issue トラッカー・トリアージ・エージェント向け規約
│   ├── features/          # 機能仕様
│   └── user/              # 利用者向けドキュメント
├── CONTEXT.md             # ドメイン用語集
└── .cursor/               # Cursor エージェント用ルール・スキル
```

### コントリビューション

#### Issue

バグ報告・機能要望は [GitHub Issues](https://github.com/JO3QMA/vrctweaker/issues) から作成してください。テンプレートに沿って記入すると、再現や受け入れ条件の確認がしやすくなります。

| テンプレート | 用途 |
|--------------|------|
| Bug report | 不具合（`bug` ラベル） |
| Feature request | 機能追加・改善（`enhancement` ラベル） |

新規 Issue には `needs-triage` が付きます。トリアージ用ラベルの意味は [`docs/agents/triage-labels.md`](./docs/agents/triage-labels.md) を参照してください。

**プライバシー**: Issue 本文に VRChat の表示名・ユーザー ID・ログの生データを載せないでください。必要ならマスクするか、再現手順だけを書いてください。

#### Pull Request

`main` 向けの PR では、`.github/pull_request_template.md` のチェックリストを埋めてください。

- 関連 Issue があれば `Closes #123` などでリンク
- コード変更後は `make fmt` → `make test` → `make lint`（フロント `src` 変更時は `make test-e2e` も）
- `make setup-hooks` 済みなら commit 時に Go lint / テスト compile / フロント型チェックが自動実行される
- UI 変更時は Storybook 側の更新も検討（`.cursor/rules/storybook-wails-ui.mdc`）

エージェント駆動で開発する場合は [`.cursor/AGENTS.md`](./.cursor/AGENTS.md) と [`docs/agents/`](./docs/agents/) も参照してください。

### CI

`main` への push / PR で `.github/workflows/ci.yml` が `make lint` と `make test` を実行します。

## ライセンス

本リポジトリは **[MIT License](./LICENSE)** の下で公開されています。
