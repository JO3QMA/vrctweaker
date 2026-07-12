# ADR 0007: yt-dlp Cookie linkage

## Status

**Blocked**（実装保留。[Issue #8](https://github.com/JO3QMA/vrctweaker/issues/8)）

同梱 yt-dlp は Cookie オプション非対応。公式 exe の維持手段は [ADR 0008](0008-ytdlp-tools-replace-maintain.md)（Proposed）と実機 PoC に委譲。grill-with-docs で固めた Cookie config UI の設計（下記 Decision）は、前提が解けてから再開する。

関連: [Issue #9](https://github.com/JO3QMA/vrctweaker/issues/9)、[PR #40](https://github.com/JO3QMA/vrctweaker/pull/40)（起動前ワンショットは未達）。

## Context

- VRChat の動画プレイヤーは `%LocalAppData%Low\VRChat\VRChat\Tools\yt-dlp.exe` を使う
- Issue #8 は `%APPDATA%\yt-dlp\config`（yt-dlp user config）へ `--cookies-from-browser` / `--cookies` を書き、VRChat／EAC を改変せず制限付き動画を再生できると想定していた
- その前提は **Cookie 対応の公式 exe が Tools に載っていること**。維持の製品方針は ADR 0008。出荷順は **#9 先・#8 は別 PR**
- grill-with-docs では config の Managed 行 upsert、Settings UI、risk acknowledgment などを合意した（下記 Decision は設計メモとして残す）

用語は [`CONTEXT.md`](../../CONTEXT.md) の **yt-dlp** セクションを正本とする。

## Blocking findings

### A. 同梱ビルドに Cookie オプションが無い（Issue #8）

同梱 `yt-dlp.exe`（`--version` → `2026.06.09`）:

- `--help` に `--cookies` / `--cookies-from-browser` が無い（ドメイン allow-list など削った／独自オプションのみ）
- `--cookies-from-browser chrome -g <URL>` → `error: no such option: --cookies-from-browser`

**config に Cookie 参照を書いても、素の同梱バイナリでは効かない。**

### B. 公式 exe の維持が未検証（Issue #9）

起動前ワンショット置換（PR #40）は VRChat のハッシュ検証で同梱版へ戻され、RO 固定は再生不能の報告がある。維持の方式・製品 Decision は **ADR 0008**。本 ADR では「Cookie 実装の前提がまだ充足していない」ことだけを Blocking とする。

## Prior art findings (2026-07-12)

置換タイミングと先行事例の詳細は調査時点で ADR 0007 に記録し、製品方針は **ADR 0008** へ移した。要約:

- 巻き戻しの原因は置換タイミング（ログイン／ワールド参加時のハッシュ検証）
- [VRChat-YT-DLP-Fix](https://github.com/ShizCalev/VRChat-YT-DLP-Fix) は起動後・再生成待ち置換で RO なしに公式 exe を載せる（config への `--cookies-from-browser` upsert も行う — 本 Issue の user config 案の実例）
- [VRCVideoCacher](https://github.com/EllyVR/VRCVideoCacher) は stub＋サーバ方式（Tweaker の Cookie linkage 責務を超える）

## Validation plan（Unblock 条件 2 — Cookie 再開の前提）

リポジトリ内の単発 Go PoC（`cmd/`、Windows）で以下を実機確認する（製品コードではない）:

1. 置換後 `yt-dlp.exe --version` が公式のまま維持されるか（ログイン直後／ワールド移動を跨いで）
2. `--cookies-from-browser chrome -g <URL>` がエラーにならないか（**本 ADR / #8 の前提確認**）
3. **RO なしで動画再生が通るか**
4. 巻き戻しのタイミングと回数（製品の監視要否の材料）

PoC は yt-dlp user config を書かない。結果は Issue #9 と ADR 0008 に記録し、成功時に本 ADR の Blocked 解除を検討する。

## Decision（実装時の設計メモ — 現状は未採用）

前提が満たされた場合に限り、以前の合意を再開候補とする:

1. **責務**: Tweaker は **yt-dlp Cookie linkage** として yt-dlp user config の **Managed cookie options** の書き込み／削除だけを行う。Cookie 本体の取得・検証、cookies ファイルの作成、yt-dlp／動画再生の実行・成否確認は行わない
2. **ファイル操作**: 有効化は Managed 行の **upsert**。無効化は Managed 行の **削除のみ**（ファイル全体のリネーム退避・丸ごと置換はしない）。他行は常に残す。親ディレクトリが無ければ作成する。無効化後に他行が無く空ならファイルを削除する
3. **方式**: **Browser cookie source** と **Cookies file source** を v1 で両方提供し、有効時は **排他**。Browser の v1 選択肢は `chrome` / `edge` / `firefox` の既定プロファイルのみ。Cookies file は書き込み前に **ファイル存在**を必須（形式パースはしない）
4. **正本**: **Cookie linkage effective state** は yt-dlp user config が正。**Cookie linkage draft**（方式・ブラウザ・パスの下書きと **Cookie linkage risk acknowledgment**）はアプリ側。有効時の変更は **即時書き込み**。書き込み失敗時はエラー表示、UI を操作前の Effective に戻し、Draft の試行値は残す。config 未作成の読み取りはエラーにせず Effective＝無効
5. **UI**: Settings の独立セクション（Config＝VRChat `config.json` 画面には載せない。動画タブ＝Tools replace maintain にも載せない）。**初回有効化のみ** risk acknowledgment 必須。以降は常時警告文。v1 は **Windows のみ**表示（他 OS はセクション非表示）
6. **ログ**: ブラウザ名と cookies ファイルパスは可。Cookie 中身は禁止

## Out of scope（当面）

- 上記 Decision（#8）の実装
- Tools replace maintain の製品実装（ADR 0008）
- PR #40 相当のワンショット置換のマージを「完了」とみなすこと
- RO による戻し防止を製品機能にすること
- Cookie ファイルの作成・エクスポート
- 同梱ビルドへのパッチ／再配布

## Unblock 条件（いずれか）

1. 同梱 `yt-dlp.exe` が `--cookies` または `--cookies-from-browser` を受け入れるようになる（`--help` / 実コマンドで確認）
2. **RO なしで**公式 exe を起動後も維持でき、かつ動画再生が通る手順が再現できる（ADR 0008 の方式／PoC）。そのうえで Cookie UI を再 grill。ワンショット置換 alone や RO 固定は Unblock に含めない
3. Issue #8 を閉じる、または前提が違う別 Issue に切り替える

## Consequences

- **#8 はいま出荷しない**。#9 の製品方針は ADR 0008（Proposed）にあり、PoC 前は実装しない
- #8 は公式（または Cookie 対応）バイナリ前提。その維持が実機で確認できてから Decision を再開する
- grill で固めた Cookie 用語・設計メモは、Unblock 後の再開用として残す
