# ADR 0015: Launch profile duplicate（選択中の即保存複製）

## Status

Accepted（grill-with-docs で合意、[Issue #209](https://github.com/JO3QMA/vrctweaker/issues/209)）

## Context

- Launcher エディタ上部は `[起動]` `[保存]` と **「⋯」オーバーフロー**（中身は削除のみ）だった
- 「⋯」では何ができるか一目で分からず、**複製**自体も未実装（UI リデザインではスコープ外）
- [ADR 0013](0013-launch-profile-create.md) は **Launch profile create** で Default／選択中の複製を明示的に却下した（create＝空／GUI 既定の即保存）

用語は [`CONTEXT.md`](../../CONTEXT.md) の **Launcher** セクション（**Launch profile duplicate**、**Launch profile editor actions**、**Launch profile create**、**Unsaved launch profile edits**、**Discard launch profile edits**、**Default launch profile**）を正本とする。

## Decision

1. **Launch profile duplicate**: 選択中 Launch profile の設定を引き継ぎ、新しい Launch profile として**即保存**する。create（空／GUI 既定）とは別操作
2. **Launch profile editor actions**: 選択中があるとき、左から Profile launch・明示保存・duplicate・削除を**ラベル付き**で並べる。削除は danger。「⋯」オーバーフローは廃止（削除・複製を隠さない）
3. **未保存ガード**: create／切替と同じ。保存／破棄／キャンセルのあと（キャンセル時は複製しない）。解決後は選択中の**保存済み内容**を引き継ぐ
4. **引き継ぎ**: 表示名の基点＋起動引数（Primary / Advanced）。複製先の `isDefault` は**常に偽**（Default は移さない）
5. **表示名**: 元の表示名を基点に、create と同じ連番ルール（`nextDefaultLaunchProfileName`）。「コピー」接尾辞は使わない
6. **確認ダイアログ**: 削除のような追加確認は出さない（未保存ガード以外）
7. **成功／失敗フィードバック**: create と同型。成功トーストなし（一覧出現・選択がフィードバック）。失敗は新行なし・選択据え置き・`ElMessage.error`
8. **削除**: 確認文言・削除後選択など**現行維持**。出す場所だけラベル付きボタンへ
9. **API**: 新規 Wails／usecase は置かない。フロントがペイロードを組み立て、既存 **`SaveLaunchProfile`**（空 `id` → UUID）で保存する（ADR 0013 と同型）
10. **連打**: それぞれ保存済みとして複製されうる（連番）。専用キューは持たない。await 中のルート離脱は**世代ガード**

## Failure modes（review-ready）

| 状況 | ユーザー |
|------|----------|
| 未保存編集あり → キャンセル | 複製しない。選択・編集内容はそのまま |
| 未保存編集あり → 保存失敗 | 複製に進まない。明示保存／create と同じ `ElMessage.error`。選択は現状のまま |
| 未保存編集あり → 破棄後、複製の保存失敗 | 同上。新行なし。選択は変えない |
| 複製の保存成功 | 一覧に新行、新 profile を選択、未保存バナーなし |
| 連打 | それぞれ保存済みとして複製されうる（連番） |
| 複製の await 中にルート離脱 | **世代ガード**: unmount 後は `profiles`／`selected` を更新しない。`ElMessage` も出さない |

## Return-value contract（review-ready）

- **複製経路**は既存 `SaveLaunchProfile(p) error` のみ（空 `id` → UUID）
- 失敗 → `error`。フロントは一覧を楽観追加しない（保存成功後に `launchProfiles()` で再取得して選択）
- エラー表示: create／明示保存と共通（`ElMessage.error`）
- async: 複製は **generation counter**（state-changing）。未保存確認ダイアログは capture のみで bump しない

## Considered Options

| 案 | 採否 | 理由 |
|----|------|------|
| create とは別操作の duplicate（本 Decision） | 採用 | 0013 の create 定義を壊さず Issue を満たす |
| create に「選択中を複製」モードを足す | 却下 | create＝空／GUI 既定の契約が崩れる |
| 未保存内容をそのまま複製（ガードなし） | 却下 | create／切替と未保存モデルがずれる |
| 「コピー」接尾辞の表示名 | 却下 | create の連番ルールと揃える方が単純 |
| 複製先を Default にする／移す | 却下 | Default は意図的な既定変更のみ |
| 専用 `DuplicateLaunchProfile` API | 却下 | create と同型の `SaveLaunchProfile` で足りる |
| ⋯ に複製を足すだけ | 却下 | Issue はラベル付き明示と ⋯ 廃止 |

## Consequences

- create と duplicate が並ぶが、初期値（空 vs 引き継ぎ）と Default 規則が違う
- 連打すると連番付きの保存済み profile が増えうる（create と同じ）
- Last launch profile は起動成功時のみ更新するため、複製だけでは変わらない
- エディタ上部が横に広がる（起動・保存・複製・削除）。狭い幅での折り返しは実装時に既存 toolbar スタイルに合わせる
