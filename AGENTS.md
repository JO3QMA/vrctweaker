# Agent guide (VRChat Tweaker)

リポジトリ横断のエージェント向け短いガイド。ドメイン用語は [`CONTEXT.md`](CONTEXT.md)、実装規約は [`.cursor/rules/`](.cursor/rules/) を参照。

## 日本語 UI 文言の表記ルール

- 英数字と日本語のあいだに半角スペースを入れない（例: 「yt-dlpの置き換え」「VRChatで」「14日」「2時間」）
- 英単語どうしのスペースはそのまま（例: 「VRChat Tools」）
- 句読点は全角「、」「。」を使う
- `{count}件` のように、プレースホルダと日本語のあいだも詰める

`frontend/src/i18n/locales/ja.json` の各文言は Vitest（`frontend/src/i18n/locales/ja.spacing.spec.ts`）で上記の**英数字↔日本語の半角スペース**を検査する。文言を追加・変更したらテストが通ることを確認する。

## 関連

- 公開 PR/Issue の redaction: [`docs/agents/redaction.md`](docs/agents/redaction.md)（`.cursor/rules/redaction-public-artifacts.mdc`）
- Issue → PR フロー: [`.cursor/skills/issue-to-pr-workflow/SKILL.md`](.cursor/skills/issue-to-pr-workflow/SKILL.md)
