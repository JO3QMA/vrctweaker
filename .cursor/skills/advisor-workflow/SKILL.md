---
name: advisor-workflow
description: >-
  Composer-led development with short advisory turns on Claude Sonnet or Codex
  when stuck. Use when architecture forks, ambiguous acceptance criteria, QA
  loops fail repeatedly, or before large refactors. Read advisor agent and Chat
  or mcp_task paths; max 2 advisory rounds per feature or Issue.
---

# Advisor ワークフロー（Composer メイン・Sonnet / Codex 相談）

「アドバイザー戦略」に**相当する**運用を、**メインは Composer**、相談は **Claude Sonnet** または **Codex** で行う。

## 前提

- **オーケストレーション**（パイプライン進行・Implementer への再依頼など）は **Composer** 上のエージェントが行う。
- **相談**は Composer **以外**の枠で行い、**短い要約だけ** Composer に戻して実装を継続する。
- モデル指定は **Cursor の UI** に依存する。相談用 Chat / SubAgent 起動時に、利用可能なら **Sonnet** または **Codex** を選ぶ。

## When to Use

次のいずれかに該当したら相談を検討する。

- アーキテクチャやモジュール境界の**分岐**で自信が持てない
- 受け入れ条件・Issue 解釈が**曖昧**
- **大規模リファクタ**や破壊的変更の直前
- 同一機能・同一 Issue の **QA が連続して失敗**し、同じ修正ループに陥っている

## 相談回数の上限（max_uses 相当）

- **1 機能**（`docs/features` の 1 本）または **1 Issue** あたり、Advisor 相談は **合計 2 回まで**とする。
- **二段相談**（例: 1 回目 Sonnet → まだ分岐が残る → 2 回目 Codex）は **2 回枠の中**で行う（Sonnet + Codex で計 2 回まで）。

## 手順 A: Chat 経由（手軽）

1. Composer で詰まったら、**新規 Chat** を開く。
2. モデルを **Claude Sonnet** または **Codex** に切り替える（利用可能なら）。
3. `.cursor/agents/advisor.md` の**出力形式**に従うよう指示し、Composer 側でまとめた**状況要約・選択肢・失敗ログ**を貼る。
4. 返答の **Recommendation / Alternative / Risks / Stop** を Composer に**要約して貼り戻し**、実装を継続する。

## 手順 B: `mcp_task`（SubAgent）経由

1. `.cursor/agents/advisor.md` を読み、本文（frontmatter 除く）を取得する。
2. `mcp_task` を呼ぶ:
   - `readonly`: `true`（相談専用・変更禁止）
   - `description`: `Advisory guidance for stuck implementation`
   - `prompt`: [advisor.md の本文] + 改行 + 「以下は Composer から渡すコンテキスト:」+ 改行 + [状況要約・選択肢・失敗ログ]
   - `subagent_type` の目安:

| 状況 | 推奨 `subagent_type` |
|------|------------------------|
| コードベース探索が主 | `explore` |
| 方針・トレードオフの整理が主 | `planner` |
| 両方まざる | `generalPurpose` |

3. 返答を Implementer / Composer 向けプロンプトに**追記**してから実装を再開する。

## 関連パイプライン

- 機能ドキュメント駆動: [feature-implementation-pipeline](../feature-implementation-pipeline/SKILL.md)（Build / QA で本 Skill を参照）
- Issue 起点: [issue-to-pr-workflow](../issue-to-pr-workflow/SKILL.md)（実装ステップで本 Skill を参照）

## Sonnet と Codexの使い分け

- 多くの場合は **どちらか一方**で十分。
- 二段が必要なら **先に Sonnet で整理 → まだ分岐が残るときだけ Codex**（またはその逆）とし、上記 **合計 2 回**を超えない。
