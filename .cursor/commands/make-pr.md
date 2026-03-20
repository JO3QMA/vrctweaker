# Pull Request 作成

現在のブランチをリモートにプッシュし、`gh` で Pull Request を開く。

## 前提

- コミットはすでにローカルにあり、検証（`run-verify.md`）とレビュー方針の整理が済んでいること
- `gh auth login` 済みであること

## 手順

1. リモートへプッシュする  

   ```bash
git push -u origin HEAD
```

2. **PR のタイトルと本文をドラフト**し、ユーザーに確認を取ってから作成する（承認後に実行）。

3. PR を作成する  

   ```bash
gh pr create --title "<タイトル>" --body "<本文>"
```

   インタラクティブにする場合は `gh pr create` のみでもよい。

## PR テンプレート（本文の目安）

```markdown
### 概要
<Issue の要約と変更の目的>

### 関連 Issue
close #<Issue 番号>

### 変更内容
- 

### テスト
- `make fmt` / `make test` / `make lint` を実行済み
- （あれば手動確認手順）
```

## 注意

- タイトル・本文の **`gh pr create` 実行前** に必ずユーザー承認を得る
- ドラフト PR にしたい場合は `gh pr create --draft ...`
