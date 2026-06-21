# Mic Mute Sync: VRChat OSC + Discord Voice RPC

Mic Mute Sync は VRChat 側を VRChat OSC（`MuteSelf` の読み取り、`/input/Voice` の制御）、Discord 側を Discord デスクトップクライアントへの Voice RPC（`GET_VOICE_SETTINGS` / `SET_VOICE_SETTINGS`）で実装する。キーバインド送信やログ解析は採用しない。

双方向かつ状態ベースの同期には、両プラットフォームで「現在ミュートか」を読み、必要なら絶対状態へ揃えられる API が必要である。Discord ではキーバインド送信だけではユーザーが UI で操作した変化を検知できない。VRChat ではマイクミュート状態を安定して読む公式手段が OSC の `MuteSelf` のみであり、ログからの代替は信頼できない。Discord Voice RPC は音声設定の排他ロックがあるが、Mic Mute Sync Session 中は Tweaker が一時的に音声設定を握る前提で許容する。Client ID は VRChat Tweaker プロジェクトが Developer Portal に登録したアプリを同梱する。

## Considered Options

- **Discord キーバインド送信（`SendInput`）**: 認可不要だが状態検知が弱く、ユーザーのキー設定・フォーカスに依存する。双方向には不向き。
- **Discord RPC 検知 + キーバインド制御**: 複雑さに対しメリットが小さい。
- **VRChat ログ解析**: マイクミュートの安定した検知経路がない。

## Consequences

- 初版は Windows のみ（Discord デスクトップ + VRChat PC 想定）。
- VRChat で OSC が無効、または Toggle Voice が OFF の間は同期しない。
- Discord RPC 接続には初回認可が必要。同梱 Client ID の登録・維持がプロジェクトの運用責務になる。
- 用語はルートの `CONTEXT.md` を参照する。
