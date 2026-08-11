# ADR 0022: Feedback & badges design system (VtAlert / VtTag / showToast)

## Status

Accepted（grill-with-docs で合意）

## Context

- `el-alert` / `el-tag` / `ElMessage` の直書きが画面ごとに散在し、**Section alert**（持続）と **Toast**（操作直後）の使い分けがばらつく
- **Color** ADR（[0019](0019-color-design-system.md)）で **Semantic color catalog** は定義済みだが、フィードバックチャネルの配置ルールとラッパーは未定義
- **Form control** ADR（[0021](0021-form-control-design-system.md)）はフィールド文脈のエラー 3 層を持つが、Alert / Toast / Badge / Loading の横断正本は別に必要
- `ElNotification` は未使用。`v-loading` / スケルトンも未使用
- **Button loading state**（[ADR 0017](0017-button-design-system-vtbutton.md)）は操作待ちの正本として既にある

用語は [`CONTEXT.md`](../../CONTEXT.md) の **Design system** 節（**Feedback & Badges**、**Section alert**、**Toast**、**Status badge**、**Loading feedback** 等）を正本とする。

## Decision

### 1. 責務

- **Feedback & Badges** = 横断 **Feedback channel** の正本（Alert / Toast / Badge / Loading）
- Form ADR のエラー 3 層は維持し、チャネル選択だけ本 ADR で結線する

### 2. チャネル選択（Feedback channel selection）

| 状況 | チャネル |
|------|----------|
| 解消まで残す状態・常時警告・ブロック内失敗・fetch failure | **Section alert**（**VtAlert**） |
| ユーザー明示操作の直後（保存・起動・反映の成否） | **Toast**（**showToast**） |
| 短い状態・結果・属性ラベル | **Status badge**（**VtTag**；**Domain badge** は例外） |
| 操作ボタン待ち | **Button loading state**（ADR 0017） |
| ブロック初回取得待ち | **Section loading**（中立テキスト 1 行） |

**主軸**: 持続性（残す → alert、操作直後 → toast）。**補足**: 文脈の近さ（ブロック内 → alert）。

#### Form ADR との結線

| Form 用語 | チャネル |
|-----------|----------|
| Form field validation error | `el-form-item`（Feedback 外） |
| Form save failure feedback | **Toast**（`showToast.error`） |
| Immediate toggle failure feedback | **Section alert**（**VtAlert**） |
| Form fetch failure | **Section alert** 型（Toast 禁止） |

### 3. VtAlert

- 新規: `frontend/src/components/VtAlert.vue`
- **`variant` は必須**: `success` | `warning` | `danger` | `info`
- `danger` → EP `type="error"`（**Semantic color catalog**）
- 既定: `show-icon`、`:closable="false"`
- `title` / `description` / `data-testid` 等は透過
- **VtAlert adoption**（触ったところから `el-alert` 直書きを置換）

### 4. VtTag

- 新規: `frontend/src/components/VtTag.vue`
- **`variant` は必須**: `success` | `warning` | `danger` | `info` | `neutral` | `primary`
- `primary` は **Launch profile default label** のみ（Brand mapping）
- **Domain badge**（VrcStatusTag / VrcUserTagChip）は対象外
- **VtTag adoption**（触ったところから `el-tag` 直書きを置換）

### 5. showToast

- 新規: `frontend/src/utils/showToast.ts`
- `showToast.success` / `.warning` / `.error` / `.info` — 引数は **i18n 済み string のみ**
- 内部は `ElMessage` に委譲。例外は呼び出し側で `formatError` 等してから渡す
- **VtToast**（DOM ラッパー）は v1 では作らない
- **showToast adoption**（触ったところから `ElMessage.*` 直叩きを置換）

### 6. Loading feedback（v1）

- **Action loading**: ADR 0017 を正本（変更なし）
- **Section loading**: ブロック内中立テキスト（`common.loading` 等）。`v-loading` / スケルトンは v1 外

### 7. Storybook

- `VtAlert.stories.ts` — 4 variant + description + ブロック error / 常時 warning 例
- `VtTag.stories.ts` — 6 variant + semantic 行 + default label 例

### 8. Agent / 実装ルール

- `.cursor/rules/feedback-design-system.mdc` 新設
- ESLint による `el-alert` / `el-tag` / `ElMessage` 禁止は v1 では入れない

## Considered options

| 案 | 却下理由 |
|----|----------|
| Form ADR に統合 | フォーム外の Gallery / Dashboard / Cookie 警告が Form 文脈に収まらない |
| `ElNotification` 採用 | 未使用。Toast と役割が重複 |
| `v-loading` 標準化 | 現行 UI に無く、YAGNI |
| VtToast コンポーネント | Toast はグローバル API。薄い util で十分 |
| showToast + errorFromUnknown | 魔法 API。呼び出し側で format して明示する方がレビューしやすい |
| v1 で全面置換 | Button / Form と同型の adoption で十分 |

## v1 scope

**含む:**

- VtAlert / VtTag / showToast + 単体テスト
- VtAlert / VtTag Storybook catalog
- `.cursor/rules/feedback-design-system.mdc`
- 本 ADR + `CONTEXT.md`

**含まない:**

- `ElNotification`、OS 通知、`v-loading`、スケルトン、進捗％
- 全 View 一括置換、ESLint 強制
- Dashboard refresh 警告・Automation トグル失敗の v1 強制移行
- **Domain badge** の変更

## Consequences

- 新規・改修 PR では VtAlert / VtTag / showToast とチャネル選択表が期待される
- Form ADR のエラー表は Toast / VtAlert へリンク更新する（実装は触ったファイルから）
- 既存 `el-alert` / `el-tag` / `ElMessage` は adoption まで併存可
