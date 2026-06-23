# ローカル実データ置き場

開発・検証用のログや DB など、リポジトリに含めない実データを置くディレクトリです。
中身は `.gitignore` で除外されています。

## WSL で開発し Windows で動作確認する場合（推奨）

Windows 上の本番データをコピーせず、**symlink でそのまま参照**できます。

```bash
cp var/local.env.example var/local.env
# var/local.env の VRCTWEAKER_WIN_USER を Windows のユーザー名に合わせる

make link-var
```

張られるリンク:

| リポジトリ内 | 実体（Windows /mnt/c 経由） |
|--------------|----------------------------|
| `var/data/win/` | `%AppData%\Roaming\vrchat-tweaker\`（DB など） |
| `var/logs/vrchat/` | `%USERPROFILE%\AppData\LocalLow\VRChat\VRChat\` |
| `var/logs/latest-output_log.txt` | 上記フォルダ内で最新の `output_log_*.txt` |

Agent や手元の調査では次のパスを使ってください。

- DB: `var/data/win/vrchat-tweaker.db`
- 最新ログ: `var/logs/latest-output_log.txt`

ログファイル名が変わったあとは `make link-var` を再実行すると `latest-output_log.txt` が更新されます。

## ディレクトリ（手動利用）

| パス | 用途の例 |
|------|----------|
| `var/data/` | SQLite DB、認証トークンなど |
| `var/logs/` | `output_log.txt` のコピーやテスト用ログ |

## 補足

Wails アプリ本体のデータ保存先は変更しません（Windows 実行時は従来どおり `%AppData%\Roaming\vrchat-tweaker\`）。
`var/` は WSL 上の Cursor Agent や CLI から Windows 側の実データを読むための窓口です。
