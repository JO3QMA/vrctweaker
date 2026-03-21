# ブランチ作成（Issue 起点）

GitHub Issue またはリポジトリ内の Issue ドキュメントから作業ブランチを切る。

## 前提

- [GitHub CLI](https://cli.github.com/)（`gh`）が利用可能であること（GitHub Issue の場合）
- ベースブランチは **`main`**（リポジトリのデフォルトが異なる場合は置き換える）

## 手順

### A. GitHub Issue の URL または番号がある場合

1. Issue 番号を特定する  
   - URL 例: `https://github.com/owner/repo/issues/123` → `123`  
   - 番号のみの場合はそのまま使う
2. リポジトリを特定する  
   - カレントディレクトリのリモートから: `gh repo view --json nameWithOwner -q .`  
   - または URL から `owner/repo` を抽出し、`gh issue view <番号> --repo owner/repo`
3. Issue 本文を取得する  

   ```bash
gh issue view <番号> --json title,body,labels,number
```

4. **ブランチ名**を決める（下記「命名規則」）。スラグは英小文字・ハイフン、最大長に注意する。
5. ベースを最新にしてからブランチを作成する  

   ```bash
git fetch origin
git checkout main
git pull origin main
git checkout -b <ブランチ名>
```

### B. ローカルの Issue メモ（`docs/ai_dlc/issues/*.md` 等）のみの場合

1. 該当 Markdown を読み、タイトル・受け入れ条件を把握する
2. ファイル名またはタイトルから英語スラグを作る（例: `issue-10-server-status` → `server-status`）
3. ラベルが無いためプレフィックスは **`chore/`** または内容がバグ修正なら **`fix/`**、機能追加なら **`feature/`** を人間が選ぶ
4. 上記と同様に `main` から `git checkout -b <ブランチ名>`

## ブランチ命名規則

形式: `<プレフィックス>/<issue番号または略号>-<短い英語スラグ>`

GitHub のラベル（大文字小文字は正規化してマッチ）に応じたプレフィックス:

| ラベルの目安 | プレフィックス |
|-------------|----------------|
| `bug`, `fix` | `fix/` |
| `enhancement`, `feature`, `type: feature` | `feature/` |
| 上記以外・不明 | `chore/` |

例:

- `feature/123-add-server-status-indicator`
- `fix/45-ytdlp-cookie-path`
- `chore/10-server-status`（ローカル docs のみで番号をファイル名から取った場合）

## 注意

- ブランチ作成前にユーザーへブランチ名の確認を取る（Skill 側の確認ポイントと併用）
- 既に同名ブランチがある場合は別名を提案する
