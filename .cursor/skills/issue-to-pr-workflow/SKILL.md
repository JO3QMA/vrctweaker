---
name: issue-to-pr-workflow
description: >-
  GitHub Issue の URL・番号、または作業対象の Issue を渡されて開発を開始する。
  ブランチ作成、TDD での実装、fmt/test/lint 検証、コードレビュー、PR 作成までの
  一連のフローを順に制御する。オーケストレーターとして各 Cursor command を参照する。
---

# Issue → PR 開発ワークフロー

[弥生開発者ブログ: Skills × Commands の考え方](https://tech-blog.yayoi-kk.co.jp/entry/2026/03/04/110000) に沿い、**手順の詳細は command ファイル**に置き、本 Skill は **順序と確認ポイント**のみを担う。

## When to Use

- GitHub の Issue URL や番号が含まれ、「この Issue を実装して」「作業を開始して」「PR まで出して」等と依頼されたとき
- リポジトリ内の Issue メモ（例: `docs/ai_dlc/issues/issue-*.md`）を渡され、同様のフローで進めたいとき

## Instructions

### 全体の流れ

各ステップの直後に **確認ポイント** でユーザーの判断を挟む。`gh` や `git push` などのコマンド実行前は Cursor の確認 UI に従う。

```text
Step 1: ブランチ作成     → 【確認】ブランチ名
Step 2: 実装（TDD）      → 【確認】方針が割れたとき / 完了報告
Step 3: 検証             → 【確認】失敗時は修正方針
Step 4: コードレビュー   → 【確認】Critical の扱い
Step 5: PR 作成          → 【確認】タイトル・本文 → 承認後に gh 実行
```

### Step 1: ブランチ作成

`.cursor/commands/create-branch.md` の手順に従う。

**確認ポイント**: ブランチ名を提示し、作成前にユーザーへ「この名前でよいか」を確認する。

### Step 2: 実装

- Issue（またはローカル Issue メモ）の受け入れ条件・本文を満たすコードを書く
- **TDD**: `.cursor/rules/tdd-workflow.mdc` および `.cursor/agents/implementer.md` の原則に従う（テストを先に書く、単体テスト必須）
- Go / Vue の規約は `.cursor/rules/go-conventions.mdc`、`.cursor/rules/vue-conventions.mdc` に従う

**確認ポイント**: 要件が曖昧なときは解釈案を出して確認する。実装完了後、コミットメッセージに Issue 番号を含める（例: `fix: ... (#123)`）。

### Step 3: 検証

`.cursor/commands/run-verify.md` の手順に従い、`make fmt` → `make test` → `make lint` を全パスするまで繰り返す。

**確認ポイント**: 連続して失敗する場合は原因整理をユーザーと共有する。

### Step 4: コードレビュー

`.cursor/commands/run-review.md` の手順に従う。レビュー観点は `.cursor/agents/reviewer.md`。

**確認ポイント**: Critical がある場合は修正するかどうかをユーザーが判断できるよう、選択肢を明示する。修正したら Step 3 に戻る。

### Step 5: PR 作成

`.cursor/commands/make-pr.md` の手順に従う。

**確認ポイント**: `gh pr create` は、タイトル・本文を提示したうえで **ユーザーが承認した後** に実行する。勝手に PR を開かない。

## 個別実行

フロー全体ではなく一部だけ行う場合は、Cursor の **Commands** から単体実行する（例: `/make-pr` は PR 作成のみ）。

| Command ファイル | 用途 |
|------------------|------|
| `create-branch.md` | ブランチ作成のみ |
| `run-verify.md` | fmt / test / lint のみ |
| `run-review.md` | 差分レビューのみ |
| `make-pr.md` | プッシュ + PR のみ |

## 既存スキルとの役割分担

- **`feature-implementation-pipeline`**: `docs/features` の機能ドキュメント向け（Plan→Build→Review→QA の自動パイプライン）
- **本 Skill**: **Issue 起点**の開発（GitHub またはローカル Issue メモ）から PR まで

## 制約（記事・プロダクトの前提）

- スキル間の自動チェーン起動は Cursor 側の制約により前提にしない
- シェルコマンドは実行前確認が入ることがある。頻用コマンドは Cursor の許可リスト設定を検討する
