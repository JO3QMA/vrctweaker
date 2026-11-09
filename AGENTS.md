# Agent guide (VRChat Tweaker)

リポジトリ横断のエージェント向け短いガイド。ドメイン用語は [`CONTEXT.md`](CONTEXT.md)、実装規約は [`.cursor/rules/`](.cursor/rules/) を参照。

## 日本語 UI 文言の表記ルール

- 英数字と日本語のあいだに半角スペースを入れない（例: 「yt-dlpの置き換え」「VRChatで」「14日」「2時間」）
- 英単語どうしのスペースはそのまま（例: 「VRChat Tools」）
- 句読点は全角「、」「。」を使う
- `{count}件` のように、プレースホルダと日本語のあいだも詰める

`frontend/src/i18n/locales/ja.json` の各文言は Vitest（`frontend/src/i18n/locales/ja.spacing.spec.ts`）で上記の**英数字↔日本語の半角スペース**を検査する。文言を追加・変更したらテストが通ることを確認する。

## Element Plus フォーム（`el-form`）

- **制約のある入力**（パス、起動引数、数値の範囲・必須など）には `el-form` の **`rules` と保存/適用前の `validate()`** を使う。エラー文言は i18n キー（`formValidation.*`）で各項目直下に表示する。
- **レイアウト目的だけ**の `el-form`（ラジオ群・チェックボックスのみなど）はそのままでよい。
- 共有ルールは `frontend/src/utils/elFormRules.ts`、検証ヘルパは `frontend/src/utils/elFormValidate.ts` を参照。

## 関連

- 公開 PR/Issue の redaction: [`docs/agents/redaction.md`](docs/agents/redaction.md)（`.cursor/rules/redaction-public-artifacts.mdc`）
- ダークモード固定: [`docs/agents/dark-mode-fixed.md`](docs/agents/dark-mode-fixed.md)
- Issue → PR フロー: [`.cursor/skills/issue-to-pr-workflow/SKILL.md`](.cursor/skills/issue-to-pr-workflow/SKILL.md)
