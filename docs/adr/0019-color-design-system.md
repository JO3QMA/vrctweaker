# ADR 0019: Color design system

## Status

Accepted（grill-with-docs で合意）

## Context

- 色が `--accent` / `--bg-primary` / `--el-color-primary` / コンポーネント内 hex 直書きに分散し、同じ意味の色が二重定義されている
- **VtButton**（[ADR 0017](0017-button-design-system-vtbutton.md)）と **Spacing**（[ADR 0018](0018-spacing-design-system.md)）は整備済みだが、カラーパレットの正本は未定義
- **Primary button**（操作様式）と「メインカラー」の俗称が衝突しやすい
- **Semantic color**（`el-tag` / `ElMessage` のフィードバック）と **Semantic button**（Presence 色ボタン）、**Server status color**（5 段＋メンテ青）が「状態色」として混同されやすい
- アプリは **ダークテーマのみ**（`html.dark`）。ライトモード切替はない
- Element Plus は `--el-color-*-light-*` 等の派生色を多数持つが、画面作者が直接触る必要はない

用語は [`CONTEXT.md`](../../CONTEXT.md) の **Design system** セクション（**Color**、**Brand color**、**Neutral color**、**Semantic color**、**Domain color** 等）を正本とする。

## Decision

### 1. 正本と委譲

- **App color token**（`:root` の `--color-*`）が唯一の色の出所
- **Element Plus color mapping**（`html.dark` 内）で `--el-color-*` 等へ委譲する
- 画面・共有スタイルは原則 **App color token** を参照する。`--el-*` 直参照は EP 上書きブロックと既存 EP 利用箇所に限定する
- EP の `light-*` / `dark-*` は **Element Plus color derivative** として mapping 内のみ定義し、**App color token** には含めない

### 2. カテゴリとトークン（v1・ダークのみ）

#### Brand color

| 役割 | CSS 変数 | 備考 |
|------|----------|------|
| 本体 | `--color-brand` | リンク、フォーカス、**Primary button** 塗り等 |
| hover | `--color-brand-hover` | |

「Primary color」は使わない（**Primary button** と混同するため）。

#### Neutral color

| 役割 | CSS 変数 | レガシー alias（移行期） |
|------|----------|--------------------------|
| 表面 base | `--color-bg-base` | `--bg-primary` |
| 表面 elevated | `--color-bg-elevated` | `--bg-secondary` |
| 表面 muted | `--color-bg-muted` | `--bg-tertiary` |
| 本文 | `--color-text-primary` | `--text-primary` |
| 補足 | `--color-text-secondary` | `--text-secondary` |
| プレースホルダ・無効 | `--color-text-muted` | （新設。EP `placeholder` / `disabled` へ mapping） |
| 枠線 | `--color-border` | `--border` |

Neutral border の多段（`border-light` 等）は v1 では **App color token** に載せず、mapping 内導出のみ。

#### Semantic color

| 意味 | CSS 変数 | レガシー alias |
|------|----------|----------------|
| danger | `--color-danger` | `--danger` |
| success | `--color-success` | `--success` |
| warning | `--color-warning` | （新設。現行 `--el-color-warning` と同値） |
| info | `--color-info` | （新設。現行 `--el-color-info` と同値） |

`--el-color-error` は `--color-danger` と同値でよい（EP 互換）。hover や `light-*` は **Semantic color token** に含めない。

#### Domain color

**Server status color**（`ServerStatusSection.vue`）:

| 状態 | CSS 変数 |
|------|----------|
| operational | `--color-status-operational` |
| degraded | `--color-status-degraded` |
| partial | `--color-status-partial` |
| major | `--color-status-major` |
| maintenance | `--color-status-maintenance` |
| unknown | `--color-status-unknown` |

**Presence color**（`PresenceChangeSection.vue`）:

| プレゼンス | 背景 | 枠（現行どおり 2 トークン） |
|------------|------|------------------------------|
| Join Me | `--color-presence-join-me` | `--color-presence-join-me-border` |
| Active | `--color-presence-active` | `--color-presence-active-border` |
| Ask Me | `--color-presence-ask-me` | `--color-presence-ask-me-border` |
| Busy | `--color-presence-busy` | `--color-presence-busy-border` |

**Domain color** を **Semantic color** へ流用しない（メンテナンス青・partial 橙のため）。

#### その他レガシー alias

| 旧 | 新 |
|----|-----|
| `--accent` | `--color-brand` |
| `--accent-hover` | `--color-brand-hover` |

**Color alias** は委譲のみ（別 hex を持たない）。v1 では削除しない。

### 3. Color v1 visual policy

- tokenize 時は **既存 hex をそのまま移す**。見た目・コントラスト・色相の調整は v1 スコープ外
- パレット改善はトークン整備後の別 PR / Issue とする

### 4. 実装正本

- トークン定義・EP mapping・alias: `frontend/src/assets/style.css`
- カタログ定数（Storybook・テスト用）: `frontend/src/design/colorTokens.ts`
- 見た目と使い方: **Color Storybook catalog**（`frontend/src/design/Color.stories.ts`）
- Agent 規約: `.cursor/rules/color-design-system.mdc`（実装 PR で追加）

### 5. Color adoption

- 新規・改修で触ったファイルでは **App color token** を使い、hex 直書きやレガシー名を増やさない
- **Color v1 deliverables** で寄せる対象:
  - `style.css` の全カテゴリトークンと共有クラス（`.page-title`、`.section-card` 等）
  - **Server status color**・**Presence color** の既知 2 コンポーネント
- 既存 View の scoped hex は一括置換しない
- Element Plus 内部色の全面再設計は v1 スコープ外
- ESLint 強制は v1 では入れない

## Considered options

| 案 | 却下理由 |
|----|----------|
| Element Plus `--el-color-*` を正本 | Server status / Presence 等のドメイン色と二重管理が続く |
| App と EP の二重カタログ（別 hex） | 同期ずれとメンテ負荷 |
| 「Primary color」でメイン色を命名 | **Primary button**（ADR 0017）と用語衝突 |
| Semantic 3 色のみ（info なし） | EP `type="info"` とずれる |
| Server status を Semantic color に畳む | メンテナンス青・partial 橙が壊れる（[CONTEXT Server status presentation](../../CONTEXT.md) と矛盾） |
| ライト＋ダーク両方を v1 で定義 | 現行 UI にライトモードがなく YAGNI |
| v1 で hex 値も調整 | リネーム PR と視覚回帰レビューが混ざり、影響範囲が大きい |
| EP 派生色をすべて App token 化 | カタログ肥大。画面から直接参照する需要がない |
| 全 View 一括置換 | Spacing / VtButton adoption と方針が矛盾しリグレッションリスクが大きい |

## v1 scope

**含む:**

- `style.css` への **App color token**、**Element Plus color mapping**、**Color alias**
- `colorTokens.ts` + 単体テスト
- **Color Storybook catalog**（Brand / Neutral / Semantic / Server status / Presence）
- `ServerStatusSection.vue` / `PresenceChangeSection.vue` のトークン参照化
- 共有スタイルの Neutral 置換
- `.cursor/rules/color-design-system.mdc`
- `CONTEXT.md` / 本 ADR

**含まない:**

- ライトテーマ用トークン、`prefers-color-scheme`、テーマ切替 UI
- hex 値・コントラストのリデザイン
- 全 View の scoped hex 一括置換
- Neutral border 多段の App token 化
- EP 派生色の App token 化
- ESLint hex 直書き禁止

## Consequences

- 新規・改修 PR では `var(--color-*)` が期待される（レガシー alias は移行期のみ）
- **Color Storybook catalog** が色の正本となり、Spacing / VtButton catalog と並べて参照できる
- Server status と Presence は **Domain color** として独立したトークン名を持ち、Semantic フィードバック色と混同しにくくなる
- 見た目不変のため v1 マージ直後の視覚差は意図しない（alias 経由の参照先同一）
- パレット調整はトークン名が安定した後に、Storybook 上で差分比較しやすくなる
