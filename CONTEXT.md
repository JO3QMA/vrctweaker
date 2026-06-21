# VRChat Tweaker

VRChat 周辺の起動・監視・自動化を行うデスクトップアプリのドメイン用語集。

## Integrations

**Mic Mute Sync**:
VRChat と Discord のマイクミュート状態を双方向に揃える機能。
_Avoid_: ミュート連動, ボイス同期

**Echo Suppression**:
Mic Mute Sync で Tweaker が一方へ送った変更に起因する、反対側からの状態変化を一時的に無視する仕組み。双方向ループの防止に使う。
_Avoid_: デバウンス（単独では意味が広すぎる）

**Mic Mute**:
マイク入力だけをオフにする状態。Discord のデフン（スピーカーミュート）は含まない。
_Avoid_: ミュート全般, ボイスオフ

**Discord Voice RPC**:
Discord デスクトップクライアントとの IPC 経由で、Mic Mute Sync がマイクミュート状態を読み書きする連携方式。
_Avoid_: Discord API, Bot API

**VRChat OSC**:
Mic Mute Sync が VRChat のマイク状態を読み書きする唯一の連携方式。OSC が無効な間は同期を行わない。
_Avoid_: OSC連携, VRChat API

**Mic Mute Sync Session**:
ユーザーが設定で Mic Mute Sync を ON にし、VRChat（OSC 有効）と Discord の両方が利用可能な間、同期が動作している状態。
_Avoid_: 同期モード, 連動セッション

**Session Baseline**:
Mic Mute Sync Session の開始時（および再開時）に、VRChat の Mic Mute 状態を正として Discord を揃えること。
_Avoid_: 初期同期, マスター同期

**Toggle Voice**:
VRChat のマイク操作モード。Mic Mute Sync は Toggle Voice が ON のときのみ動作する前提とする。
_Avoid_: トグルボイス, Push-to-Mute 設定

**Discord Application**:
Mic Mute Sync が Discord Voice RPC に接続するとき使う、VRChat Tweaker プロジェクトが登録・同梱する Discord アプリ（Client ID）。
_Avoid_: Discord Bot, OAuth アプリ（汎称）

**Sync Pause**:
Mic Mute Sync Session 中に VRChat または Discord のどちらかが利用不能になった間、同期を一時停止すること。
_Avoid_: 同期オフ, セッション終了

**Mic Mute Sync Availability**:
Mic Mute Sync が提供される実行環境。初版は Windows のみを対象とする。
_Avoid_: プラットフォーム対応, OS 制限

**Mic Mute Sync Settings**:
Settings 画面にある Mic Mute Sync 専用の設定領域。Automation ルールとは別に管理する。
_Avoid_: 連動設定, ミュート設定

**Sync Status**:
Mic Mute Sync Settings に表示する接続・同期の状態。VRChat と Discord それぞれの Mic Mute 状態、および前提条件（OSC 接続、Toggle Voice、Discord RPC 認可など）のチェックリストを含む。問題発生時もデスクトップ通知は使わず、ここで理由を示す。
_Avoid_: 同期状態, 接続ステータス

**OSC Endpoint**:
Mic Mute Sync が VRChat と通信する OSC の接続先（`inPort:outIP:outPort`）。Mic Mute Sync Settings で設定し、未指定時は VRChat デフォルトを使う。
_Avoid_: OSC ポート, OSC 設定
