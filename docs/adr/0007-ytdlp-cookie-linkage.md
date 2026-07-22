# ADR 0007: yt-dlp Cookie linkage

## Status

**Accepted**（grill-with-docs 再合意済み。[Issue #8](https://github.com/JO3QMA/vrctweaker/issues/8)。製品実装完了: [PR #214](https://github.com/JO3QMA/vrctweaker/pull/214), [PR #215](https://github.com/JO3QMA/vrctweaker/pull/215), [PR #218](https://github.com/JO3QMA/vrctweaker/pull/218)）

- **前提**: Cookie オプションは **VRChat-bundled yt-dlp** には無い。制限付き再生の目的を満たすには [ADR 0008](0008-ytdlp-tools-replace-maintain.md) の **Official yt-dlp cache**（Tools symlink 経由）が必要。Cookie linkage は maintain にハード依存しないが、**Cookie linkage official hint** で案内する
- **出荷ゲート（Unblock 条件 2）**: Official yt-dlp（GitHub 公式リリース）が `--cookies` / `--cookies-from-browser` を受け入れること — 下記 Validation results で充足

関連: [Issue #9](https://github.com/JO3QMA/vrctweaker/issues/9)（closed）、[PR #40](https://github.com/JO3QMA/vrctweaker/pull/40)（起動前ワンショットは未達）。

## Context

- VRChat の動画プレイヤーは `%LocalAppData%Low\VRChat\VRChat\Tools\yt-dlp.exe` を使う
- Issue #8 は `%APPDATA%\yt-dlp\config`（yt-dlp user config）へ `--cookies-from-browser` / `--cookies` を書き、VRChat／EAC を改変せず制限付き動画を再生できると想定していた
- その前提は **Cookie 対応の公式 exe が Tools から参照されていること**（ADR 0008: Official cache＋symlink）。出荷順は **#9 先・#8 は別 PR**（#9 は完了）
- grill-with-docs で config の Managed 行 upsert、動画タブ UI、risk acknowledgment、maintain とのソフト結合などを合意した（下記 Decision が正本。UI 配置は当初 Settings 案だったが動画タブへ移した）

用語は [`CONTEXT.md`](../../CONTEXT.md) の **yt-dlp** セクションを正本とする。

## Historical blocking findings

### A. 同梱ビルドに Cookie オプションが無い（継続）

同梱 `yt-dlp.exe`（調査時 `--version` → `2026.06.09`）:

- `--help` に `--cookies` / `--cookies-from-browser` が無い（ドメイン allow-list など削った／独自オプションのみ）
- `--cookies-from-browser chrome -g <URL>` → `error: no such option: --cookies-from-browser`

**config に Cookie 参照を書いても、素の同梱バイナリでは効かない。** 本機能の出荷は同梱版の対応を待たず、Official cache 経由で成立する。

### B. 公式 exe の維持（Issue #9）— 充足済み

起動前ワンショット直置き（PR #40）は未達だったが、Official cache＋Tools symlink＋watcher（ADR 0008 Accepted）で RO なし再生を確認し #9 は実装・closed。

## Prior art findings (2026-07-12)

置換タイミングと先行事例の詳細は調査時点で ADR 0007 に記録し、製品方針は **ADR 0008** へ移した。要約:

- 巻き戻しの原因は置換タイミング（ログイン／ワールド参加時のハッシュ検証）
- [VRChat-YT-DLP-Fix](https://github.com/ShizCalev/VRChat-YT-DLP-Fix) は起動後・再生成待ち置換で RO なしに公式 exe を載せる（config への `--cookies-from-browser` upsert も行う — 本 Issue の user config 案の実例）
- [VRCVideoCacher](https://github.com/EllyVR/VRCVideoCacher) は stub＋サーバ方式（Tweaker の Cookie linkage 責務を超える）

## Validation results（出荷ゲート — 2026-07-18 充足）

維持・再生は ADR 0008 / #9 で充足済み。Cookie オプション受理について:

1. **Official yt-dlp cache**（GitHub `releases/latest` の `yt-dlp.exe`）は upstream 標準ビルドであり、`--cookies` / `--cookies-from-browser` を受け付ける（Tweaker の DL 元と同一系列）。同梱版との差は Blocking finding A のとおり
2. Tweaker は yt-dlp を**実行して** Cookie オプションを検証しない（Decision の責務外）。config 書き込みと Effective 表示のみ
3. **VRChat 内の制限付き動画の実再生**は CI 対象外。ユーザーが maintain 有効＋Cookie linkage 有効のうえで手動確認する（PoC 手順は Test plan の Manual）

Blocked (ship) は解除。Decision に従い PR #214 / #215 / #218 で実装済み。

## Decision（実装正本）

1. **責務**: Tweaker は **yt-dlp Cookie linkage** として yt-dlp user config の **Managed cookie options**（`--cookies-from-browser` / `--cookies` のみ）の書き込み／削除だけを行う。Cookie 本体の取得・検証、cookies ファイルの作成、yt-dlp／動画再生の実行・成否確認、sleep／リクエスト間隔などの他オプションの挿入は行わない
2. **ファイル操作**: 有効化は Managed 行の **upsert**。無効化は Managed 行の **削除のみ**（ファイル全体のリネーム退避・丸ごと置換はしない）。他行は常に残す。親ディレクトリが無ければ作成する。無効化後に他行が無く空ならファイルを削除する。パス解決: **`%APPDATA%\yt-dlp\config` を正本**とし、それが無く `config.txt` だけあればそちらを読み書き対象にする。新規作成は常に `config`。`config` と `config.txt` が両方あるときは **`config` のみ**を Managed 対象とする（`config.txt` は触らない）。根拠: yt-dlp User config は候補を順に試し **最初に存在する 1 ファイルだけ**を読む（`options.py` の `next(filter(None, ...))`）。`%APPDATA%/yt-dlp` 内の順は `config` → `config.txt` なので、両方あるときも runtime は `config` のみを読む
3. **方式**: **Browser cookie source** と **Cookies file source** を v1 で両方提供し、有効時は **排他**。Browser の v1 選択肢は `chrome` / `edge` / `firefox` の既定プロファイルのみ。Cookies file は **テキスト＋参照ボタン**（既存 `OpenFileDialog`）でパス指定し、書き込み前に **ファイル存在**を必須（空ファイルも可。形式パース・非空チェックはしない）。**Cookie linkage unsupported form**（コンテナ・未対応ブラウザ・両オプション併記など）は Effective＝有効＋未対応表示。v1 保存で全 Managed 行を選んだ一方へ置換、無効化で Cookie 参照行をすべて削除
4. **正本**: **Cookie linkage effective state** は yt-dlp user config が正。**Cookie linkage draft**（方式・ブラウザ・パスの下書きと **Cookie linkage risk acknowledgment**）はアプリ側。有効時の変更は **即時書き込み**。書き込み失敗時はエラー表示、UI を操作前の Effective に戻し、Draft の試行値は残す。config **未作成**の読み取りはエラーにせず Effective＝無効。**Cookie linkage config read failure**（存在するが読めない）は UI エラーとし、Effective を無効と偽らず、書き込みは止める。**risk acknowledgment** は Tweaker が初めて config へ書き込む操作の前に必須（他ツール由来で Effective が既に有効でも、閲覧のみなら不要）。ack 後の書き込みでは再確認しない
5. **UI**: 動画タブの独立セクション（Tools replace maintain の下。Config＝VRChat `config.json` 画面や Settings には載せない）。初回書き込み前に risk acknowledgment。以降は常時警告文。v1 は **Windows のみ**表示（他 OS はセクション非表示）。**Tools replace effective state** が偽のときは **Cookie linkage official hint**（同一画面の Tools replace 操作への案内）を出すが、有効化・書き込みは止めない（ソフト結合）。ブラウザロック時の失敗は自動検知せず、ツールチップ等で「ブラウザを閉じる／Cookies file source を使う」を案内するにとどめる
6. **ログ**: ブラウザ名と cookies ファイルパスは可。Cookie 中身は禁止。config 読み取り失敗の詳細はログ可（Cookie 中身を含まない範囲）
7. **Wails 契約**: Get は `(DTO, error)`。ファイル無し／Managed 無しは DTO＝無効・`error=nil`。**Cookie linkage config read failure** は `error≠nil`（無効 DTO で偽らない）。書き込み失敗・cookies ファイル不存在は `error`。risk ack 未了は maintain と同様の専用 sentinel。フロントは Video 用 classifier で英語メッセージを i18n キーへ（詳細をキーだけで捨てない）
8. **FS 書き込み**: Managed upsert／削除は **同一ディレクトリの一時ファイル → rename 置換**。失敗時は元ファイルを残す。常時 `.bak` 保持はしない

## Implementation

| 層 | 場所 |
|----|------|
| Usecase | `internal/usecase/ytdlp_cookie_linkage.go` |
| Wails | `app.go`（`GetYTDLPCookieLinkageStatus` ほか） |
| UI | `frontend/src/components/CookieLinkageSection.vue`（動画タブ） |
| i18n | `frontend/src/i18n/locales/*.json` の `video.cookieLinkage.*` |

PR 分割: (1) Go usecase＋単体テスト (#214) (2) Wails＋UI＋i18n (#215) (3) Settings から動画タブへ UI 移動 (#218)

## Out of scope（v1）

- Cookie 有効化の maintain／Effective ハードゲート
- sleep／リクエスト間隔など Cookie 以外の config 自動挿入
- PR #40 相当のワンショット直置きの復活
- RO による戻し防止を製品機能にすること
- Cookie ファイルの作成・エクスポート
- 同梱ビルドへのパッチ／再配布
- ブラウザコンテナ／非既定プロファイル指定
- yt-dlp 実行による Cookie オプション検証・制限付き URL の自動試行

## Consequences

- **#8 は製品実装済み**（Issue クローズ可）。#9（ADR 0008）は完了済み
- 同梱 yt-dlp のまま Cookie linkage だけ有効にしても制限付き再生の目的は満たせない（official hint を参照）
- 用語は `CONTEXT.md` を正本とする

## Test plan

### Go 単体（PR #214 — 正本）

| 入力 | 期待 | テスト名 |
|------|------|----------|
| config 無し | Effective＝無効、error=nil | `TestGet_noFile` |
| config のみ Managed 無し | 無効、他行保持 | `TestGet_noManagedLines` |
| `config.txt` のみ | そちらを対象 | `TestResolve_configTxtOnly` |
| `config` と `config.txt` 両方 | `config` のみ触る | `TestResolve_prefersConfig` |
| Browser 有効化 | upsert 一方、他行残す | `TestUpsert_browserSource` |
| Cookies file・パス無し | error | `TestUpsert_cookiesFileMissing` |
| Cookies file・空ファイル | 可 | `TestUpsert_cookiesFileEmptyOK` |
| 無効化 | Managed のみ削除、空ならファイル削除 | `TestDisable_removesManagedOnly` |
| unsupported 形 → v1 保存 | 全 Managed を一方に置換 | `TestUpsert_replacesUnsupported` |
| パスがディレクトリ等で読めない | error（無効と偽らない） | `TestGet_readFailure` |
| rename 失敗 | 元ファイル残存 | `TestWrite_keepsOriginalOnFailure` |
| risk ack 未了で書き込み | sentinel | `TestWrite_requiresRiskAck` |

（行の増減はこの表を更新して正本を保つ。）

### Manual（Windows — 任意）

1. 動画タブで Tools replace maintain を有効化し、Tools が Official cache を指すことを確認
2. Official cache の `yt-dlp.exe --help` に `--cookies-from-browser` があること（または `… --cookies-from-browser chrome -g <公開 URL>` が `no such option` にならないこと）
3. Cookie linkage を有効化し、`%APPDATA%\yt-dlp\config` に Managed 行が書き込まれること
4. （任意）VRChat 内で制限付き動画の再生を確認
