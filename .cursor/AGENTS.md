# VRChat Tweaker - Agent 利用ガイド

## パイプライン自動実装

`docs/features` の機能を実装する際は、**feature-implementation-pipeline** Skill を使用する。

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
| 1. Plan | planner | 仕様から実装計画を作成 |
| 2. Build | implementer | 計画に基づきコード実装 |
| 3. Review | reviewer | 変更のコードレビュー |
| 4. QA | qa | Lint・テスト実行 |

各 Agent は `.cursor/agents/` に定義され、`mcp_task` で起動される。  
依存関係は `.cursor/skills/feature-implementation-pipeline/reference-dependencies.md` に明示。

## 手動で各 Agent を使う

- **計画だけ欲しい**: 「planner サブエージェントで ui-gallery-view の実装計画を立てて」
- **レビューだけ**: 「reviewer サブエージェントで変更をレビューして」
- **QA だけ**: 「qa サブエージェントでテストとlintを実行して」

## プロジェクトルール

- `.cursor/rules/go-conventions.mdc`: Go の規約
- `.cursor/rules/vue-conventions.mdc`: Vue/TypeScript の規約
