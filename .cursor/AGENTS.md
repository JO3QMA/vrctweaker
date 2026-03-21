# VRChat Tweaker - Agent 利用ガイド

## Issue から PR まで（GitHub / ローカル Issue メモ）

GitHub Issue の URL・番号、または `docs/ai_dlc/issues/` 等の Issue ドキュメントを渡して「実装して」「PR まで」と依頼する場合は、**issue-to-pr-workflow** Skill がオーケストレーターになる。

流れは **ブランチ作成 → TDD 実装 → `make fmt/test/lint` → レビュー → PR**（各所でユーザー確認）。手順の本体は `.cursor/commands/` の Markdown（`/create-branch` 等で単体実行も可）。

| 段階 | 参照 |
|------|------|
| ブランチ | `commands/create-branch.md` |
| 検証 | `commands/run-verify.md`（`qa` エージェントと同等コマンド） |
| レビュー | `commands/run-review.md` + `agents/reviewer.md` |
| PR | `commands/make-pr.md` |

詳細は `.cursor/skills/issue-to-pr-workflow/SKILL.md`。

## パイプライン自動実装

`docs/features` の機能を実装する際は、**feature-implementation-pipeline** Skill を使用する。

パイプラインは **TDD**（テスト駆動開発）に則り、**テストを先に書く**。実装完了後は **fmt → テスト → Lint** の検証ループを、全パスするまで繰り返す。

### 使い方

1. **単一機能の実装**
   ```
   docs/features/ui-gallery-view.md の機能を実装して
   ```

2. **パイプライン明示**
   ```
   feature-implementation-pipeline で activity-log-monitoring-output_log を実装して
   ```

3. **全機能一括実行**
   ```
   docs/features の機能を一括で実装して
   ```
   → 依存順に全15機能を Plan → Build → Review → QA で順次実行

### パイプラインの流れ

| ステップ | Agent | 役割 |
|----------|-------|------|
| 1. Plan | planner | 仕様から実装計画を作成（テスト観点含む） |
| 2. Build | implementer | TDD で実装（テスト先、単体テスト必須） |
| 3. Review | reviewer | 変更のコードレビュー |
| 4. QA | qa | fmt → テスト → Lint、失敗時は修正して再実行 |

各 Agent は `.cursor/agents/` に定義され、`mcp_task` で起動される。  
依存関係は `.cursor/skills/feature-implementation-pipeline/reference-dependencies.md` に明示。

## 手動で各 Agent を使う

- **計画だけ欲しい**: 「planner サブエージェントで ui-gallery-view の実装計画を立てて」
- **レビューだけ**: 「reviewer サブエージェントで変更をレビューして」
- **QA だけ**: 「qa サブエージェントでテストとlintを実行して」

## プロジェクトルール

- `.cursor/rules/post-change-fmt.mdc`: コード変更後は `make fmt`（常時適用）
- `.cursor/rules/tdd-workflow.mdc`: TDD と検証ループ（常時適用）
- `.cursor/rules/go-conventions.mdc`: Go の規約
- `.cursor/rules/vue-conventions.mdc`: Vue/TypeScript の規約

## スキル

- **issue-to-pr-workflow**: Issue 起点でブランチ〜実装〜検証〜レビュー〜PR までを順に制御（Commands を束ねるオーケストレーター）
- **tdd-workflow**: テスト駆動開発と fmt→test→lint の検証ループ。コード変更時に適用。
