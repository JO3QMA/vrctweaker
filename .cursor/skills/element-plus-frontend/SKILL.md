---
name: element-plus-frontend
description: >-
  VRChat Tweaker の frontend で Element Plus を使う UI 実装、Vitest/Playwright のセレクタ、
  アイコン利用、公式ドキュメントの参照タイミングを扱う。el-* コンポーネント追加・E2E 失敗調査・
  フォーム・ダイアログ実装時に読む。UI 変更・抽象化時は Storybook（.cursor/rules/storybook-wails-ui.mdc）も更新する。
---

# Element Plus フロントエンド

## いつ読むか

- 新規画面やコンポーネントで **ボタン／フォーム／テーブル／フィードバック** を足すとき。
- **E2E / 単体テスト** で Element Plus 由来の DOM が原因でセレクタが合わないとき。
- **API・props・スロット** が曖昧なとき（推測で独自ラッパーを増やさない）。
- **UI の変更・共通コンポーネント化** をするとき（**Storybook への反映**が必須。ルールは `.cursor/rules/storybook-wails-ui.mdc`）。

## 手順

1. **プロジェクトルール** `.cursor/rules/element-plus-ui.mdc` を前提にする（グローバル登録・ダーク CSS・`.el-*` DOM）。
2. **Storybook / Wails**: WebView のデザインは Storybook と `wails dev` で扱う。**UI 変更や抽象化**をしたら `.cursor/rules/storybook-wails-ui.mdc` に従い、**`*.stories.ts` の追加・更新**や Wails 用**デコレータ**を揃える。
3. **公式ドキュメント** [element-plus.org](https://element-plus.org/) で該当コンポーネントの props / events / slots を確認する（依存バージョンは `frontend/package.json` の `element-plus` に合わせる）。
4. 実装後、テストでは次を意識する:
   - **Vitest + Vue Test Utils**: 必要なら `el-*` をスタブする。表示テキストや `data-testid` でアサートする。
   - **Playwright**: `section` + `h2` 前提を避け、`el-card` ヘッダーや `.setting-row` など **実際のマークアップ**に合わせる。`el-input-number` は `input[type="number"]` 固定としない。

## アイコン

- `@element-plus/icons-vue` は `main.ts` で全コンポーネント登録済み。テンプレートでは `<Plus />` のように **PascalCase タグ**で使う（名前は [Icons 一覧](https://element-plus.org/en-US/component/icon.html) に従う）。

## 既存パターン

- ランチャー・設定などでは `el-form` / `el-input` / `el-button` / `el-tag` / `el-card` が多用されている。新規でも同密度の **日本語ラベル・レイアウト**に揃える。

## 関連

- Vue 一般: `.cursor/rules/vue-conventions.mdc`
- Storybook と Wails WebView: `.cursor/rules/storybook-wails-ui.mdc`
- 変更後検証: `.cursor/rules/tdd-workflow.mdc`（`make fmt` / `make test` / `make lint` に加え、`frontend/src` を変えたら **`make test-e2e`**。インタラクション検証は可能なら **`data-testid`** で i18n に依存させない）
