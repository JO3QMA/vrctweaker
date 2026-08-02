# ADR 0017: Button design system (VtButton)

## Status

Accepted（grill-with-docs で合意）

## Context

- 画面ごとに `el-button` の `type` / `plain` / `text` の使い方がばらつき、主操作・破壊操作・補助操作の優先度が読み取りにくい
- Element Plus は `primary` / `danger` / `text` 等の見た目 API を持つが、**意味論（いつどれを使うか）**はフレームワークが規定しない
- ダークテーマの中立塗り・disabled トークンは `frontend/src/assets/style.css` で既に定義済み
- **Presence change section** のプレゼンス色ボタンなど、色そのものがドメイン意味を持つ UI は 4 様式の対象外にすべき

用語は [`CONTEXT.md`](../../CONTEXT.md) の **Design system** セクション（**Primary button**、**VtButton**、**Semantic button** 等）を正本とする。

## Decision

### 1. 4 様式 + 状態

| 概念 | 役割 |
|------|------|
| **Primary** | Action group 内の主操作（**高々 1 つ**） |
| **Secondary** | 実行操作（起動・更新・複製・参照・編集を破棄しないキャンセル） |
| **Tertiary** | 補助・低優先（再読み込み、パネル閉じ、**Draft row removal**） |
| **Danger** | **永続データ**の破壊操作。破壊確認ダイアログでは確認 = Danger、キャンセル = Secondary。**Primary は置かない** |
| **disabled** | 様式ではない状態。標準の `disabled`（または同等） |
| **loading** | 押したボタンだけに付ける一時状態（実質 disabled） |

### 2. VtButton コンポーネント

- 新規: `frontend/src/components/VtButton.vue`
- **`variant` は必須**: `'primary' | 'secondary' | 'tertiary' | 'danger'`（暗黙既定なし）
- 内部は `el-button` をラップ。`disabled`・`loading`・`size`・`data-testid` 等は **透過**（`$attrs` / `inheritAttrs`）
- 画面テンプレートでは **VtButton を使う**。`el-button` 直書きは **VtButton adoption**（触ったところから順次）まで併存可
- **ESLint による `el-button` 禁止は v1 では入れない**。遵守は `.cursor/rules/` で案内

### 3. Element Plus マッピング（実装正本）

| `variant` | `el-button` 相当 |
|-----------|------------------|
| `primary` | `type="primary"` |
| `secondary` | 無指定（中立色塗り。`style.css` の既定） |
| `tertiary` | `text` |
| `danger` | `type="danger"` |
| `danger`（弱い強調） | `type="danger"` + `plain` — **同一 Action group に Primary があるときのみ**（例: Settings のログアウト） |

`VtButton` は上記以外の `type`（`success` / `warning` / `info`）を **公開しない**。

### 4. 例外

| 種別 | 扱い |
|------|------|
| **Semantic button** | VtButton の variant に含めない。専用 UI（例: Presence 色ボタン）。disabled と a11y ラベルは適用 |
| **Dialog confirm button** | `ElMessageBox` 等は VtButton 非対応。class 定数で Danger / Secondary 見た目を共有（下記） |
| お気に入り★・`+` アイコン追加 | Semantic 例外にせず **Tertiary** または **Secondary** |

### 5. MessageBox class 定数

`VtButton.vue` と同ディレクトリ（または隣接モジュール）に export する:

- `VT_BUTTON_DANGER_CONFIRM_CLASS` → 破壊確認の OK（`el-button--danger` 相当）
- `VT_BUTTON_SECONDARY_CANCEL_CLASS` → キャンセル（中立塗り相当。必要なら `plain` 併用をここで固定）

既存の `confirmButtonClass: "el-button--danger"` は定数へ寄せる（全面置換は adoption 方針に従い触ったファイルから）。

### 6. Storybook

- `frontend/src/components/VtButton.stories.ts`
- **VtButton Storybook catalog**: 4 variant ×（通常 / disabled / loading）+ Action group 並び例ストーリー
- 見た目の正本は Storybook（`.cursor/rules/storybook-wails-ui.mdc` に従う）

### 7. Agent / 実装ルール

- `.cursor/rules/` にボタン規約を追記（`element-plus-ui.mdc` 拡張または `button-design-system.mdc` 新設）
- マッピング表・例外・MessageBox 定数・adoption 方針を記載

## Considered options

| 案 | 却下理由 |
|----|----------|
| `el-button` 規約のみ（ラッパーなし） | variant 必須・マッピングの強制力が弱く、画面追加のたびに `type` / `plain` / `text` の誤用が再発しやすい |
| 一括 `el-button` → VtButton 移行 | リグレッションリスクと PR スコープが大きい |
| 移行期から ESLint 禁止 | 未移行の既存 `el-button` と矛盾し CI が常時赤になる |

## v1 scope

**含む:**

- `VtButton.vue` + 単体テスト + Storybook catalog
- MessageBox 用 class 定数
- `.cursor/rules/` 更新
- `CONTEXT.md` / 本 ADR

**含まない:**

- 全画面の一括置換
- ESLint `el-button` 禁止
- `VtConfirmDialog` による MessageBox 置換
- `success` / `warning` / `info` variant の追加
- Semantic button の VtButton 統合

## Consequences

- 新規・改修 PR では VtButton + 必須 `variant` が期待される
- 破壊確認は引き続き `ElMessageBox` だが、class は定数経由で VtButton と揃える
- Presence 色など Semantic UI は現行コンポーネントを維持
