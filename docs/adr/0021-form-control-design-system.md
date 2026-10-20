# ADR 0021: Form & input control design system

## Status

Accepted（grill-with-docs で合意。2026-10-20 改訂: 透過ラッパーは廃し、フォーム部品は `el-*` を直接使う）

## Context

- 画面ごとに `el-input` / `el-select` / `el-checkbox` / `el-switch` の直書きが散在し、レイアウト（`el-form` top vs `setting-row`）、checkbox/switch の意味論、エラー表示がばらつく
- **VtButton**（[ADR 0017](0017-button-design-system-vtbutton.md)）、**Spacing**（[ADR 0018](0018-spacing-design-system.md)）、**Color**（[ADR 0019](0019-color-design-system.md)）、**Typography**（[ADR 0020](0020-typography-design-system.md)）は整備済み。入力コントロールの色は `style.css` の `--el-input-*` / `--el-switch-*` で既にマッピング済みだが、コンポーネント規約と adoption は未定義
- Element Plus は見た目 API を持つが、**checkbox vs switch**・**即時反映 vs 一括保存**の意味論は規定しない
- パス入力は `.path-input-group` + `el-input` が実装主流だが、`path-input-pattern.mdc` の例は素 `<input>` のまま

用語は [`CONTEXT.md`](../../CONTEXT.md) の **Design system** セクション（**Form control**、**Form layout**、**Form control checkbox–switch rule** 等）を正本とする。

## Decision

### 1. フォーム部品は Element Plus を直接使う

| 用途 | コンポーネント |
|------|----------------|
| テキスト系（`type` 透過）。`textarea` は別ラッパーを作らない | `el-input` |
| 単一選択。filterable / multiple / remote は v1 外 | `el-select` + `el-option` |
| **Form layout** 内・他フィールドと同じ保存タイミングの真偽 | `el-checkbox` |
| **Immediate setting toggle** / **List enable toggle** | `el-switch` |

- 中身が `$attrs` だけのラッパー（旧 VtInput / VtSelect / VtCheckbox / VtSwitch）は置かない
- **variant の対応が要るものだけラップする**: **VtButton** / **VtAlert** / **VtTag** / **VtIcon**
- `disabled`・`data-testid`・`size` 等は Element Plus の props をそのまま渡す
- v1 では入力部品に暗黙の `variant` は不要（ボタンと異なり 4 様式ではない）
- 見た目は `style.css` の `html.dark` にある `--el-input-*` / `--el-switch-*` が `el-*` に効く。部品ごとの追加 class は要らない

### 2. Element Plus マッピング

- 色: 既存 `--el-input-*` / `--el-switch-*`（**Color** ADR）を維持。v1 では見た目変更なし
- サイズ: **`el-form size="default"`** で子へ継承。入力部品に毎回 `size` は付けない
- **Form control size**: 省略時 `default`。`small` は **List enable toggle** の `el-switch` と密集レイアウト（Launcher Advanced 等）の明示指定のみ。`large` は v1 不使用

### 3. レイアウト

| パターン | 正本 | 用途 |
|----------|------|------|
| **Form layout** | `el-form` + `label-position="top"` | 複数項目の編集フォーム。項目間 `--space-form-field` |
| **Setting row** | `.setting-row` | カード内 1 行（ラベル＋hint 左、コントロール右） |
| **Path input row** | `.path-input-group` + `el-input` + VtButton Secondary | パス + 参照（複数アクション可）。**VtPathInput** は v1 では作らない |

### 4. Checkbox vs switch（**Form control checkbox–switch rule**）

| 文脈 | コントロール |
|------|--------------|
| 編集フォーム内の真偽（他項目と同じ保存・反映） | `el-checkbox` |
| 即時反映の単発設定 | `el-switch` + 多くは **Setting row** |
| 一覧・カードの有効フラグ即切替 | `el-switch`（`small` 可） |
| UI 表示モード切替（Friends Online/Offline 等） | **UI mode toggle**（Semantic 例外。toggle-group 等） |
| 複数選択（曜日など） | `el-checkbox` + `el-checkbox-group` |

- `el-switch` の `active-text` / `inline-prompt` 等の内蔵ラベルは v1 不使用

### 5. disabled / readonly

- **Form control disabled state**: 標準 `disabled` 透過。理由は周辺 hint / 文言
- `readonly` は v1 の Form control では使わない（`disabled` に統一）
- `el-input` / `el-select` に loading 様式は設けない

### 6. エラー表示（3 層）

| 層 | 示し方 |
|----|--------|
| **Form field validation error** | `el-form-item` 直下、i18n 固定文のみ |
| **Form save failure feedback** | `ElMessage.error`（一括保存失敗） |
| **Immediate toggle failure feedback** | ブロック内 `el-alert` 等（Cookie linkage 型）。一覧トグルも同型へ寄せる |
| **Form fetch failure** | カード内メッセージ＋disabled / 非表示（Presence 型）。別契約 |

### 7. Storybook

- 部品ごとの専用ストーリーは置かない。`el-input` / `el-select` / `el-checkbox` / `el-switch` の見た目はグローバル CSS が担う
- `FormLayoutPatterns.stories.ts` — Form controls / Form layout / Setting row / Path input row
- レイアウトの見本は **Form control Storybook catalog**（`FormLayoutPatterns.stories.ts`）

### 8. Agent / 実装ルール

- `.cursor/rules/form-design-system.mdc` 新設
- `path-input-pattern.mdc` を `el-input` + VtButton にする
- ESLint による `el-input` 等禁止は v1 では入れない

## Considered options

| 案 | 却下理由 |
|----|----------|
| 透過ラッパー（`$attrs` のみ）を残す | variant 対応が無く、`el-*` 直書きと同じ DOM になる。2026-10-20 に廃止 |
| **VtPathInput** を v1 で統合 | Settings path-row の複数アクションで props が膨らむ |
| レイアウトを `el-form` に全面統一 | Setting row の横並び説明＋スイッチが縦に伸びる |
| 即時トグル失敗を ElMessage のみ | 一覧トグルと Setting row で表現がずれる |
| Storybook 統合 1 ファイルのみ | コンポーネント別の正本が薄れる |

## v1 scope

**含む:**

- `el-input` / `el-select` / `el-checkbox` / `el-switch` の直接利用
- Form layout patterns の Storybook
- `.cursor/rules/form-design-system.mdc`
- `path-input-pattern.mdc` 更新
- `CONTEXT.md` / 本 ADR

**含まない:**

- `el-input-number` / `el-radio` / `el-autocomplete` ラッパー
- `VtPathInput` / `VtFormItem` / `VtTextarea`
- 全 View 一括置換、ESLint 強制
- 全フォーム `:rules` 一括導入、サーバーフィールド別エラー自動マッピング
- `resolution-selection-ui` の変更
- Element Plus 入力色・見た目の変更

## Consequences

- 新規・改修 PR では `el-input` / `el-select` / `el-checkbox` / `el-switch` と、Form layout / Setting row / Path input row の使い分けが期待される。透過ラッパーは足さない
- 即時トグル失敗は触ったファイルからブロック内エラーへ寄せる（Automation 一覧の ElMessage は段階移行）
- UI モード切替（Friends 等）は現行パターンを維持
