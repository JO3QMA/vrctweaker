# VRChat Tweaker — はじめに

**VRChat Tweaker** は、VRChat 周辺のログ・スクリーンショット・フレンド情報などをまとめて扱う **Windows 向け** デスクトップアプリです。

> **非公式ツール:** 本アプリは VRChat 公式の製品・サービスではなく、VRChat 株式会社とは無関係です。VRChat の利用規約・Community Guidelines を遵守したうえでご利用ください。

## 対応環境

| 項目 | 要件 |
|------|------|
| OS | **Windows 10 / 11（64 ビット）** |
| VRChat | Steam 版または VRChat クライアント（通常の PC 版） |
| ネットワーク | VRChat API 利用のためインターネット接続 |
| アカウント | VRChat アカウント（ログイン機能を使う場合） |

Linux / macOS は **未サポート** です（将来の検討対象）。

## 入手方法

初回頒布以降、次のいずれかから **Windows 用 zip** を入手します（内容は同一を想定）。

- **BOOTH** — 商品ページのダウンロード
- **[GitHub Releases](https://github.com/JO3QMA/vrctweaker/releases)** — タグ付きリリースの Assets

zip には実行ファイル（例: `vrchat-tweaker.exe`）と同梱ファイルが含まれます。詳細は各リリースの説明を参照してください。

## インストール（zip）

1. zip を任意のフォルダに展開します（例: `C:\Apps\VRChat-Tweaker\`）。
2. **インストーラは不要** です。展開したフォルダ内の `vrchat-tweaker.exe` を起動します。
3. 初回起動時、Windows の SmartScreen 等で警告が出る場合があります。信頼できる入手元（BOOTH / 公式 GitHub Release）から取得した場合のみ、表示に従って実行を許可してください。
4. ショートカットが必要なら、`.exe` を右クリック → **ショートカットの作成**、またはタスクバーへピン留めしてください。

アップデート時は、新しい zip でフォルダを上書きするか、別フォルダに展開して設定を引き継ぐ運用を推奨します（自動更新機能は初回版ではありません）。

## 初回起動

1. アプリを起動すると **Dashboard** が表示されます。
2. 左メニューから **Settings（設定）** を開き、UI 言語を選べます（日本語 / English / 한국어 / 简体中文 / 繁體中文）。
3. **VRChat のパス**（`launch.exe` 推奨）と **output_log** の監視フォルダが未設定の場合、Settings のパス設定で指定するか、既定の検出結果を確認してください。
4. フレンド一覧や API 連携を使う場合は、Settings の **VRChat ログイン** でアカウント情報を入力します（2FA 有効時は認証コードも必要）。

## VRChat ログインについて

- ログインは VRChat の公式 API 経由です。パスワードはアプリ内フォームから送信され、**セッション用トークン** として保存されます。
- トークンは **Windows の資格情報ストア（Credential Manager）** を優先して利用します。環境によってはアプリデータフォルダ内の保護ファイルに退避する場合があります。
- **ログアウト** または資格情報の削除で、保存されたトークンは破棄されます。
- 本アプリは VRChat 公式クライアントの代替ではありません。ゲーム内操作は VRChat 本体で行ってください。

## データの保存場所（Windows）

| 種類 | おおよその場所 |
|------|----------------|
| アプリ本体の DB・設定 | `%AppData%\Roaming\vrchat-tweaker\`（例: `vrchat-tweaker.db`） |
| VRChat 側 config / ログ | `%USERPROFILE%\AppData\LocalLow\VRChat\VRChat\`（`config.json`、`output_log*.txt` 等） |
| スクリーンショット本体 | VRChat が保存した画像ファイル（Gallery はインデックス・メタデータをアプリ DB に保持） |

パスは環境や設定で異なります。Settings 画面の **パス設定** で上書きできます。

## 基本的なトラブルシュート

| 症状 | 確認すること |
|------|----------------|
| 起動しない / すぐ終了 | zip を再ダウンロード、ウイルス対策ソフトの除外、Visual C++ ランタイム（Windows 更新） |
| ログインできない | ユーザー名・パスワード・2FA、VRChat 側のメンテナンス、[VRChat Status](https://status.vrchat.com/) |
| ギャラリーが空 | VRChat のスクショ保存先、Settings のパス、DB メンテナンスでインデックスを消していないか |
| Activity が更新されない | `output_log` フォルダのパス、VRChat 起動中か、ログ保持日数（Settings） |
| 二重起動のメッセージ | 既に Tweaker が起動中。タスクトレイのアイコンを確認 |

ログや Issue 報告時は、**VRChat の表示名・ユーザー ID（`usr_…`）・プロフィール URL** を公開場所に載せないでください。

## サポート・フィードバック

不具合報告・機能要望は GitHub Issues から受け付けています。

**[GitHub Issues（JO3QMA/vrctweaker）](https://github.com/JO3QMA/vrctweaker/issues)**

Bug report / Feature request テンプレートに沿って、再現手順・期待動作・環境（Windows 版、Tweaker のバージョン）を記載してください。

## ライセンス

本ソフトウェアは [MIT License](../../LICENSE) の下で公開されています。同梱の **OSS ライセンス** はアプリ内 Settings → OSS ライセンス一覧からも確認できます。
