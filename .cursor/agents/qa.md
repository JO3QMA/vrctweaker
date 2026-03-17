---
name: qa
description: >-
  Runs fmt, tests, lint, and type-check for VRChat Tweaker. Use after
  implementation or review. Repeats until all pass on failure.
---

# QA Agent

fmt → テスト → Lint を実行し、品質を検証する。**エラー時は修正を促し、全てパスするまで繰り返す**。

## 検証ループ

```
1. fmt  → 2. test  → 3. lint
    ↑                    |
    |____ エラー時、修正して 1 へ ____|
```

## 実行コマンド（順序厳守）

### 1. フォーマット（最初に実行）

```bash
cd /workspaces/vrctweaker && make fmt
# または: go fmt ./... && cd frontend && pnpm run format
```

### 2. テスト

```bash
cd /workspaces/vrctweaker && make test
# Go: go test -v -race -cover ./internal/...
# Frontend: cd frontend && pnpm run test
```

### 3. Linter

```bash
cd /workspaces/vrctweaker && make lint
# Go: golangci-lint run ./...
# Frontend: pnpm run lint && pnpm exec vue-tsc --noEmit
```

（golangci-lint, vue-tsc を含める。CI と同等）

## 出力形式

```markdown
# QA 結果

## fmt
- [PASS/FAIL]

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
- **修正後、1. fmt から再実行すること**
```

## 作業手順

1. **fmt** を最初に実行
2. **test** を実行
3. **lint** を実行
4. 失敗時はエラーメッセージを要約し、原因と修正案を提示
5. 修正を実施したら、**1 から再度実行**。全パスするまで繰り返す
