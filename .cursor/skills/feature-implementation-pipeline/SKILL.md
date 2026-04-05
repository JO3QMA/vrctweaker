---
name: feature-implementation-pipeline
description: >-
  Orchestrates automated feature implementation from docs/features. Use when
  implementing features from feature documents, running Plan→Build→Review→QA
  pipeline, or automating the implementation workflow.
---

# Feature Implementation Pipeline

docs/features の機能仕様を読み、実装計画→実装→レビュー→QA のパイプラインを自動化する。

## パイプライン概要

1. **Plan**: 機能ドキュメントから実装計画を作成（タスクにテスト観点を含める）
2. **Build**: 計画に基づき**TDD で実装**（テストを先に書く、単体テストは必須）
3. **Review**: 変更をコードレビュー
4. **QA**: fmt → テスト → Lint を実行、失敗時は修正を依頼し**全パスまで繰り返す**

## トリガー

次のいずれかの場合はこのパイプラインを実行する：
- 「docs/features の機能を実装して」「一括で実装して」→ **全15機能を依存順に実行**
- 「feature-implementation-pipeline で○○を実装」→ 指定機能のみ
- 個別機能（例: ui-gallery-view.md）の実装を依頼されたとき

## 一括実行（全機能）

対象が明示されていない、または「一括」「すべて」の場合は、[reference-dependencies.md](reference-dependencies.md) の推奨順序で全15機能を順次パイプラインに通す。

1. `reference-dependencies.md` を読み、実装順序を取得
2. 各機能を順番に Plan → Build → Review → QA 実行
3. QA 失敗時はその機能の Implementer に修正を依頼し、再QA
4. 全機能完了まで繰り返す

## 実行フロー

### 事前確認

- 対象機能ファイル（docs/features/*.md）を特定する
- **一括時**: [reference-dependencies.md](reference-dependencies.md) の推奨順序を厳守
- **単一/複数指定時**: 依存関係を [reference-dependencies.md](reference-dependencies.md) で確認し、依存先を先に実装する

### Step 1: Planner（計画）

1. `.cursor/agents/planner.md` を読み、本文（frontmatter除く）を取得
2. `mcp_task` を呼ぶ:
   - `subagent_type`: `generalPurpose`
   - `description`: `Create implementation plan for {feature}`
   - `prompt`: [planner.md の本文] + 改行 + `対象: docs/features/{対象ファイル}.md を読んで実装計画をMarkdownで作成せよ。`

結果を `plan_output` として保持。

### Step 2: Implementer（実装・TDD）

1. `.cursor/agents/implementer.md` の本文を取得
2. `mcp_task` を呼ぶ:
   - `subagent_type`: `generalPurpose`
   - `description`: `Implement feature from plan using TDD`
   - `prompt`: [implementer.md の本文] + 改行 + `実装計画:` + 改行 + [Step 1 の plan_output] + 改行 + `上記に従い TDD（テストを先に書く、単体テスト必須）で実装せよ。`

**frontend の UI を変更・抽象化する場合**: `.cursor/rules/storybook-wails-ui.mdc` に従い **Storybook（`*.stories.ts` 等）も更新**する。

### Step 3: Reviewer（レビュー）

1. `.cursor/agents/reviewer.md` の本文を取得
2. `mcp_task` を呼ぶ:
   - `subagent_type`: `generalPurpose`
   - `description`: `Code review of implementation changes`
   - `prompt`: [reviewer.md の本文] + 改行 + `git diff で変更を確認し、 Critical/Suggestion/Nice でレビューせよ。`

### Step 4: QA（品質保証・検証ループ）

`mcp_task` で `shell` を起動:
- `subagent_type`: `shell`
- `description`: `Run fmt, Go tests, golangci-lint, Vue lint and tests`
- `prompt`: 以下を**順に**実行せよ:
  1. `cd /workspaces/vrctweaker && make fmt`
  2. `cd /workspaces/vrctweaker && make test`
  3. `cd /workspaces/vrctweaker && make lint`

（fmt を最初に実行。golangci-lint を必ず含める。CI と同等）

もしくは `generalPurpose` で `.cursor/agents/qa.md` の本文 + 上記コマンド実行指示を渡す。

**失敗時**:
1. エラー内容を報告する
2. Implementer に修正を依頼する（修正依頼をプロンプトに含める）
3. 修正後、**1. make fmt から再度 QA を実行**する
4. 全パスするまで 2→3 を繰り返す

## 並列と順序

- **順次実行**: Plan → Build → Review → QA は順番に行う
- **一括時**: 全15機能を依存順に 1 機能ずつパイプラインを通す（依存順は reference-dependencies.md を参照）

## 出力

各ステップの結果をユーザーに要約して伝える。QA失敗時は修正内容と再実行の提案を行う。
