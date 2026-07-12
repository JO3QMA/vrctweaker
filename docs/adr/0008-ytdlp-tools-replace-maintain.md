# ADR 0008: yt-dlp Tools replace maintain

## Status

**Proposed**（grill-with-docs で製品方針を合意。[Issue #9](https://github.com/JO3QMA/vrctweaker/issues/9)。実機 PoC で [ADR 0007](0007-ytdlp-cookie-linkage.md) の Unblock 条件 2 を満たすまで**製品実装は保留**）

起動前ワンショット置換（[PR #40](https://github.com/JO3QMA/vrctweaker/pull/40)）は望む動作に未達。本 ADR は維持モードの Decision の正本。Cookie linkage は [ADR 0007](0007-ytdlp-cookie-linkage.md)。

## Context

- VRChat はログイン／ワールド参加時に Tools の `yt-dlp.exe` をハッシュ検証し、不一致なら **VRChat-bundled yt-dlp** へ戻す
- したがって起動**前**の一回きりの **yt-dlp Tools replace** だけでは維持できない。RO 固定は再生不能の報告がある
- 先行事例 [VRChat-YT-DLP-Fix](https://github.com/ShizCalev/VRChat-YT-DLP-Fix) は「起動後・同梱再生成を待ってから置換」で RO なしに公式 exe を載せている
- Issue #9 の目的は公式バイナリの適用しやすさ。Issue #8（Cookie）はその公式（Cookie 対応）exe を前提とするが、**出荷は #9 を先・#8 は別**とする

用語は [`CONTEXT.md`](../../CONTEXT.md) の **yt-dlp** セクションを正本とする。調査メモは ADR 0007 の Prior art / Validation plan を参照。

## Decision

1. **方式**: ログイン（または同梱再生成）**後**に Official yt-dlp cache から Tools へ **yt-dlp Tools replace** し、VRChat 稼働中は Tools を監視して巻き戻されたら再置換する（watcher）。stub exe／ローカルサーバ方式と RO 固定は採らない
2. **ライフサイクル**: Tweaker 常駐＋VRChat.exe 起動検知。VRChat が先に起動していても Tweaker 起動時にその場でアタッチする。監視は VRChat 終了で止めてよい
3. **yt-dlp Tools replace maintain**: オプトイン（既定オフ）。オンのときだけ起動検知・置換・監視を行う
4. **無効化**: 監視停止のみ。Tools 上のファイルは触らない。同梱版への復元は次の VRChat 起動に任せる
5. **Official yt-dlp cache**: 初回適用と明示の「更新を確認」で GitHub 公式 `yt-dlp.exe` を取得。以降のセッションはキャッシュから配置する（起動のたびに latest を取りに行かない）
6. **正本の分離**: desired＝maintain オン／オフ（アプリ設定）。**Tools replace effective state**＝Tools の exe が Official yt-dlp cache と一致するか。UI は両方を示す
7. **使用中競合**: 再置換時はファイル解放待ちリトライ。上限超過時は保留表示し、次の監視イベント／次セッションで再試行。プロセス強制終了はしない
8. **UI**: 専用の「動画」タブ（維持トグル・バージョン・更新・警告）。Settings / Config（`config.json`）には載せない。**Windows のみ**表示
9. **Tools replace risk acknowledgment**: 初回オン時のみ必須。以降は常時警告文
10. **出荷順**: 本機能（#9）を先に出す。**yt-dlp Cookie linkage**（#8 / ADR 0007）は別 PR。PR #40 はワンショット案として close し、DL／GitHub API／UI コードは再利用候補とする

## Out of scope

- Cookie linkage の実装（ADR 0007）
- PoC バイナリそのものの製品化（検証用 `cmd/` は別フェーズ）
- stub／キャッシュプロキシ方式（VRCVideoCacher 型）
- RO による巻き戻し防止
- 非 Windows

## Consequences

- #9 の製品スコープは「維持」であり、起動前ワンショット適用ボタン alone ではない
- #8 は本 ADR の方式が実機で Unblock 条件 2 を満たしてから再開する
- 動画タブと Settings の Cookie セクションは別画面・別概念のまま
