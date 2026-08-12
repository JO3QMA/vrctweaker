# ADR 0024: Bulk design-system adoption campaign

## Status

Accepted（Done）

## Context

- Design system の正本（[ADR 0017](0017-button-design-system-vtbutton.md)–[0023](0023-iconography-design-system.md)）と Vt* / トークン / Storybook は揃っている
- 各 ADR の adoption 方針は「新規・改修で触ったファイルから順次。一括置換しない」
- 既存 View / 共有 UI には `el-button` / Form 直書き / `ElMessage` / レガシー CSS が残っており、画面間の一貫性が低い
- 意図的に全画面・全層を一度に寄せるキャンペーンを行う

用語は [`CONTEXT.md`](../../CONTEXT.md) の **Design system** 節および本 ADR の **Bulk design-system adoption campaign** を正本とする。

## Decision

### 1. キャンペーン範囲

| 含む | 含まない |
|------|----------|
| 全 View と関連共有 UI コンポーネント | ESLint による `el-*` / `ElMessage` 禁止 |
| Button / Form / Feedback / Color / Spacing / Typography / Iconography の全層 | Sidebar 絵文字の SVG 化 |
| 画面グループ単位の複数 PR | レイアウトリデザイン・機能追加 |
| | Domain badge（`VrcStatusTag` / `VrcUserTagChip`）の VtTag 化 |
| | Semantic button（Presence 色ボタン）の VtButton 化 |
| | TitleBar ウィンドウクローム固有サイズの強制トークン化 |

本キャンペーンは ADR 0017–0023 の「一括置換しない」adoption 方針の **意図的な例外**である。通常の機能 PR では引き続き「触ったところから」を適用する。

### 2. 置換対応

| 現行 | 寄せ先 |
|------|--------|
| `el-button` | **VtButton**（`variant` 必須） |
| `el-input` / `el-select` / `el-checkbox` / `el-switch` | **VtInput** / **VtSelect** / **VtCheckbox** / **VtSwitch** |
| `el-alert` | **VtAlert** |
| 汎用 `el-tag` | **VtTag**（Domain badge 除く） |
| `ElMessage.*` | **showToast** |
| `ElMessageBox` の confirm/cancel class | `VT_BUTTON_DANGER_CONFIRM_CLASS` / `VT_BUTTON_SECONDARY_CANCEL_CLASS` |
| rem / hex / レガシー色変数 | `--space-*` / `--color-*` / `--font-*` / `--icon-*` |
| `<el-icon>` / 素のアイコン寸法 | **VtIcon** + icon size token |

Spacing / Typography / Icon のスケール外値は四捨五入で最寄り段階へ。Color は値を変えず参照名のみ。

### 3. PR 単位

1 PR = 対象画面（と直結コンポーネント）の **全層**適用。層ごと横断置換はしない。

### 4. 完了条件（ゲート）

対象外を除き、Views + 共有 UI で次が残っていないこと。

- テンプレートの `el-button` / `el-input` / `el-select` / `el-checkbox` / `el-switch` / `el-alert` / 汎用 `el-tag`
- 本番の `ElMessage.success|error|warning|info` 直叩き（`showToast` 実装・テスト除く）
- scoped のレガシー `--accent` / `--bg-primary` 等の参照（`:root` alias 定義は残してよい）
- スケール外の余白・font-size・icon 辺長の直書き（TitleBar クローム等の明示除外を除く）

完了後、本 ADR の Status を **Accepted（Done）** に更新する。

## Consequences

- レビュー負荷が高いため画面グループ分割が必須
- DOM は多くが `.el-*` のままなので E2E は比較的安定するが、コンポーネント名アサーションは要確認
- 通常 adoption 文言と矛盾しないよう、CONTEXT に本キャンペーン用語を置く

## Completion notes（Done）

意図的に残した例外（ゲート対象外）:

- **Semantic button**: `PresenceChangeSection` の Presence 色 `el-button`
- **UI mode toggle**: `FriendsViewToolbar` の Online/Offline `el-switch`
- **`el-checkbox-group`**: Automation 曜日選択のグループラッパー（子は `VtCheckbox`）
- **未整備 EP**: `el-input-number` / `el-slider` / `el-date-picker` / `el-progress` / `el-autocomplete` / `el-radio-*` / `el-table` / `el-card` / `el-form*` 等
- **Domain badge**: `VrcStatusTag` / `VrcUserTagChip`
- **TitleBar** クローム固有寸法、**Sidebar** 絵文字（サイズトークンのみ）
- **レイアウト sizing**（`max-width: Nrem` 等）は Spacing 対象外として残存可
