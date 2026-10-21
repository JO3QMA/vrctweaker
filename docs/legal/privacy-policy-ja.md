# プライバシーポリシー（VRChat Tweaker）

最終更新: 2026-10-14

本ポリシーは、Windows 向けデスクトップアプリ **VRChat Tweaker**（以下「本アプリ」）が、お使いの端末上でどのデータを扱い、どの通信を行うかを、**リポジトリ上の実装に基づいて**説明します。本アプリは VRChat 株式会社の公式製品ではありません。

## 1. 提供者とお問い合わせ

- 本アプリは [MIT License](../../LICENSE) の下で公開されるオープンソースソフトウェアです。
- プライバシーに関する問い合わせは、[GitHub Issues](https://github.com/JO3QMA/vrctweaker/issues) を利用してください（Issue 本文に VRChat の表示名・ユーザー ID 等を載せないでください）。

## 2. 収集の考え方

本アプリは **利用者専用の分析・広告用サーバーへ、利用データを自動送信する仕組みはコード上にありません**。データは主に **お使いの PC 上のファイルと SQLite データベース** に保存されます。インターネット通信は、VRChat 連携・サーバー状態表示・yt-dlp 更新など、**機能のためにコードから明示的に行われる通信**に限られます。

## 3. 端末に保存されるデータ

### 3.1 アプリデータフォルダ

Windows では `os.UserConfigDir()` 配下の `vrchat-tweaker`（通常は `%AppData%\Roaming\vrchat-tweaker\`）がアプリデータディレクトリです（`internal/platform/paths.AppDataDir`）。

| 保存先 | 内容 |
|--------|------|
| `vrchat-tweaker.db` | SQLite データベース（下表） |
| `tray.ico` | タスクトレイ用アイコン |
| `.vrchat-auth-token` | 資格情報ストアが使えない場合の **認証情報ファイル退避**（パーミッション 0600）。環境変数 `VRCHAT_TWEAKER_USE_FILE_CREDENTIALS=1` では常にこちら |

### 3.2 SQLite（`vrchat-tweaker.db`）に保存される主なデータ

スキーマは `internal/infrastructure/sqlite/db.go` に定義されています。

| テーブル / 設定 | 内容の概要 |
|-----------------|------------|
| `app_settings` | UI 言語、VRChat / Steam / output_log パス、ログ保持日数、ギャラリー自動補完、タスクトレイ終了、スリープ抑制、起動プロファイル ID、Activity ログ再開用チェックポイント（`activity_log_checkpoint`）、yt-dlp 関連設定、フレンド同期時刻（`vrchat_friends_synced_at`）、ステータス説明履歴（`presence_description_history`）など |
| `launch_profiles` | 起動プロファイル名と起動引数 |
| `screenshots` | スクリーンショット **ファイルパス**、撮影日時、ワールド ID、ファイルサイズ、作者の VRChat ユーザー ID（取り込み時に得られる場合） |
| `screenshot_thumbnails` | ギャラリー表示用の JPEG サムネイル（BLOB） |
| `screenshot_enrichment` | Activity ログと照合した **メタデータ補完結果**（インスタンス ID、参加者 JSON 等）。画像ファイルへの XMP 埋め込みを行う場合あり |
| `play_sessions` | output_log 由来のプレイセッション（開始・終了、インスタンス、参照したログファイルパス） |
| `user_encounters` | output_log 由来の **ユーザー遭遇記録**（ユーザー ID、表示名、インスタンス、ワールド、参加・退出時刻、ログファイルパス） |
| `friend_avatar_observations` | output_log 由来のフレンドアバター観測 |
| `video_playback_history` | output_log 由来の動画再生試行（URL、結果、解決 URL 等） |
| `users_cache` | VRChat API から取得したユーザー／フレンド情報のキャッシュ（表示名、ステータス、プロフィール画像 URL、バイオ等） |
| `world_info` | ワールド ID と表示名、最終訪問時刻 |
| `avatar_cache` | アバター ID と名前 |
| `automation_rules` / `automation_items` | オートメーション設定（条件・スクリプト等） |

スクリーンショット **画像本体** は VRChat が保存したファイルパスを参照します。ギャラリー「インデックスをクリア」は DB 上のレコードを削除し、**ディスク上の画像ファイルは削除しません**。

### 3.3 VRChat 認証情報

| 場所 | 内容 |
|------|------|
| Windows 資格情報マネージャー（優先） | サービス名 `vrchat-tweaker`、ユーザー名 `auth-token` に **ラップ済み認証 blob** またはレガシー形式のトークン（`internal/infrastructure/vrchatapi`） |
| 上記 `.vrchat-auth-token` | キーリング不可時の退避 |
| WebView の IndexedDB `vrctwk-credentials` | AES-GCM でトークンをラップする **非抽出可能な暗号鍵**（`frontend/src/services/credentialCrypto.ts`）。鍵は端末内のみ |

ログイン時に入力した **パスワードは VRChat API へのログイン要求時のみ送信**され、本アプリのデータベースや資格情報ストアに **平文では保存されません**。セッションは認証トークン（および HTTP Cookie `auth`）として扱われます。

### 3.4 VRChat クライアント側のファイル（読み書き）

| 場所（Windows の例） | 本アプリの扱い |
|----------------------|----------------|
| `%USERPROFILE%\AppData\LocalLow\VRChat\VRChat\` | 既定の `config.json`・`output_log*.txt` の場所。ログは **読み取り監視**（設定でパス変更可）。設定画面から `config.json` の削除操作が可能 |
| `%LocalAppData%\vrchat-tweaker\ytdlp\yt-dlp.exe` | 公式 yt-dlp バイナリのキャッシュ（`OfficialYTDLPCachePath`） |
| `%LocalLow%\VRChat\VRChat\Tools\yt-dlp.exe` | 上記キャッシュへのシンボリックリンク（メンテナンス機能 ON 時） |
| `%AppData%\yt-dlp\config` または `config.txt` | yt-dlp の Cookie 連携機能が書き込む **ローカル設定**（`--cookies-from-browser` または `--cookies` 行） |

Cookie 連携は **ブラウザや Cookie ファイルから yt-dlp が読み取る設定を書き込む**ものであり、Cookie の内容を本アプリのサーバーへ送る処理はありません（本アプリに専用サーバーはありません）。実際の Cookie 読み取りは **VRChat が yt-dlp を起動したとき**に yt-dlp が行います。

### 3.5 その他

- 二重起動防止用の Windows ミューテックス／イベント、およびローカルのアクティベーション用ソケット（`internal/infrastructure/singleinstance`）は、端末内のプロセス間制御のみです。

## 4. 外部への通信

コード上で確認できる通信先は次のとおりです。

### 4.1 VRChat

| 通信先 | 目的 |
|--------|------|
| `https://api.vrchat.cloud/api/1` | REST API（ログイン、現在ユーザー、フレンド一覧、ユーザー／アバター取得、ステータス更新、お気に入り等） |
| `wss://pipeline.vrchat.cloud/?authToken=…` | ログイン中の Pipeline WebSocket（フレンド関連イベント等） |
| API レスポンスに含まれる画像 URL（例: `api.vrchat.cloud` 配下） | UI でアバター等を表示する際、WebView が **GET で取得** |

VRChat 向け HTTP / WebSocket には User-Agent `VRChat Tweaker/<版>` と連絡先 URL（`https://github.com/JO3QMA/vrctweaker/issues`）が付与されます（`internal/appversion`）。

### 4.2 VRChat サービス状態

| 通信先 | 目的 |
|--------|------|
| `https://status.vrchat.com/api/v2/` | ダッシュボードのサーバー状態表示（`internal/infrastructure/statuspage`） |

### 4.3 GitHub（yt-dlp 更新）

| 通信先 | 目的 |
|--------|------|
| `https://api.github.com/repos/yt-dlp/yt-dlp/releases/latest` | 最新リリース情報の取得 |
| `https://github.com/…`、`https://objects.githubusercontent.com/…`、`https://release-assets.githubusercontent.com/…` | `yt-dlp.exe` のダウンロード（ホストはコードで検証） |

利用者が Video 画面で更新を実行した場合、または「Tools 内 yt-dlp のメンテナンス」が有効な場合に該当します。

### 4.4 動画サイト（間接）

本アプリ本体は動画 URL を **output_log から解析して DB に記録**します。動画ホストへの HTTP は、**VRChat クライアントが yt-dlp を実行したとき**に yt-dlp 側で発生します（本アプリが直接ストリーミングサイトへアクセスする常時機能はありません）。

### 4.5 利用者が明示的に開くリンク

設定・ライセンス画面等の **GitHub 上のドキュメントリンク**は、既定のブラウザで開きます（例: 本ポリシー、利用規約、Getting Started）。

## 5. 提供者に送信されないもの

次は **本アプリのコードから自動的には送信されません**。

- SQLite の遭遇ログ・フレンドキャッシュ・スクリーンショット索引の内容一式
- output_log の全文
- スクリーンショット画像ファイルの内容
- 認証トークンやラップ済み blob（VRChat API・Pipeline へのセッション利用を除く）
- 利用状況の分析・クラッシュレポート用の専用エンドポイント（実装なし）

不具合報告を GitHub Issues に **利用者自身が投稿**した場合は、その投稿内容が GitHub 上に保存されます。

## 6. データの削除・保持

| 操作 | 効果 |
|------|------|
| **ログアウト**（設定） | 資格情報ストアの削除、メモリ上のセッション解除、自己プロフィール行の削除等（`Logout`） |
| **Activity データの保存期間**（設定） | 指定日数より古い遭遇記録・プレイセッション・動画再生履歴・フレンドアバター観測の自動削除 |
| **遭遇ログをクリア** | `user_encounters` と `friend_avatar_observations` の全削除 |
| **スクリーンショットインデックスをクリア** | `screenshots` 関連レコードの削除（画像ファイルは残る） |
| **フレンドキャッシュをクリア** | `users_cache` の全削除とフレンド同期タイムスタンプのクリア |
| **DB 最適化 (VACUUM)** | SQLite の最適化のみ |
| **VRChat 資産キャッシュをクリア**（該当機能） | VRChat のアセットキャッシュディレクトリ内の削除 |
| **VRChat config.json の削除**（該当機能） | VRChat 側設定ファイルの削除 |
| **アプリのアンインストール** | 上記フォルダ（`%AppData%\Roaming\vrchat-tweaker\`、`%LocalAppData%\vrchat-tweaker\` 等）を利用者が手動削除すれば、当該データは端末から除去できます |

IndexedDB の暗号鍵は、WebView のサイトデータをブラウザ／アプリのデータ削除機能で消去できます。

## 7. 本ポリシーの変更

内容はリポジトリの `docs/legal/privacy-policy-ja.md` を更新して公開します。重要な変更がある場合はリリースノート等で告知するよう努めます。

## 8. 英語版

[Privacy Policy (English)](./privacy-policy-en.md)
