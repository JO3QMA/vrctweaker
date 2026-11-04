# ダークモード（固定）

VRChat Tweaker は **ダークテーマ固定** です。ライトモード切り替えは仕様外（実装しない）。

## 必須

- `frontend/index.html` の `<html>` に **`class="dark"`** を付けたままにする。
- `frontend/src/main.ts` で Element Plus ダーク CSS 変数（`element-plus/theme-chalk/dark/css-vars.css`）を読み込む。

`class="dark"` を外すと Element Plus とアプリトークン（`style.css`）の見た目が崩れます。

## 関連

- 色トークン: `docs/adr/0019-color-design-system.md`
- Element Plus: `.cursor/rules/element-plus-ui.mdc`
