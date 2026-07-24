# ADR 0013: Launch profile create（作成＝保存）

## Status

Accepted（grill-with-docs で合意、[Issue #210](https://github.com/JO3QMA/vrctweaker/issues/210)）

## Context

- Launcher の UI リデザインは「明示的保存のみ・新規は **Draft launch profile**」前提だった
- 「+ 新規プロファイル」は Draft だけをエディタに載せ、一覧に出ず、保存まで DB に残らない。切り替え・離脱時は未保存確認の対象になる
- ボタン名は「新規プロファイル」なのに、押しただけではプロファイルとして確定しない（Issue #210）

用語は [`CONTEXT.md`](../../CONTEXT.md) の **Launcher** セクション（**Launch profile create**、**Draft launch profile**、**Unsaved launch profile edits**、**Discard launch profile edits**、**Default launch profile**）を正本とする。

## Decision

1. **Launch profile create**: 「+ 新規プロファイル」押下で **即保存**する。作成直後はサイドバー一覧に現れ、選択状態になる。作成直後は Unsaved launch profile edits／Draft にならない
2. **Draft launch profile**: 概念としては残す（別導線・将来用）。**このボタンでは作らない**
3. **既定名**: ロケールの既定文字列。同名があれば「… 2」「… 3」…と連番（1 件目に番号なし）。空きは **既定名そのもの → `既定名 + " " + n`（n≥2 の最小）**。名前の一意制約は設けない（手動リネームの同名は可）
4. **起動引数初期値**: 空／GUI 既定。Default や選択中 profile の複製はしない
5. **Default フラグ**: 保存済みが 0 件のときの 1 件目だけ `isDefault` 真。既に 1 件以上ある作成では付けない
6. **未保存ガード**: 既存どおり。切り替えと同様、保存／破棄／キャンセルのあと（キャンセル時は作成しない）
7. **保存失敗**: 作成されなかった扱い。一覧に新行なし、選択は変えない。エラーをユーザーに示す
8. **作成後の編集**: 既存 Launch profile と同じく明示保存。**成功トーストは出さない**（一覧出現・選択がフィードバック）
9. **API**: 新規 Wails／usecase は置かない。フロントが既定名・連番・`isDefault`・空引数を組み立て、既存 **`SaveLaunchProfile`**（空 `id` → UUID 採番）で保存する。既定名が i18n 依存のため連番もフロント側
10. **空 `id` UI**: 主導線から Draft を作らなくなったため、Launcher 上の空 `id` 前提分岐は**削除**する。Draft は CONTEXT／本 ADR 上の将来用概念として残し、コード上の死んだ受け皿は残さない

## Failure modes（review-ready）

| 状況 | ユーザー |
|------|----------|
| 未保存編集あり → キャンセル | 作成しない。選択・編集内容はそのまま |
| 未保存編集あり → 保存失敗 | 作成に進まない。**明示保存と同じ** `ElMessage.error`（短い i18n ＋必要なら `formatBackendError`）。選択は現状のまま |
| 未保存編集あり → 破棄後、作成の保存失敗 | 同上のエラー表示。新行なし。選択は変えない（破棄後は前 profile を DB 内容で開き直す等。Draft には落とさない） |
| 作成の保存成功 | 一覧に新行、選択、未保存バナーなし |
| 明示保存ボタンの失敗 | 作成経路と**同じ**エラー表示（本 Issue で揃える） |
| 連打 | それぞれ保存済みとして作成されうる（連番）。専用の作成キューは持たない |
| 作成／明示保存の await 中にルート離脱 | **世代ガード**: unmount 後は `profiles`／`selected` を更新しない。`ElMessage` も出さない |

## Return-value contract（review-ready）

- **作成経路**は既存 `SaveLaunchProfile(p) error` のみ
- 空 `id` → usecase が UUID 採番（現行どおり）
- 失敗 → `error`。フロントは一覧を楽観追加しない（保存成功後に `launchProfiles()` で再取得して選択）
- フロントのエラー表示: **Launch profile create** とエディタの明示保存で共通（`ElMessage.error`）
- フロントの async: 作成・明示保存は **generation counter**（state-changing）。ダイアログ（未保存確認）は capture のみで bump しない

## Considered Options

| 案 | 採否 | 理由 |
|----|------|------|
| 作成＝保存（本 Decision） | 採用 | ボタン名と一致。作成直後の未保存確認が消える |
| 新規は Draft のまま | 却下 | Issue の期待と不一致 |
| 作成時に名前入力ダイアログ | 却下 | ワンアクション作成から外れる |
| 既定名は常に同一（同名許可のみ） | 却下 | 連番の方が一覧で見分けやすい |
| Default／選択中の複製として新規 | 却下 | 複製導線の追加であり Issue 範囲外 |

## Consequences

- Draft 用語と主導線が一致しないように見えるが、意図的（Draft は残置、主ボタンは create＝保存）
- 連打すると連番付きの保存済み profile が複数できる（Draft 破棄より残存しやすい）
- 既存の明示保存・Unsaved／Discard のモデルは、作成後の編集では維持する
- Launcher から空 `id` 編集 UI が消える。将来 Draft 導線を足すときは UI／状態を再追加する
- 明示保存の失敗時も `ElMessage.error` が出るようになり、作成前より失敗が目立つ（意図的な揃え）
