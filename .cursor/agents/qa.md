---
name: qa
description: >-
  Runs tests, lint, and type-check for VRChat Tweaker. Use after implementation
  or review to verify code quality and prevent regressions.
---

# QA Agent

Lint・テスト・型チェックを実行し、品質を検証する。

## 実行コマンド

### Go

```bash
cd /workspaces/vrctweaker && go test -v -race -cover ./internal/...
cd /workspaces/vrctweaker && golangci-lint run ./...
```

### Vue / Frontend

```bash
cd /workspaces/vrctweaker/frontend && pnpm install --frozen-lockfile
cd /workspaces/vrctweaker/frontend && pnpm run lint
cd /workspaces/vrctweaker/frontend && pnpm exec vue-tsc --noEmit
cd /workspaces/vrctweaker/frontend && pnpm run test
```

（golangci-lint, vue-tsc を含める。CI と同等）

## 出力形式

```markdown
# QA 結果

## Go
- テスト: [PASS/FAIL] (概要)
- Lint: [PASS/FAIL]

## Frontend
- Lint: [PASS/FAIL]
- 型チェック: [PASS/FAIL]
- テスト: [PASS/FAIL]

## 問題があれば
- エラー内容の要約
- 修正の提案
```

## 作業手順

1. 上記コマンドを順に実行
2. 失敗時はエラーメッセージを要約
3. 原因と修正案を提示
4. 修正後は再実行を促す
