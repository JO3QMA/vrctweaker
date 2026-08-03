# ADR 0018: Spacing design system

## Status

Accepted（grill-with-docs で合意）

## Context

- 画面ごとに `margin` / `padding` / `gap` が `rem`（14px ルート）と任意 `px` で混在し、7px・10.5px・14px・21px など **Spacing scale に乗らない値**が散見される
- **VtButton**（[ADR 0017](0017-button-design-system-vtbutton.md)）は整備済みだが、余白ルールは未定義
- `html { font-size: 14px }` のため `rem` ベースの余白トークンは 8 の倍数スケールと噛み合わない
- Element Plus コンポーネント内部の余白はフレームワーク既定のまま触らない方が安全（全面上書きの影響範囲が大きい）

用語は [`CONTEXT.md`](../../CONTEXT.md) の **Design system** セクション（**Spacing**、**Spacing scale**、**Spacing pattern catalog** 等）を正本とする。

## Decision

### 1. Spacing scale

| px | CSS 変数 |
|----|----------|
| 4 | `--space-4` |
| 8 | `--space-8` |
| 12 | `--space-12` |
| 16 | `--space-16` |
| 24 | `--space-24` |
| 32 | `--space-32` |
| 48 | `--space-48` |
| 64 | `--space-64` |

- 上記以外の任意余白は新規・改修で増やさない
- 値は **px リテラル**。`rem` / `em` は使わない
- 14px ルートの `font-size` は Spacing のために変更しない

### 2. Spacing pattern catalog

| Pattern | 委譲先 | 用途 |
|---------|--------|------|
| `--space-inline-tight` | `--space-4` | チップ横・バッジ内・インライン詰め |
| `--space-action-group` | `--space-8` | **Action group** 内 `gap`、ツールバー内ボタン間 |
| `--space-form-field` | `--space-12` | フォーム項目間・ラベルと入力の縦間隔 |
| `--space-block` | `--space-16` | 見出し下・段落間・カード内小ブロック間 |
| `--space-section` | `--space-24` | セクションカード間・画面内の大きな区切り |
| `--space-page` | `--space-32` | ページ外周・ヒーロー上下など広い余白 |

48 / 64 は数値トークンのみ（v1 では pattern 名なし）。

### 3. 実装正本

- トークン定義: `frontend/src/assets/style.css` の `:root`
- カタログ定数（Storybook・テスト用）: `frontend/src/design/spacingTokens.ts`
- 見た目と使い方: **Spacing Storybook catalog**（`frontend/src/design/Spacing.stories.ts`）
- Agent 規約: `.cursor/rules/spacing-design-system.mdc`

### 4. Spacing adoption

- 新規・改修で触ったファイルでは **Spacing token** または **Spacing pattern** を使う
- `style.css` の共有クラス（`.page-title`、`.section-card`、`.section-card__toggle` 等）は v1 でトークンへ寄せる
- 既存 View の `rem` / 任意 px は一括置換しない
- **Spacing migration rounding**: スケール外の値は四捨五入で最寄りトークンへ（例: 14px → 16、21px → 24）
- Sizing 寸法・意図的な負の margin は対象外
- Element Plus 内部余白の上書きは v1 スコープ外
- ESLint 強制は v1 では入れない

### 5. スコープ外（v1）

- **Border radius**（既存 `--radius` はそのまま）
- Sizing（アイコン・アバター寸法など）
- 全 View 一括トークン化
- ユーティリティクラス（`.gap-8` 等）

## Considered options

| 案 | 却下理由 |
|----|----------|
| `rem` トークン + `html` を 16px に変更 | 既存 UI 全体のタイポグラフィが変わり、影響範囲が大きい |
| 8 の倍数のみ（4px 禁止） | チップ横など既存の 4px 用途と衝突 |
| ユーティリティクラス体系 | 導入コストが高く、Vue SFC の scoped style と二重管理になりやすい |
| 全 View 一括置換 | リグレッションリスクと PR スコープが大きい |

## v1 scope

**含む:**

- `style.css` へのトークン + pattern 定義、共有クラスのトークン化
- `spacingTokens.ts` + 単体テスト
- Spacing Storybook catalog
- `.cursor/rules/spacing-design-system.mdc`
- `CONTEXT.md` / 本 ADR

**含まない:**

- 全画面の一括トークン化
- Element Plus 内部余白の上書き
- ESLint 任意余白禁止
- 48 / 64 の pattern 名付け

## Consequences

- 新規・改修 PR では `var(--space-*)` または pattern 変数が期待される
- 共有クラス経由の画面は v1 マージ後に数 px の視覚差が出うる（四捨五入による意図的な寄せ）
- Storybook が Spacing の正本となり、VtButton catalog と並べて参照できる
