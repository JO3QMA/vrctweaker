# ADR 0004: Listable friend と Pipeline プレゼンスの分離

## Status

Accepted（grill-with-docs セッションで合意）

## Context

- VRChat Pipeline の `friend-active` 等は **プレゼンス**（status / platform / location）のみを送り、表示名を含まないことがある
- 従来実装は Pipeline マージ時に `user_kind=friend` へ昇格し、表示名空の行が Friends 一覧に露出した（Issue #122）
- 実 DB では無名 `friend` 行が REST フレンド同期を経ず、遭遇履歴も無い状態で残っていた
- `users_cache` の `user_kind=friend` は Friends 一覧・お気に入り・オンライン通知など複数経路で参照される

用語は [`CONTEXT.md`](../../CONTEXT.md) の **User detail** セクション（**Listable friend**、**Profile resolution**、**Unresolved friend presence**）を正本とする。

## Decision

1. **Listable friend**（Friends マスター一覧に載せる行）は、**表示名非空**かつ **VRChat API がフレンド関係を肯定**（`isFriend=true` または Friends REST 同期一覧に含まれる）ことが必須。Pipeline プレゼンス単体では Listable friend にしない
2. **二重ガード**: 書き込み時に Pipeline プレゼンス系で `user_kind=friend` へ昇格させない。読み取り時（`List` / `ListFavorites`）でも `display_name` 非空を要求する
3. **Profile resolution（ハイブリッド）**: Pipeline 受信時に `GET /users/{id}` を試行。失敗時は **Unresolved friend presence** として `user_kind=contact` でプレゼンスのみ保持し、Reconcile / RefreshFriends で再試行する
4. **`friend` 昇格**は `GET /users/{id}` で `isFriend=true` のとき、または Friends REST 同期で一覧に含まれたときに限る
5. **既存データ**: 起動時マイグレーションで表示名空の `friend` 行を `contact` に降格し、お気に入りフラグをクリアする。降格した ID へログイン済みセッションで Profile resolution を 1 回試行する

## Considered Options

| 論点 | 採用 | 却下した案 |
|------|------|------------|
| 無名行の一覧表示 | 出さない（Listable friend 条件） | 出す＋ UI フォールバックのみ |
| ガード層 | 読み取り＋書き込み | 読み取りのみ / 書き込みのみ |
| 解決失敗時 | `contact` でプレゼンス保持 | 捨てる / 既存行があるときだけ更新 |
| 昇格条件 | `isFriend=true` 等 API 肯定 | Pipeline イベントだけで昇格 |
| 既存無名行 | 起動時降格＋バックフィル | Reconcile 待ち / 読み取り隠蔽のみ |
| 降格時のお気に入り | クリア | 保持 |

## Consequences

### 正

- Friends 一覧と User detail の前提（表示名あり）が揃う
- Pipeline のリアルタイム性と REST のプロフィール正本性の役割分担が明確になる
- お気に入り通知が無名行に対して誤動作しにくい

### 負

- Pipeline ハンドラに `GET /users/{id}` が増え、イベント処理が重くなる（失敗時は contact 保留で緩和）
- マイグレーション直後、バックフィル完了まで一瞬フレンドが一覧から消えて見える可能性がある
- `user_kind=contact` が「遭遇由来 Contact」と「Unresolved friend presence」の両方を含む（遭遇履歴の有無で区別）

## Implementation

Issue #122 参照。`identity_usecase_pipeline.go`、`identity_repo.List`、`db` マイグレーション、Reconcile バックフィル。
