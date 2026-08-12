# ADR 0023: Iconography design system

## Status

Accepted（grill-with-docs で合意）

## Context

- **UI icon** が `<el-icon>` + 任意 `font-size`（14px・`1rem`・`0.9rem` 等）で混在し、**Icon size scale** が未定義
- **Navigation glyph**（サイドバー絵文字）と **Domain icon**（Encounter friend mark 等）が同じ寸法・色の規律を共有していない
- **Spacing**（[ADR 0018](0018-spacing-design-system.md)）・**Typography**（[ADR 0020](0020-typography-design-system.md)）はアイコン寸法をスコープ外とした
- **Color**（[ADR 0019](0019-color-design-system.md)）は整備済みだが、アイコンへの **Neutral text token** / `currentColor` の適用ルールが未定義
- カスタム SVG の追加手順・標準ラッパーが無い

用語は [`CONTEXT.md`](../../CONTEXT.md) の **Design system** セクション（**Iconography**、**UI icon**、**Navigation glyph**、**Domain icon**、**VtIcon** 等）を正本とする。

## Decision

### 1. スコープ

| 含む | 含まない |
|------|----------|
| **UI icon**（`@element-plus/icons-vue`） | アバター・サムネ |
| **Navigation glyph**（サイドバー絵文字。v1 はサイズ規律のみ） | Server status 色ドット |
| **Domain icon**（friend mark 等。色はドメイン正本） | **Semantic button**（Presence 色ボタン） |
| **Custom icon** 手順（v1 は Storybook サンプルのみ） | `main.ts` での EP アイコン全登録 |

### 2. Icon size scale

| px | CSS 変数 |
|----|----------|
| 12 | `--icon-size-12` |
| 16 | `--icon-size-16` |
| 20 | `--icon-size-20` |
| 24 | `--icon-size-24` |

**Icon size pattern**:

| Pattern | CSS 変数 | 委譲先 |
|---------|----------|--------|
| compact | `--icon-size-compact` | `--icon-size-12` |
| default | `--icon-size-default` | `--icon-size-16` |
| emphasis | `--icon-size-emphasis` | `--icon-size-20` |
| large | `--icon-size-large` | `--icon-size-24` |

- 上記以外の任意 px は新規・改修で増やさない
- 値は **px リテラル**。`rem` / `em` は使わない
- **Typography scale** / **Spacing scale** とは別カテゴリ（数値が近くても流用しない）

**Icon size legacy derivative**（v1 visual policy・スケール外）:

| 用途 | CSS 変数 | 値 | adoption 先 |
|------|----------|-----|-------------|
| Section card toggle | `--icon-size-legacy-toggle` | `var(--font-size-14)`（14px） | `--icon-size-default`（16px） |

### 3. Icon color rule

- SVG / Element Plus アイコンは原則 **`currentColor`**（親の `color` を継承）
- 親は **Neutral text token**（`--color-text-primary` / `--color-text-secondary` / `--color-text-muted`）
- 単体装飾 **UI icon** は **secondary**、無効文脈は **muted**
- **Semantic color**・**Domain color** を汎用 **UI icon** に直付けしない
- **Navigation glyph**（絵文字期）は CSS `color` が効かない。**Icon size scale** のみ揃える
- **Domain icon** の色は各ドメイン正本（例: friend mark → `--color-text-secondary`）

### 4. Icon accessibility rule

| 種別 | ルール |
|------|--------|
| 装飾 **UI icon** | `aria-hidden="true"`（**VtIcon** 既定 `decorative=true`） |
| 意味 **Domain icon** | `role="img"` + `aria-label`（i18n）。**VtIcon** は使わず既存パターン維持 |
| アイコンのみボタン | v1 では新規禁止。既存は `aria-label` 必須 |

v1 では文書化と **VtIcon** 既定のみ。既存 View の一括 a11y 修正は **Iconography adoption** に委ねる。

### 5. VtIcon

- **UI icon** と **Custom icon** の標準ラッパー。内部は `el-icon`
- `size`: `compact` \| `default` \| `emphasis` \| `large`（**必須**、暗黙既定なし）
- `decorative`: 既定 `true` → `aria-hidden="true"`
- `data-testid` 等は透過（`inheritAttrs: false` + `v-bind="$attrs"`）
- **Navigation glyph** の絵文字は v1 では `VtIcon` で包まない。共有クラス `.nav-glyph-size-default` 等で **Icon size token** を当てる

### 6. Custom icon component

1. `frontend/src/icons/FooIcon.vue` を追加
2. 単一ルート `<svg>`、`viewBox` 必須、`width`/`height` なし
3. `fill="currentColor"` または `stroke="currentColor"`
4. 利用: `<VtIcon size="default"><FooIcon /></VtIcon>`

v1 deliverables では `SampleIcon.vue`（Storybook 用）のみ。

### 7. Iconography v1 visual policy

tokenize 時は **見た目を変えない**。14px の toggle は **Icon size legacy derivative** を維持。**Iconography migration rounding**（14→16 等）は adoption 時のみ。サイドバー絵文字の SVG 化は v1 外。

### 8. 実装正本

- トークン・共有クラス: `frontend/src/assets/style.css`
- カタログ定数: `frontend/src/design/iconTokens.ts`
- ラッパー: `frontend/src/components/VtIcon.vue`
- 見た目と使い方: **Iconography Storybook catalog**（`frontend/src/design/Iconography.stories.ts`）
- Agent 規約: `.cursor/rules/iconography-design-system.mdc`

### 9. Iconography adoption

- 新規・改修で **VtIcon** + **Icon size token** + **Icon color rule** を使う
- `style.css` の共有スタイル（`.section-card__toggle-icon` 等）は v1 で icon トークンへ寄せる（legacy derivative 可）
- 既存 View の `<el-icon>` 直書き・絵文字サイズ直書きは一括置換しない
- **Iconography migration rounding**: スケール外は四捨五入で最寄りトークンへ（v1 deliverables では寄せない）
- ESLint 強制は v1 では入れない

## Considered options

| 案 | 却下理由 |
|----|----------|
| Typography `--font-size-*` をアイコンに流用 | Sizing カテゴリが混ざり、14px 本文とアイコン 16px の役割が曖昧になる |
| 16 / 20 / 24 のみ（12 なし） | friend mark 12px スロットを 16 に上げると行密度が落ちる |
| ナビ active を **Brand color** 固定 | サイドバー左ボーダーと二重強調になりうる |
| `main.ts` で EP アイコン全登録 | v1 スコープ肥大。現行はファイル単位 import で十分 |
| v1 で絵文字を SVG 一括置換 | 見た目・選定コストが大きく visual policy と矛盾 |
| アイコン専用 hex パレット | **App color token** と二重管理になる |
| 全 View 一括 VtIcon 化 | 他 adoption と方針が矛盾しリグレッションリスクが大きい |

## v1 scope

**含む:**

- `style.css` への **Icon size token**・pattern・legacy derivative・`.vt-icon--size-*`・`.nav-glyph-size-*`
- `iconTokens.ts` + 単体テスト
- **VtIcon** + 単体テスト
- **Iconography Storybook catalog**（EP 代表 + SampleIcon + 絵文字サイズ例）
- 共有スタイルのトークン参照化（`.section-card__toggle-icon`）
- `.cursor/rules/iconography-design-system.mdc`
- `CONTEXT.md` / 本 ADR

**含まない:**

- サイドバー絵文字 → SVG / EP 置換
- 各 View の `<el-icon>` 一括 VtIcon 化
- **Custom icon** の製品画面への追加（Storybook サンプル除く）
- `main.ts` EP グローバル登録
- 最小タップ領域 44px 等の新規 a11y 規定
- ESLint 強制

## Consequences

- 新規・改修 PR では **VtIcon** + `var(--icon-size-*)` が期待される
- **Iconography Storybook catalog** がアイコンの正本となり、Spacing / Color / Typography catalog と並べて参照できる
- 14px legacy toggle は adoption で 16px へ寄せる入口が残る
- v1 マージ直後の視覚差は意図しない（legacy derivative は現行 computed と同等）
- 絵文字ナビはサイズクラスで揃えつつ、SVG 化は後続 Issue で **Navigation glyph** として実施できる
