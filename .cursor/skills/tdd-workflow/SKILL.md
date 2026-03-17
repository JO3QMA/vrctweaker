---
name: tdd-workflow
description: >-
  Enforces Test-Driven Development and verification loop when making code
  changes. Use when implementing features, writing new code, or after completing
  implementation. Tests first, mandatory unit tests, fmt→test→lint loop until
  all pass.
---

# TDD ワークフロースキル

コード変更時に、テスト駆動開発（TDD）と検証ループを適用する。

## トリガー

- コードの実装・変更を行うとき
- 「実装して」「機能を追加して」「TDDで」と依頼されたとき
- feature-implementation-pipeline の Build ステップ

## 実行フロー

### Phase 1: テストファースト（TDD）

1. **実装前にテストを書く**
   - 期待する振る舞いをテストで定義する
   - 実行すると失敗する（Red）状態にする
2. **最小限の実装でテストをパスさせる**（Green）
3. **リファクタリング**（Refactor）

単体テストは**必ず**作成する。新規の関数・メソッド・コンポーネントには対応するテストを追加する。

### Phase 2: 検証ループ

実装完了後、以下を**順に**実行する：

```
1. fmt   → 2. test  → 3. lint
     ↑                    |
     |____ エラー時、修正して 1 へ ____|
```

**コマンド**（プロジェクトルートは `/workspaces/vrctweaker`）:

```bash
# 1. フォーマット
make fmt
# または個別: go fmt ./... && cd frontend && pnpm run format

# 2. テスト
make test
# または個別: go test -v -race -cover ./internal/... && cd frontend && pnpm run test

# 3. Linter
make lint
# または個別: golangci-lint run ./... && cd frontend && pnpm run lint && pnpm exec vue-tsc --noEmit
```

**エラー時**: 原因を修正し、**1 から再実行**。全パスまで繰り返す。

## Subagent 利用

検証ループの実行には `mcp_task` で `shell` サブエージェントを使用：

```yaml
subagent_type: shell
description: Run fmt, tests, and linter
prompt: |
  cd /workspaces/vrctweaker で以下を順に実行せよ。
  1. make fmt
  2. make test
  3. make lint
  エラーがあればその内容を報告せよ。
```

失敗時は `generalPurpose` の Implementer に修正を依頼し、再度 shell で検証を実行する。

## 参照

- [.cursor/rules/tdd-workflow.mdc](../../rules/tdd-workflow.mdc): 常時適用のTDDルール
- feature-implementation-pipeline: 本ワークフローを Build と QA に組み込み済み
