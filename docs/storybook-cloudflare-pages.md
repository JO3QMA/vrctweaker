# Storybook on Cloudflare Pages

VRChat Tweaker の Storybook 静的サイトを [Cloudflare Pages](https://pages.cloudflare.com/) に公開し、ブラウザで UI レビューできるようにする手順です。Wails アプリ本体はデプロイしません。

## 公開 URL の例

| 種別 | URL パターン |
|------|----------------|
| 本番（`main`） | `https://vrctweaker-storybook.pages.dev` |
| PR / ブランチプレビュー | `https://<branch>.vrctweaker-storybook.pages.dev` |
| デプロイごとのプレビュー | `https://<hash>.vrctweaker-storybook.pages.dev` |

プロジェクト名を変えた場合は `vrctweaker-storybook` をその名前に置き換えてください。

## 1. Cloudflare API トークン

1. [Cloudflare Dashboard](https://dash.cloudflare.com/) → **My Profile** → **API Tokens** → **Create Token**
2. **Edit Cloudflare Workers** テンプレート、またはカスタムトークンで次を付与:
   - **Account** → **Cloudflare Pages** → **Edit**
3. トークンをコピー（再表示不可のため安全な場所に保管）

アカウント ID は **Workers & Pages** → 右サイドバー **Account ID** から取得します。

## 2. GitHub シークレット

リポジトリ **Settings → Secrets and variables → Actions → Secrets** に追加:

| 名前 | 内容 |
|------|------|
| `CLOUDFLARE_API_TOKEN` | 上で作成した API トークン |
| `CLOUDFLARE_ACCOUNT_ID` | Cloudflare アカウント ID |

環境ごとに分けたい場合は **Environments** で `storybook-pages` などを作り、同じ名前のシークレットをそこに置き、`.github/workflows/storybook-pages.yml` の deploy ジョブに `environment: storybook-pages` を追加してください。

## 3. （任意）Pages プロジェクト名

既定のプロジェクト名は `vrctweaker-storybook` です。変更する場合:

- **Variables** に `CLOUDFLARE_PAGES_PROJECT_NAME` を追加する  
  または
- Cloudflare 側で同名の Pages プロジェクトを先に作成する

初回デプロイ時、Wrangler がプロジェクトを自動作成する場合もあります（アカウント設定によります）。

## 4. ワークフロー

`.github/workflows/storybook-pages.yml` が次のタイミングで動きます:

- `main` への push（`frontend/**` または当該 workflow の変更）
- `main` 向け pull request（同上）
- 手動 **workflow_dispatch**

手順:

1. `cd frontend && pnpm install --frozen-lockfile && pnpm run build-storybook`
2. `frontend/storybook-static` を Cloudflare Pages にデプロイ（[wrangler-action](https://github.com/cloudflare/wrangler-action) 経由）

`main` は本番ブランチ、`pull_request` はプレビューデプロイ（ブランチ別 URL）になります。

## 5. ローカルでの確認

```bash
cd frontend
pnpm install
pnpm run storybook          # 開発サーバー (http://localhost:6006)
pnpm run build-storybook    # 静的ビルド → storybook-static/
```

ローカルに Wrangler は不要です。デプロイは GitHub Actions のみで完結します。

## 関連

- Storybook 運用: `.cursor/rules/storybook-wails-ui.mdc`
- CI（Lint / テスト）: `.github/workflows/ci.yml`
