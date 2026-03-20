# 検証ループ（fmt / test / lint）

実装後の品質検証。プロジェクトの **TDD ワークフロー**と CI 相当のチェックに従う。

## 手順

リポジトリルートで、次を **この順序** で実行する。

1. **フォーマット**

   ```bash
cd /workspaces/vrctweaker && make fmt
```

2. **テスト**

   ```bash
cd /workspaces/vrctweaker && make test
```

3. **Lint**

   ```bash
cd /workspaces/vrctweaker && make lint
```

## 失敗時

- エラーを修正し、**1 から再度**実行する
- 詳細な観点は `.cursor/agents/qa.md` および `.cursor/rules/tdd-workflow.mdc` に従う

## 完了条件

- 上記 3 ステップがすべて成功するまで繰り返す
