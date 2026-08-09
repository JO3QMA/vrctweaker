# ADR 0020: Typography design system

## Status

Accepted（grill-with-docs で合意）

## Context

- フォントサイズが `rem`（14px ルート）と任意 `px` で混在し、`0.9rem`・`12px`・`1.05rem` など **Typography scale に乗らない値**が散見される
- **VtButton**（[ADR 0017](0017-button-design-system-vtbutton.md)）、**Spacing**（[ADR 0018](0018-spacing-design-system.md)）、**Color**（[ADR 0019](0019-color-design-system.md)）は整備済みだが、タイポグラフィの正本は未定義
- `html { font-size: 14px }` のため `rem` ベースのサイズトークンはスケールと噛み合わない（Spacing ADR と同様、ルートは v1 では変更しない）
- **Spacing** v1 では line-height を対象外とした。用途別の行高は Typography で正本化する
- Element Plus は `--el-font-size-*` を持つが、画面作者が直接触る必要は限定的

用語は [`CONTEXT.md`](../../CONTEXT.md) の **Design system** セクション（**Typography**、**Typography scale**、**Text style catalog** 等）を正本とする。

## Decision

### 1. Typography scale（font-size）

| px | CSS 変数 |
|----|----------|
| 10 | `--font-size-10` |
| 12 | `--font-size-12` |
| 14 | `--font-size-14` |
| 16 | `--font-size-16` |
| 18 | `--font-size-18` |
| 20 | `--font-size-20` |
| 24 | `--font-size-24` |

**Font size derivative**（スケール外・レガシー維持）:

| 用途 | CSS 変数 | 値 |
|------|----------|-----|
| Heading 1 / page title | `--font-size-h1` | `calc(var(--font-size-14) * 1.4)`（19.6px @ 14px root） |

- 上記以外の任意サイズは新規・改修で増やさない
- 値は **px リテラル**。`rem` / `em` は使わない
- `html { font-size: 14px }` は Typography のために変更しない
- `body` 既定は **14** 段

### 2. Line height scale

| Pattern | CSS 変数 | 値（無単位） |
|---------|----------|--------------|
| tight | `--line-height-tight` | 1.25 |
| normal | `--line-height-normal` | 1.5 |
| relaxed | `--line-height-relaxed` | 1.75 |

`body` 既定は `--line-height-normal`。

### 3. Font weight scale

| 値 | CSS 変数 |
|----|----------|
| 400 | `--font-weight-400` |
| 500 | `--font-weight-500` |
| 600 | `--font-weight-600` |
| 700 | `--font-weight-700` |

`body` 既定は **400**。非標準値（`450` / `550` 等）は **Typography adoption** で最寄り段階へ四捨五入。

### 4. Font family

| CSS 変数 | 値 |
|----------|-----|
| `--font-family-ui` | `"Segoe UI", -apple-system, BlinkMacSystemFont, sans-serif` |

**Code font**（モノスペース）は v1 スコープ外。Web フォント追加も v1 外。

### 5. Text style catalog

共有クラス（`style.css`）。色は含めない（**Neutral text token** は Color 側）。

| Style | font-size | line-height | font-weight | クラス |
|-------|-----------|-------------|-------------|--------|
| Heading 1 | `--font-size-h1` | tight | 600 | `.text-h1`（`.page-title` は同型 alias + margin + 色。line-height は継承） |
| Heading 2 | `--font-size-18` | tight | 600 | `.text-h2` |
| Heading 3 | `--font-size-16` | tight | 600 | `.text-h3` |
| Heading 4 | `--font-size-14` | normal | 600 | `.text-h4` |
| Body | `--font-size-14` | normal | 400 | `.text-body` |
| Body small | `--font-size-12` | normal | 400 | `.text-body-sm` |
| Caption | `--font-size-10` | normal | 400 | `.text-caption` |

### 6. Element Plus typography mapping

`html.dark` 内で **Font size token** へ委譲:

| EP 変数 | 委譲先 |
|---------|--------|
| `--el-font-size-extra-large` | `var(--font-size-20)` |
| `--el-font-size-large` | `var(--font-size-18)` |
| `--el-font-size-medium` | `var(--font-size-16)` |
| `--el-font-size-base` | `var(--font-size-14)` |
| `--el-font-size-extra-small` | `var(--font-size-12)` |

`--el-font-size-small: 13px` は **Element Plus typography derivative** として mapping 内に残す（Typography scale 外・見た目不変）。

### 7. Typography v1 visual policy

tokenize 時は **既存の見た目を変えない**。Heading 1 は `--font-size-h1`（19.6px）を維持。`.page-title` は line-height を継承（従来どおり 1.5）。タイポグラフィのリデザインは v1 スコープ外。

### 8. 実装正本

- トークン定義・Text style・EP mapping: `frontend/src/assets/style.css`
- カタログ定数（Storybook・テスト用）: `frontend/src/design/typographyTokens.ts`
- 見た目と使い方: **Typography Storybook catalog**（`frontend/src/design/Typography.stories.ts`）
- Agent 規約: `.cursor/rules/typography-design-system.mdc`

### 9. Typography adoption

- 新規・改修で触ったファイルでは **Typography token** または **Text style catalog** を使う
- `style.css` の共有スタイル（`body`、`.page-title`、`.section-card__toggle` 等）は v1 でトークンへ寄せる
- 既存 View の `rem` / 任意 px は一括置換しない
- **Typography migration rounding**: スケール外の値は四捨五入で最寄りトークンへ（v1 deliverables では寄せない）
- TitleBar クローム・**Code font** は v1 スコープ外
- Element Plus 内部タイポの全面上書きは v1 スコープ外
- ESLint 強制は v1 では入れない

## Considered options

| 案 | 却下理由 |
|----|----------|
| `rem` トークン + `html` を 16px に変更 | 既存 UI 全体のタイポグラフィが変わり、Spacing ADR の前提も崩れる |
| px スケールに 11 / 13 を含める | 段数が増え、EP derivative との境界が曖昧になる |
| Heading 1 を `--font-size-20` 固定 | 19.6px → 20px で見た目が変わる（v1 visual policy と矛盾） |
| Text style に色を含める | **Color** の Neutral text と二重管理になる |
| EP `--el-font-size-small` を 12px に寄せる | EP コンポーネントの見た目が変わる |
| 全 View 一括置換 | Spacing / Color adoption と方針が矛盾しリグレッションリスクが大きい |
| Noto Sans JP 等の Web フォント追加 | v1 visual policy（見た目不変）と矛盾 |

## v1 scope

**含む:**

- `style.css` へのトークン + Text style 定義、共有クラスのトークン化、EP typography mapping
- `typographyTokens.ts` + 単体テスト
- Typography Storybook catalog
- `.cursor/rules/typography-design-system.mdc`
- `CONTEXT.md` / 本 ADR

**含めない:**

- 全画面の一括トークン化
- TitleBar クロームのトークン化
- **Code font** カタログ
- Element Plus 内部タイポの全面上書き
- ライトテーマ・テーマ切替
- 見た目のリデザイン（H1 を 20px 固定へ寄せる等）
- ESLint 任意サイズ禁止

## Consequences

- 新規・改修 PR では `var(--font-size-*)` / **Text style** クラスが期待される
- **Typography Storybook catalog** がタイポの正本となり、Spacing / Color / VtButton catalog と並べて参照できる
- Heading 1 の `calc` は将来 `--font-size-20` へ寄せるリデザイン PR の入口を残す
- v1 マージ直後の視覚差は意図しない（token 値は現行 computed と同等）
