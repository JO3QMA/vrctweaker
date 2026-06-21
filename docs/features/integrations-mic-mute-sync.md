# 機能: Mic Mute Sync（VRChat ↔ Discord マイクミュート連動）

## 概要

VRChat と Discord の **Mic Mute**（マイクミュートのみ。Discord デフンは含まない）を双方向に揃える。ユーザーが Settings で明示的に ON にしたときだけ動作する（デフォルト OFF）。

設計判断の背景は `docs/adr/0001-mic-mute-sync-vrchat-osc-discord-rpc.md`。ドメイン用語はルート `CONTEXT.md` を参照。

## ゴール

- VRChat でマイクをミュート／解除すると Discord の Mic Mute が追従する
- Discord でマイクをミュート／解除すると VRChat が追従する
- ループは **Echo Suppression** で防止する
- Settings の **Sync Status** で両側の状態と前提条件を確認できる（デスクトップ通知は使わない）

## 非ゴール（初版）

- Discord デフンの連動
- Toggle Voice OFF（プッシュトゥミュート）のサポート
- Linux / macOS 向け提供（**Mic Mute Sync Availability** は Windows のみ）
- Automation ルールとしての露出（**Mic Mute Sync Settings** は独立機能）

## 仕様

### 同期の前提

| 条件 | 未充足時の挙動 |
|------|----------------|
| Mic Mute Sync が ON | 同期しない（設定 OFF がデフォルト） |
| VRChat 起動かつ OSC 有効 | **Sync Pause**。Sync Status に理由表示 |
| Discord デスクトップ起動かつ Voice RPC 認可済み | 同上 |
| VRChat Toggle Voice が ON | 同上 |
| Windows 上で Tweaker 実行 | 他 OS では設定 UI を無効化または非表示 |

### Session Baseline と Sync Pause

- **セッション開始**（ユーザーが ON にし、前提がすべて揃った瞬間）: VRChat の Mic Mute を正として Discord を揃える
- **切断検知**（VRChat 終了、OSC 不通、Discord 切断など）: **Sync Pause**（状態は変更しない）
- **再開**（前提が再びすべて揃った瞬間）: Session Baseline を再適用してから監視を再開

### Echo Suppression

- Tweaker が一方へ送った Mic Mute 変更に起因する、反対側からの状態変化は短時間無視する
- 具体的な抑制ウィンドウ（例: 500ms〜2s）は実装時に調整。単体テストで「ループしない」ことを検証する

### VRChat OSC

- **読み取り**: `/avatar/parameters/MuteSelf`（Bool）
- **書き込み**: Toggle Voice ON 前提で `/input/Voice` に `1` → `0` のパルス（現在値と目標値が異なるときのみトグル）
- **OSC Endpoint**: Mic Mute Sync Settings で `inPort:outIP:outPort` を設定。未設定時は `9000:127.0.0.1:9001`
- **ランチャー連携**: Tweaker ランチャー経由起動時は `--osc=` を自動付与して OSC 有効化を支援（Steam 直起動時はユーザーが VRChat 側で OSC を有効にする必要あり）

### Discord Voice RPC

- **読み取り**: Voice settings の `mute` フィールド
- **書き込み**: `SET_VOICE_SETTINGS` で `mute` を設定
- **Discord Application**: プロジェクト同梱の Client ID。ユーザーごとの Developer Portal 登録は不要

### Sync Status（Settings UI）

チェックリスト形式で表示する。

- Mic Mute Sync: ON / OFF
- プラットフォーム: Windows 対応 / 非対応
- VRChat OSC: 接続済 / 未接続 + **OSC Endpoint** 表示
- Toggle Voice: 検知不能な場合は「VRChat 設定を確認」と案内
- Discord RPC: 接続済 / 未接続 / 未認可
- VRChat Mic Mute: ミュート / オン
- Discord Mic Mute: ミュート / オン
- 同期エンジン: 同期中 / Sync Pause（理由）

## 実装フェーズ（縦スライス）

### フェーズ ① — VRChat OSC 読み取り + Sync Status（片側）

**スコープ**

- OSC クライアント（受信）と MuteSelf 購読
- Settings に Mic Mute Sync Settings セクション（ON/OFF、OSC Endpoint）
- Sync Status の VRChat 関連項目

**受け入れ条件**

- VRChat 起動・OSC 有効時、Settings に MuteSelf の現在値が反映される
- OSC 未接続時、Sync Status に未接続と表示される
- Linux ビルドでは当該設定が無効または非表示

### フェーズ ② — Discord Voice RPC 接続 + 状態読み取り

**スコープ**

- Discord IPC / Voice RPC 接続と認可フロー
- Sync Status の Discord 関連項目

**受け入れ条件**

- Discord 起動時、認可後に Discord の `mute` 状態が Sync Status に表示される
- 認可拒否・Discord 未起動時、Sync Status に理由が表示される

### フェーズ ③ — VRChat → Discord 一方向 + Session Baseline

**スコープ**

- 同期エンジン（ON 時のみ稼働）
- Session Baseline（VRChat → Discord）
- Sync Pause（切断時停止、再開時 Baseline）
- Discord への `SET_VOICE_SETTINGS`
- ランチャー起動時の `--osc=` 自動付与

**受け入れ条件**

- Mic Mute Sync ON かつ前提充足時、VRChat でミュート操作すると Discord が同じ Mic Mute 状態になる
- セッション開始時、VRChat がミュートなら Discord もミュートになる
- 片方切断中は Discord / VRChat の状態を変更しない

### フェーズ ④ — 双方向 + Echo Suppression

**スコープ**

- Discord の mute 変化を購読し VRChat へ反映（OSC パルス）
- Echo Suppression
- VRChat 書き込み時の「目標状態と MuteSelf が一致しているか」判定

**受け入れ条件**

- Discord でミュート操作すると VRChat が追従する
- VRChat → Discord → VRChat のループでチラつかない（Echo Suppression）
- Toggle Voice OFF 時は同期せず Sync Status に案内が出る

## 実装方針（レイヤー）

```
internal/domain/micmutesync/     … 同期ポリシー（Baseline, Pause, Echo Suppression の純粋ロジック）
internal/infrastructure/vrchatosc/ … OSC 送受信
internal/infrastructure/discordrpc/ … Voice RPC
internal/usecase/mic_mute_sync_usecase.go … セッション生命周期・設定永続化
app.go                         … 起動時の常駐ワーカー配線
frontend SettingsView          … Mic Mute Sync Settings + Sync Status
```

- Automation の `EventBus` / ルールエンジンとは別経路で常駐する
- 設定は既存の settings 永続化（SQLite）にキーを追加

## Wails バインディング（案）

- `GetMicMuteSyncSettings()` / `SaveMicMuteSyncSettings()`
- `GetMicMuteSyncStatus()` — Sync Status 用 DTO（ポーリングまたはイベント推送）

## 依存

- 既存ランチャーの OSC 引数（`internal/domain/launcher/launch_args.go`）
- Settings 画面（`frontend/src/views/SettingsView.vue`）
- Windows ビルドタグで Discord RPC / OSC ワーカーをガード

## テスト観点

- Unit: Session Baseline / Sync Pause / Echo Suppression の状態機械
- Unit: OSC パルス生成（Toggle Voice ON、目標状態との差分のみトグル）
- Unit: Discord mute 変化 → VRChat 操作のマッピング（モック）
- Integration（可能なら）: OSC の送受信モック
- Frontend: Sync Status 表示（Vitest）、Settings の ON/OFF（E2E はフェーズ ④ 以降を推奨）

## 参考リンク

- [VRChat OSC Input](https://docs.vrchat.com/docs/osc-as-input-controller) — `/input/Voice`
- [VRChat Animator Parameters](https://creators.vrchat.com/avatars/animator-parameters/) — `MuteSelf`
- [Discord RPC Voice Settings](https://github.com/discord/discord-api-docs/blob/main/developers/topics/rpc.mdx)
