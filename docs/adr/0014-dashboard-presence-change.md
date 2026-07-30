# ADR 0014: Dashboard presence change section

## Status

Accepted（grill-with-docs で合意、[Issue #174](https://github.com/JO3QMA/vrctweaker/issues/174)）

## Context

- Dashboard のプレゼンス変更 UI が **Quick status**・**Custom status**・**Templates** の 3 カードに分かれ、色と文章が別状態になり、どれが VRChat に反映されるか分かりにくい
- VRChat のプレゼンスは実質 **4 色**（join me / active / ask me / busy）と **任意の statusDescription**（空ならクライアント既定）の組み合わせ
- 現行は色ボタンが即 `SetStatus`、文章が `SetStatusDescription`、Templates が `SetStatusAndDescription` と API 経路が分裂している
- Server status（ADR 0009）・Launch block（ADR 0010）と同様、常時表示ブロックの読み取り失敗は `ElMessage` で汚さない方がよい

用語は [`CONTEXT.md`](../../CONTEXT.md) の **Dashboard** セクション（**Presence change section**、**Presence change draft**、**Status description history** 等）を正本とする。

## Decision

1. **UI**: **Presence change section** — `PresenceChangeSection.vue`。Launch block 直下（旧 3 カードの位置）。**4 色選択**（既存パレット: 青 / 緑 / 橙 / 赤）+ **文章入力**（`el-autocomplete` で履歴候補）+ **反映ボタン** 1 つ
2. **Draft**: 色クリックは **draft の色のみ更新**（即 API しない）。文章候補選択も **draft の文章のみ**。反映まで送らない
3. **Prefill**: ログイン済みなら **Presence change section fetch** で self の `status` / `statusDescription` を draft と **sync snapshot** に取り込む。`statusDescription` が空なら欄も空
4. **Apply**: **`ApplyPresenceChange(status, description)`** 1 呼び出しで VRChat API 更新 + 成功時のみ **Status description history** 追記（trim 後非空）。空白文章は `statusDescription: ""`（ロケール既定文言を API に埋め込まない）
5. **Apply ボタン**: draft が sync snapshot と同一なら **disabled**
6. **Dirty**: ユーザーが色または文章を触ったら dirty。dirty でなければ **identity:self-cache-changed** で debounce 再取得（300ms）し draft / snapshot を同期
7. **表示**: セクションは **常に表示**。**Presence change visibility** — 未ログインは全コントロール disabled + Settings 導線。ログイン済みは loading → フォーム、失敗は **Presence change fetch failure**（カード内エラー、toast なし）
8. **履歴**: `app_settings` キー **`presence_description_history`**（JSON 文字列配列）。最大 **20** 件、完全一致は先頭へ移動（重複排除）。色は含めない。旧 Templates（i18n 固定 3 件）は廃止・シードしない
9. **Wails 読み取り**: **`GetPresenceChangeSection()`**。同一 PR で Dashboard 向け **`SetStatus` / `SetStatusDescription` / `SetStatusAndDescription` の Wails バインディングを削除**（`IdentityUseCase` 内部・Automation は既存 `SetStatus` を維持）
10. **イベント**: self 行更新時に **`identity:self-cache-changed`** を emit（Pipeline self `user-update`、`ApplyPresenceChange` 成功、Self profile refresh 等）
11. **i18n**: **`dashboard.presenceChange.*`** に統一。旧 `dashboard.quickStatus` / `customStatus` / `templates*` は削除。全 5 ロケール同期
12. **PR scope**: **1 PR**（Go + Vue + i18n×5 + Vitest + E2E mock 追随）

## v1 scope 外

- `offline` の色選択（4 色のみ）
- 履歴の手動編集・削除 UI
- Server status / Launch block の変更
- Automation `change_status`（既存のまま status のみ変更）
- 手動リトライボタン（再表示 `onMounted` で十分）

## Return-value contract

### 読み取り — `GetPresenceChangeSection() (PresenceChangeSectionDTO, error)`

| 状況 | App / Usecase | Frontend |
|------|---------------|----------|
| 未ログイン | `loggedIn: false`、history は返してよい（空可） | disabled + loginRequired 文言 + Settings リンク |
| ログイン済み・成功 | `loggedIn: true` + status + statusDescription + history | フォーム表示・prefill |
| infra / self 取得失敗 | **`error`** | カード内 `loadError`。toast なし |
| 読み込み中 | — | `loading` 文言（コントロール非表示） |

```go
type PresenceChangeSectionDTO struct {
    LoggedIn          bool     `json:"loggedIn"`
    Status            string   `json:"status"`            // loggedIn 時のみ。join me | active | ask me | busy
    StatusDescription string   `json:"statusDescription"` // loggedIn 時のみ
    History           []string `json:"history"`           // 新しい順。最大 20
}
```

- `status` は VRChat API 値（小文字、`join me` はスペース含む）。`offline` および 4 色以外は **読み取り時に `active` へ正規化**して DTO に載せる（apply は 4 色のみ受け付け）
- `History` は未ログインでも返してよい（端末ローカル履歴）

### 書き込み — `ApplyPresenceChange(status, description string) (PresenceChangeApplyResultDTO, error)`

| 状況 | App / Usecase | Frontend |
|------|---------------|----------|
| 成功 | VRChat PUT + history 追記 + self cache 更新 + **`identity:self-cache-changed` emit** | `ElMessage.success`。返却 DTO で draft / snapshot 同期 |
| 未ログイン / 空 user id | **`error`** | `ElMessage.error` |
| バリデーション（未知 status、description > 32 runes） | **`error`** | `ElMessage.error` |
| VRChat API / セッション失敗 | **`error`** | `ElMessage.error`。draft 保持 |

```go
type PresenceChangeApplyResultDTO struct {
    Status            string `json:"status"`
    StatusDescription string `json:"statusDescription"`
}
```

- `description` は trim 後を使用。空文字可
- history 追記は trim 後非空のみ

## History persistence（Go）

- キー: `presence_description_history`
- 値: `json.Marshal([]string{...})`
- 追記: 先頭に insert、既存と同一文字列は除去してから先頭へ、長さ > 20 なら末尾 truncate
- 読み取り失敗は `GetPresenceChangeSection` の infra 失敗として扱う

## Considered options

| 案 | 却下理由 |
|----|----------|
| 色クリック即 `SetStatus` | draft と API が分裂。Issue の一括反映と矛盾 |
| 空白時に i18n 既定文言を API 送信 | フレンド側表示がロケール依存になる |
| Templates パネル維持 + 履歴併存 | 「独立したテンプレート状態」が残る |
| `vrchat:friends-changed` で再同期 | フレンド一覧更新と無関係な再取得が多い |
| フロントが `setStatusAndDescription` + 別 history API | 履歴追記の原子性が弱い |
| 未ログインでセクション非表示 | Server status との並び・導線が弱い |
| 取得失敗時に空フォーム | 実プレゼンスと無関係な編集を許す |

## Consequences

- **Quick status**（CONTEXT）は旧称。UI・i18n から除去
- Frontend **Async UI**: Launch block / Server status と同型の **`generation`** + `inFlight`。`onUnmounted` で increment。load / debounce event の `await` 後に古い世代なら `ref` 更新をスキップ。apply 失敗の `ElMessage` はアンマウント後も表示してよい
- **data-testid**: `presence-change-section`、`presence-change-color-join-me` / `active` / `ask-me` / `busy`、`presence-change-description`、`presence-change-apply`、`presence-change-login-required`、`presence-change-loading`、`presence-change-load-error`
- E2E `app.spec.ts` の `dashboard-quick-status-*` は新 testid に追随

## Test plan（review-ready）

| 層 | 入力・状況 | 期待 | テスト名（案） |
|----|------------|------|----------------|
| Go | 未ログイン | `loggedIn: false` | `TestGetPresenceChangeSection_notLoggedIn` |
| Go | ログイン・self あり | status + description + history | `TestGetPresenceChangeSection_loggedIn` |
| Go | self 取得失敗 | `error` | `TestGetPresenceChangeSection_selfError` |
| Go | apply 成功・description 非空 | history 先頭追記 | `TestApplyPresenceChange_recordsHistory` |
| Go | apply 成功・description 空 | history 変化なし | `TestApplyPresenceChange_emptyDescriptionNoHistory` |
| Go | history 21 件目 | 20 件に truncate | `TestPresenceDescriptionHistory_max20` |
| Go | duplicate history | 先頭へ移動 | `TestPresenceDescriptionHistory_dedupe` |
| Go | invalid status | `error` | `TestApplyPresenceChange_invalidStatus` |
| Go | description 33 runes | `error` | `TestApplyPresenceChange_descriptionTooLong` |
| Go | apply 成功 | emit `identity:self-cache-changed` | `TestApplyPresenceChange_emitsSelfCacheChanged` |
| Vitest | load 成功 | 4 色 + 入力 + apply | `shows presence change form` |
| Vitest | 未ログイン | disabled + Settings リンク | `shows login required state` |
| Vitest | load error | インラインエラー | `shows inline error on load failure` |
| Vitest | 色クリック | API 未呼び出し | `color click updates draft only` |
| Vitest | draft 変更なし | apply disabled | `disables apply when unchanged` |
| Vitest | apply 成功 | success toast | `applies presence change` |
| Vitest | apply 失敗 | error toast、draft 保持 | `shows error on apply failure` |
| Vitest | self-cache event、not dirty | debounce reload | `reloads on self cache changed when not dirty` |
| Vitest | self-cache event、dirty | draft 保持 | `skips reload when dirty` |
| Vitest | unmount 後 load | ref 更新なし | `skips state update after unmount` |
| E2E | mock 正常 | `presence-change-section` 表示 | `dashboard shows presence change section` |

## References

- [Issue #174](https://github.com/JO3QMA/vrctweaker/issues/174)
- [Issue #32](https://github.com/JO3QMA/vrctweaker/issues/32)（Dashboard 親 Issue）
- [ADR 0009](0009-dashboard-server-status.md)
- [ADR 0010](0010-dashboard-launch-block.md)
