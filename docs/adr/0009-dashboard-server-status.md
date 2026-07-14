# ADR 0009: Dashboard Server status

## Status

Accepted（grill-with-docs / grill-review-ready で合意、[Issue #10](https://github.com/JO3QMA/vrctweaker/issues/10)）

## Context

- VRChat のサービス障害・メンテ中に、起動やログインの失敗がローカル環境由来か公式側か切り分けたい（Issue #10）
- 公式 [status.vrchat.com](https://status.vrchat.com/) は Statuspage API（`summary.json` / `components.json` 等）を公開している。認証不要
- Dashboard には既に **Quick status**（join me / busy 等の個人プレゼンス変更）があり、**Server status**（インフラ健全性）と混同しやすい
- 本番 CSP は `connect-src 'self'` で、外部 HTTP は Go 経由が既存方針（`main.go` の AssetServer middleware コメント）。フロントから status.vrchat.com へ直接 `fetch` しない
- `GetInstanceRejoinSection` は infra 失敗時 `(nil, error)` + `ElMessage.error` だが、Server status は常時表示・定期ポーリングのため別契約が必要

用語は [`CONTEXT.md`](../../CONTEXT.md) の **Dashboard** セクション（**Server status**、**Server status section**、**Quick status** 等）を正本とする。

## Decision

1. **対象**: status.vrchat.com のサービス健全性のみ。個人プレゼンス（Quick status）・フレンドオンラインとは別
2. **配置**: Dashboard の **Quick launch より上**に **Server status section**（`ServerStatusSection.vue`）
3. **表示量**: **Server status summary** を常時。`operational` 以外の component が 1 件でもあれば **Abnormal server status** とし **Server status detail** を展開。平常時はコンパクトな 1 行 + status.vrchat.com への外部リンク
4. **detail 内容**: (1) リーフ component のうち **`operational` 以外のみ**（フラット。親グループ見出しは v1 なし）、(2) 未解決インシデント見出し（あれば）、(3) 予定または進行中メンテ見出し（あれば）。component 名・インシデント／メンテ見出しは API 原文（英語）。ステータス値とサマリー文言は i18n
5. **色**: `el-card` 枠 + status.vrchat.com に近い色分け（正常=緑、低下=黄、部分障害=橙、重大=赤、メンテ=青系）
6. **ログイン**: **不要**。未ログインでも表示・取得する
7. **取得経路**: Wails **`GetServerStatus()`** — Go が status.vrchat.com を HTTP GET。`User-Agent` は `vrchatapi.UserAgent` を流用
8. **更新**: Dashboard 表示中のみ。`onMounted` で即 1 回 + **5 分**間隔ポーリング。`onUnmounted` で `clearInterval`。手動リフレッシュ・`document.visibilitychange` による停止は v1 なし
9. **外部リンク**: `<a href="https://status.vrchat.com" target="_blank" rel="noopener noreferrer">`（`LicensesView` と同パターン）
10. **出荷**: **PR1** Go（infra + usecase + binding + テスト）→ **PR2** Vue + i18n×5 + E2E mock

## Failure modes（review-ready）

| 状況 | ユーザー | サーバ |
|------|----------|--------|
| 全体取得失敗（タイムアウト、DNS、非 200、不正 JSON、body 超過） | セクション内「取得できませんでした」（i18n）。detail なし | `fetchState=unavailable` を DTO で返す。ログに `err` または非 200 + body 先頭 200 文字 |
| **部分失敗**（summary OK、components / incidents / maintenances のいずれか失敗） | サマリー表示 + detail 領域に「詳細を取得できませんでした」 | `fetchState=partial` |
| ポーリング失敗（上記） | 同上。**`ElMessage` は出さない**（5 分ごとの toast 汚染を避ける） | 同上 |
| 全 component `operational` | コンパクトサマリー + リンク。component リストなし | `fetchState=ok`、components 空 |
| Abnormal | サマリー + 非 operational component + 見出し（取れた分） | `fetchState=ok` |

Quick launch・Instance rejoin・Quick status は Server status 失敗の影響を受けない（非ブロッキング）。

## Return-value contract（review-ready）

### 読み取り — `GetServerStatus() (ServerStatusDTO, error)`

| 層 | 取得失敗（infra / HTTP） | 部分成功 | 成功 |
|----|--------------------------|----------|------|
| **Usecase / App** | **`(DTO{fetchState:"unavailable"}, nil)`** | **`(DTO{fetchState:"partial", summary:…}, nil)`** | **`(DTO{fetchState:"ok", …}, nil)`** |
| **Programming mistake**（例: subsystem 未初期化） | 読み取り系は `(DTO{fetchState:"unavailable"}, nil)` または `error` のどちらかに **統一**（yt-dlp `notInitialized` パターン） | — | — |
| **Frontend** | セクション内 i18n エラー。toast なし | サマリー + detail 失敗文言 | 通常表示 |

- `fetchState` は Go が返す生文字列 `ok` | `unavailable` | `partial`。翻訳はフロントの `dashboard.serverStatus.*`
- **`callApp` fallback**（`app.ts`）は毎回 **新規** `{ fetchState: "unavailable", … }` を返す（共有 mutable 禁止）
- Instance rejoin と異なり、infra 失敗を **`error` で Wails に上げない**（ポーリング前提）

DTO 概要:

```go
type ServerStatusDTO struct {
    FetchState   string                      `json:"fetchState"`
    Summary      ServerStatusSummaryDTO      `json:"summary"`
    Components   []ServerStatusComponentDTO  `json:"components"`   // non-operational only when ok/partial
    Incidents    []ServerStatusHeadlineDTO   `json:"incidents"`
    Maintenances []ServerStatusHeadlineDTO   `json:"maintenances"`
}
```

## Config（review-ready）

v1 は **Settings / env 追加なし**。定数は Go に固定:

| 定数 | 値 | 備考 |
|------|-----|------|
| HTTP timeout | 15s | `vrchatapi.Client` と同じ |
| Max body | 1 MiB | `LimitReader` + 超過は失敗 |
| Allowed host | `status.vrchat.com` | https のみ。リダイレクト先も同ホストのみ |
| Poll interval | 5 min | **Vue のみ**（`ServerStatusSection.vue`） |

API エンドポイント（ベース `https://status.vrchat.com/api/v2/`）:

- 常時: `summary.json`
- Abnormal 判定・detail: `components.json`
- detail 見出し: `incidents/unresolved.json`、`scheduled-maintenances/active.json`（または `upcoming.json` — 実装時に active を優先し、空なら upcoming）

同一 `GetServerStatus` 呼び出し内は **`errgroup` で並列 GET**（ネットワーク I/O は mutex 外）。

## Trust boundaries（review-ready）

| 境界 | 方針 |
|------|------|
| **URL** | コード内ベース URL 固定 + host allowlist。ユーザー入力なし |
| **リダイレクト** | 許可外ホストへ飛ぶ 302 は拒否 |
| **ログ** | 非 200 時は status + body 先頭 200 文字まで。成功レスポンス全文は出さない |
| **User-Agent** | `vrchatapi.UserAgent` |
| **Frontend** | 公開 API の component 名・インシデント見出しをそのまま表示（PII は通常含まれない）。`docs/agents/redaction.md` は公開成果物向け |

## Lifecycle（review-ready）

- **Go**: 常駐 goroutine なし。リクエストごとに `context` 付き HTTP（App の `a.ctx` またはリクエスト scope）
- **Vue**: `inFlight` — 前回 `GetServerStatus` 未完了なら poll tick をスキップ
- **Vue**: **generation カウンタ** — `onUnmounted` で increment し、await 後に古い generation なら `ref` 更新しない
- **Vue**: `setInterval` は `onUnmounted` で `clearInterval`

## Interim / v1 scope（review-ready）

### v1 でやる

- Dashboard **Server status section** のみ
- detail の component は **フラット・非 operational のみ**

```go
// ponytail: flat leaf components only; no parent-group headings yet.
// Upgrade path: group by components[].group_id to match status.vrchat.com.
```

### v1 スコープ外（Issue #10 以降）

- 障害検知時の OS 通知
- Settings でのオンオフ・ポーリング間隔変更
- リージョン絞り込み
- サイドバー等 Dashboard 以外への常設
- 取得結果のローカル履歴・グラフ
- `document.visibilitychange` による poll 停止
- status ページの親グループ見出し表示

## Edge cases → tests（review-ready）

### Backend（`internal/infrastructure/statuspage` + usecase）

| 入力 | 期待 | テスト名 |
|------|------|----------|
| 全 component `operational` | `fetchState=ok`、components 空 | `TestFetch_allOperational` |
| 1 件 `under_maintenance` | `fetchState=ok`、components 1 件 | `TestFetch_abnormalOneComponent` |
| summary 失敗 | `fetchState=unavailable` | `TestFetch_summaryFailure` |
| summary OK、components 失敗 | `fetchState=partial` | `TestFetch_partialComponentsFailure` |
| 不正 JSON | `fetchState=unavailable` | `TestFetch_invalidJSON` |
| body > 1MiB | `fetchState=unavailable` | `TestFetch_bodyTooLarge` |
| 302 → 許可外ホスト | 失敗 | `TestFetch_redirectHostRejected` |
| HTTP 503 | `fetchState=unavailable` | `TestFetch_non200` |

### Frontend（`ServerStatusSection.spec.ts` / `DashboardView.spec.ts`）

| 入力 | 期待 | テスト名 |
|------|------|----------|
| `fetchState=unavailable` | セクション内エラー、detail なし | `shows server status fetch failure` |
| `fetchState=partial` | サマリー + detail 失敗文言 | `shows partial server status detail failure` |
| `fetchState=ok`、全 operational | コンパクトサマリー、リストなし | `hides server status detail when operational` |
| `fetchState=ok`、abnormal | 非 operational のみ | `shows non-operational components only` |
| unmount 後 resolve | ref 更新なし | `does not update server status after unmount` |
| inFlight 中の tick | 2 回目スキップ | `skips overlapping server status poll` |

### E2E

| 入力 | 期待 | テスト名 |
|------|------|----------|
| mock `GetServerStatus` 正常 | section visible（`data-testid`） | `dashboard shows server status section` |

## i18n（review-ready）

- キー名前空間: **`dashboard.serverStatus.*`**（`dashboard.quickStatus` と並列）
- ロケール: `en` / `ja` / `ko` / `zh-CN` / `zh-TW` すべてに同一キー
- component 名は翻訳しない。`operational` 等のステータス ID → i18n マップはフロント（例: `dashboard.serverStatus.statusOperational`）

## Considered options

| 案 | 却下理由 |
|----|----------|
| フロントから status.vrchat.com 直 `fetch` | CSP `connect-src 'self'` 変更が必要。既存「外部 HTTP は Go」と不一致 |
| infra 失敗を `(nil, error)` + `ElMessage`（Instance rejoin 同様） | 5 分 poll で toast 汚染。セクション常時表示と矛盾 |
| 平常時も全 component 一覧 | Dashboard ノイズ。Issue の意図（障害時の切り分け）に対して過剰 |
| stale 表示（最後の成功値を保持） | 古い「正常」表示のリスク。grill-with-docs で却下 |
| 1 PR に Go+Vue 一括 | レビュー可能だが、HTTP 層と UI 層の関心分離のため **2 PR** を採用 |

## Consequences

- Dashboard にインフラ系と個人プレゼンス系の 2 種の「status」が並ぶ。用語と UI ラベルで **Server status** / **Quick status** を厳密に分ける
- `GetServerStatus` は VRChat セッション不要。オフライン時も「取得できませんでした」と区別できる
- 将来 Settings で poll 間隔を変える場合は本 ADR の Config 節と `Server status v1 scope` を更新する

## Related

- [Issue #10](https://github.com/JO3QMA/vrctweaker/issues/10)
- [status.vrchat.com API](https://status.vrchat.com/api/)
- [ADR 0006](0006-instance-rejoin-and-last-launch-profile.md)（Dashboard 上の別セクション・別 Wails 契約の先例）
