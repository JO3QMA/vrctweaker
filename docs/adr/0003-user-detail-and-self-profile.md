# ADR 0003: User detail 共通化と Self profile（`/me`）

## Status

Accepted（grill-with-docs セッションで合意）

## Context

- Friends（マスター／ディテール）、User profile（`/user-profile`）、Settings のログイン済みブロックが、それぞれ別 UI・別データ経路でユーザープロフィールを示していた
- User detail の共有コンポーネント（`VrcUserCacheDetail`）は Friends / User profile で使われるが、Settings は `VRChatCurrentUserDTO` の簡易カードのみ
- バックエンドには `users_cache` の `user_kind=self` 行と `GetCurrentUser` による upsert がある一方、フロントは二系統の DTO を参照していた
- Activity の **Encounter user navigation** はフレンド／非フレンドの二択のみで、自分自身の行き先が未定義だった
- `user_encounters` は他ユーザーの滞在区間のみを記録するため、自分の VRC user ID で絞った遭遇履歴タブは常に空になる

用語は [`CONTEXT.md`](../../CONTEXT.md) の **User detail** セクションを正本とする。

## Decision

1. **User detail** を傘用語とし、Friends・User profile・Self profile はいずれも同じ閲覧表面（ヒーロー・詳細タブ等）を共有する
2. **Self profile** は専用ルート **`/me`** で全面表示する。サイドバー常設（`nav.me`、日本語「自分」／英語「Me」）と Settings profile summary の「詳細を見る」の両方から開ける
3. **Cached VRChat user**（`users_cache`、self 行含む）を User detail のデータの正とする。Settings profile summary は同じ self 行の最小投影とし、`VRChatCurrentUserDTO` 専用経路への依存をやめる方向とする
4. **Self profile の UI 差分**: お気に入り非表示、遭遇履歴タブ非表示、詳細タブに **Self profile refresh**（Settings のプロフィール更新と同等）を置く
5. **Self profile navigation**: `vrcUserId` 導線で対象がログイン中の自分なら常に `/me` へ。Friends / User profile へフォールバックしない。Encounter user navigation も同ルールに従う
6. **未ログイン時の `/me`**: Settings へリダイレクトせず、Friends と同型のログイン必要空状態＋Settings 導線を `/me` 上に示す。サイドバーの「自分」は未ログインでも表示し、同じ空状態へ進める

## Considered Options

| 論点 | 採用 | 却下した案 |
|------|------|------------|
| Self profile の入口 | 専用ルート `/me` + Settings 要約リンク | Friends 一覧に「自分」行、Settings 内への User detail 埋め込み |
| Self のデータ源 | Cached VRChat user（self 行）に一本化 | Self だけ `GetVRChatCurrentUser` を継続 |
| Self の遭遇履歴タブ | 非表示 | 他ユーザーと同一タブ（常に空）、Activity への導線に置換のみ |
| 自分への deep link | `/me` に統一 | `/user-profile?vrcUserId=自分` のまま |
| 未ログイン `/me` | 画面上の空状態 | Settings リダイレクト、`/me` 内ログインフォーム |
| ルートパス | `/me` | `/self-profile`、`/profile/me` |

## Consequences

### 正

- Friends / User profile / Self profile の体験とデータモデルが揃う
- 自分のプロフィールは常に `/me` と覚えやすい
- 意味のない空の遭遇履歴タブを Self profile から排除できる
- Encounter user navigation の三分岐（自分／フレンド／その他）が明確になる

### 負

- `ResolveUserProfileNavigation`（または同等）に self 判定の第三分岐が必要
- Settings のプロフィール表示を self 行ベースに寄せるリファクタが発生する
- User detail 共有コンポーネントに self 向け props／分岐（お気に入り・タブ・refresh）が増える

## Implementation

Implemented per ADR 0003.
