# ADR 0021: Form & input control design system (VtInput / VtSelect / VtCheckbox / VtSwitch)

## Status

Accepted（grill-with-docs で合意）

## Context

- 画面ごとに `el-input` / `el-select` / `el-checkbox` / `el-switch` の直書きが散在し、レイアウト（`el-form` top vs `setting-row`）、checkbox/switch の意味論、エラー表示がばらつく
- **VtButton**（[ADR 0017](0017-button-design-system-vtbutton.md)）、**Spacing**（[ADR 0018](0018-spacing-design-system.md)）、**Color**（[ADR 0019](0019-color-design-system.md)）、**Typography**（[ADR 0020](0020-typography-design-system.md)）は整備済み。入力コントロールの色は `style.css` の `--el-input-*` / `--el-switch-*` で既にマッピング済みだが、コンポーネント規約と adoption は未定義
- Element Plus は見た目 API を持つが、**checkbox vs switch**・**即時反映 vs 一括保存**の意味論は規定しない
- パス入力は `.path-input-group` + `el-input` が実装主流だが、`path-input-pattern.mdc` の例は素 `<input>` のまま

用語は [`CONTEXT.md`](../../CONTEXT.md) の **Design system** セクション（**Form control**、**Form layout**、**Form control checkbox–switch rule** 等）を正本とする。

## Decision

### 1. ラッパー（Vt*）

| ラッパー | 内部 | 用途 |
|----------|------|------|
| **VtInput** | `el-input` | テキスト系（`type` 透過）。`textarea` は v1 ではラッパーなし |
| **VtSelect** | `el-select` + `el-option` | 単一選択。filterable / multiple / remote は v1 外 |
| **VtCheckbox** | `el-checkbox` | **Form layout** 内・他フィールドと同じ保存タイミングの真偽 |
| **VtSwitch** | `el-switch` | **Immediate setting toggle** / **List enable toggle** |

- `disabled`・`data-testid`・`size` 等は **透過**（`$attrs` / `inheritAttrs`、VtButton と同型）
- v1 では暗黙の `variant` は不要（ボタンと異なり 4 様式ではない）
- 新規・改修では Vt* を使う。**Form control adoption**（触ったところから順次）

### 2. Element Plus マッピング

- 色: 既存 `--el-input-*` / `--el-switch-*`（**Color** ADR）を維持。v1 では見た目変更なし
- サイズ: **`el-form size="default"`** で子へ継承。Vt* に毎回 `size` は付けない
- **Form control size**: 省略時 `default`。`small` は **List enable toggle** の VtSwitch と密集レイアウト（Launcher Advanced 等）の明示指定のみ。`large` は v1 不使用

### 3. レイアウト

| パターン | 正本 | 用途 |
|----------|------|------|
| **Form layout** | `el-form` + `label-position="top"` | 複数項目の編集フォーム。項目間 `--space-form-field` |
| **Setting row** | `.setting-row` | カード内 1 行（ラベル＋hint 左、コントロール右） |
| **Path input row** | `.path-input-group` + VtInput + VtButton Secondary | パス + 参照（複数アクション可）。**VtPathInput** は v1 では作らない |

### 4. Checkbox vs switch（**Form control checkbox–switch rule**）

| 文脈 | コントロール |
|------|--------------|
| 編集フォーム内の真偽（他項目と同じ保存・反映） | **VtCheckbox** |
| 即時反映の単発設定 | **VtSwitch** + 多くは **Setting row** |
| 一覧・カードの有効フラグ即切替 | **VtSwitch**（`small` 可） |
| UI 表示モード切替（Friends Online/Offline 等） | **UI mode toggle**（Semantic 例外。toggle-group 等） |
| 複数選択（曜日など） | **VtCheckbox** + `el-checkbox-group` |

- **VtSwitch** の `active-text` / `inline-prompt` 等の内蔵ラベルは v1 不使用

### 5. disabled / readonly

- **Form control disabled state**: 標準 `disabled` 透過。理由は周辺 hint / 文言
- `readonly` は v1 の Form control では使わない（`disabled` に統一）
- VtInput / VtSelect に loading 様式は設けない

### 6. エラー表示（3 層）

| 層 | 示し方 |
|----|--------|
| **Form field validation error** | `el-form-item` 直下、i18n 固定文のみ |
| **Form save failure feedback** | `ElMessage.error`（一括保存失敗） |
| **Immediate toggle failure feedback** | ブロック内 `el-alert` 等（Cookie linkage 型）。一覧トグルも同型へ寄せる |
| **Form fetch failure** | カード内メッセージ＋disabled / 非表示（Presence 型）。別契約 |

### 7. Storybook

- `VtInput.stories.ts` / `VtSelect.stories.ts` / `VtCheckbox.stories.ts` / `VtSwitch.stories.ts` — 各コントロールの通常 / disabled / 代表 props
- `FormLayoutPatterns.stories.ts` — Form layout / Setting row / Path input row
- 正本は **Form control Storybook catalog**（VtButton catalog と同型）

### 8. Agent / 実装ルール

- `.cursor/rules/form-design-system.mdc` 新設
- `path-input-pattern.mdc` を VtInput + VtButton に更新
- ESLint による `el-input` 等禁止は v1 では入れない

## Considered options

| 案 | 却下理由 |
|----|----------|
| 規約のみ（ラッパーなし） | adoption の強制力が弱く、checkbox/switch 誤用が再発しやすい |
| **VtPathInput** を v1 で統合 | Settings path-row の複数アクションで props が膨らむ |
| レイアウトを `el-form` に全面統一 | Setting row の横並び説明＋スイッチが縦に伸びる |
| 即時トグル失敗を ElMessage のみ | 一覧トグルと Setting row で表現がずれる |
| Storybook 統合 1 ファイルのみ | コンポーネント別の正本が薄れる |

## v1 scope

**含む:**

- 4 ラッパー + 単体テスト + Storybook（4 + Form layout patterns）
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

- 新規・改修 PR では Vt* と Form layout / Setting row / Path input row の使い分けが期待される
- 即時トグル失敗は触ったファイルからブロック内エラーへ寄せる（Automation 一覧の ElMessage は段階移行）
- UI モード切替（Friends 等）は現行パターンを維持
