# ADR 0008: yt-dlp Tools replace maintain

## Status

**Accepted**（grill-with-docs で製品方針を合意。[Issue #9](https://github.com/JO3QMA/vrctweaker/issues/9) は実装完了・closed。実機 PoC で Official cache＋Tools symlink により RO なし動画再生を確認）

起動前ワンショットの直置き（[PR #40](https://github.com/JO3QMA/vrctweaker/pull/40)）は望む動作に未達。本 ADR は維持モードの Decision の正本。Cookie linkage は [ADR 0007](0007-ytdlp-cookie-linkage.md)。

## Context

- VRChat はログイン／ワールド参加時に Tools の `yt-dlp.exe` をハッシュ検証し、不一致なら **VRChat-bundled yt-dlp** へ戻す
- したがって起動**前**の一回きりの **yt-dlp Tools replace** だけでは維持できない。RO 固定は再生不能の報告がある
- Official onefile を Tools（LocalLow）に**直置き**すると PyInstaller が一時ディレクトリを作れず再生不能。Official をアプリ Local 側に置き、Tools から **symlink** する方式で RO なし再生が通る（PoC / 製品）
- 先行事例 [VRChat-YT-DLP-Fix](https://github.com/ShizCalev/VRChat-YT-DLP-Fix) は起動後・同梱再生成待ちの置換で RO なしに公式 exe を載せている（配置手段は本 ADR の symlink とは異なるが、タイミングの教訓は共有）
- Issue #9 の目的は公式バイナリの適用しやすさ。Issue #8（Cookie）はその公式（Cookie 対応）exe を前提とするが、**出荷は #9 を先・#8 は別**とする

用語は [`CONTEXT.md`](../../CONTEXT.md) の **yt-dlp** セクションを正本とする。調査メモは ADR 0007 の Prior art / Validation plan を参照。

## Decision

1. **方式**: Official yt-dlp cache を実体とし、Tools の `yt-dlp.exe` はそれへの **symlink**（**yt-dlp Tools replace**）。VRChat 稼働中は Tools を監視し、巻き戻されたら再リンクする（watcher）。Tools への公式バイナリ直置きコピー、stub exe／ローカルサーバ方式、RO 固定は採らない
2. **ライフサイクル**: Tweaker 常駐＋VRChat.exe 起動検知。VRChat が先に起動していても Tweaker 起動時にその場でアタッチする。監視は VRChat 終了で止めてよい
3. **yt-dlp Tools replace maintain**: オプトイン（既定オフ）。オンのときだけ起動検知・置換・監視を行う
4. **無効化**: 監視停止のみ。Tools 上のファイルは触らない。同梱版への復元は次の VRChat 起動に任せる
5. **Official yt-dlp cache**: 初回適用と明示の「更新を確認」で GitHub 公式 `yt-dlp.exe` を取得。以降のセッションはキャッシュから配置する（起動のたびに latest を取りに行かない）
6. **正本の分離**: desired＝maintain オン／オフ（アプリ設定）。**Tools replace effective state**＝Tools が Official yt-dlp cache を指しているか。UI は両方を示す
7. **使用中競合**: 再リンク時はファイル解放待ちリトライ。上限超過時は保留表示し、次の監視イベント／次セッションで再試行。プロセス強制終了はしない。symlink 作成に Developer Mode または昇格が必要な場合はユーザー向けに案内する
8. **UI**: 専用の「動画」タブ（維持トグル・バージョン・更新・警告）。Settings / Config（`config.json`）には載せない。**Windows のみ**表示
9. **Tools replace risk acknowledgment**: 初回オン時のみ必須。以降は常時警告文
10. **出荷順**: 本機能（#9）を先に出す。**yt-dlp Cookie linkage**（#8 / ADR 0007）は別 PR。PR #40 はワンショット案として close し、DL／GitHub API／UI コードは再利用候補とする

## Out of scope

- Cookie linkage の実装（ADR 0007）
- PoC バイナリそのものの製品化（検証用 `cmd/` は別フェーズ）
- stub／キャッシュプロキシ方式（VRCVideoCacher 型）
- RO による巻き戻し防止
- Tools への公式バイナリ直置きコピー
- 非 Windows

## Consequences

- #9 の製品スコープは「維持（symlink＋watcher）」であり、起動前ワンショット適用ボタン alone ではない
- #8（Cookie linkage）は ADR 0007 に従い製品実装済み。制限付き再生には本 ADR の Official cache 経由が前提
- 動画タブと Settings の Cookie セクションは別画面・別概念のまま
