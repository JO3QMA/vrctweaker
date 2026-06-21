# VRChat Tweaker — Media (Gallery)

VRChat のスクリーンショットを閲覧・検索するためのドメイン用語。

## Language

**Gallery**:
スクリーンショットを一覧・詳細表示する画面体験。主に「いつ撮ったか」で写真を思い出す。
_Avoid_: ギャラリー画面, Photo library

**Screenshot**:
アプリがインデックスした VRChat スクリーンショット。ディスク上の画像ファイルと、抽出済みメタデータ（あれば）をひとまとめにした記録。
_Avoid_: Image, Photo, スクショファイル

**Date grouping**:
Gallery でスクリーンショットを並べる既定の見分け方。撮影日時（taken-at）に基づく年 → 月 → 日の階層。
_Avoid_: Timeline, カレンダー表示

**Taken-at**:
スクリーンショットを「いつ撮ったか」とみなす日時。画像メタデータの撮影日時を優先し、取れないときはファイルの更新日時で代用する。Date grouping の基準になる。
_Avoid_: 作成日時, ファイル日付（代用ルールを含意しないため）

**World search**:
日付グループを補う絞り込み。検索ボックス 1 つで、入力が `wrld_` で始まれば World ID の完全一致、それ以外はワールド表示名の部分一致として扱う。Date range filter と併用できる。
_Avoid_: ワールドフィルタ, World ID 検索（名前検索を含意しないため）

**Date range filter**:
Taken-at に基づき Gallery の Screenshot 一覧を期間で絞り込むフィルタ。開始日・終了日（from/to）を指定し、World search と組み合わせて使う。有効時も Date grouping（年→月→日）は維持する。
_Avoid_: 日付検索, カレンダーフィルタ（Date grouping と混同しやすいため）

**Picture folder**:
VRChat がスクリーンショットを保存するフォルダ。`config.json` の `picture_output_folder`、未設定時は OS 既定の VRChat Picture パス。Gallery に載せる Screenshot はこのフォルダ配下に限定する。
_Avoid_: スキャン先, 保存先パス（Launcher 設定全般と混同しやすいため）

**Gallery scope**:
Gallery に表示する Screenshot の集合。常に現行の Picture folder 配下に限定する。
_Avoid_: インデックス全体, DB 全件

**Out-of-scope screenshot**:
Picture folder 外にあり、Gallery には出さない Screenshot 記録。DB には残してよい（フォルダを戻したときの再表示などに備える）。
_Avoid_: 削除済み, アーカイブ（自動削除を連想させるため）

**Missing screenshot file**:
インデックスはあるがディスク上の画像ファイルが存在しない Screenshot。Gallery には表示しない（DB 行の去就は別判断）。一覧を取得するたびに存在を確認し、欠損は表示から除外する。
_Avoid_: 壊れたサムネ, 欠損ファイル（ユーザー向け用語として曖昧なため）

**World join**:
Gallery の詳細から、Screenshot に紐づくワールドへ VRChat を起動して入る操作。`world_id` が無い Screenshot では行えない。起動はデフォルトの Launch profile を用いる。
_Avoid_: Join ボタン, ワールド起動（Launcher 全般と混同しやすいため）

**Picture folder sync**:
現行 Picture folder と Gallery のインデックスを揃える操作。新規画像の取り込み、メタデータの再抽出（新規取込分、ソースファイルの更新があった行、`world_id` が空の行）をまとめて行う。欠損ファイルの Gallery 非表示は一覧 API が一覧取得のたびに担う（sync で DB 削除はしない）。
_Avoid_: フォルダをスキャン, 再インデックス（一部だけを指す語と混同しやすいため）

**Automatic ingest**:
Picture folder に追加された新規画像を、ウォッチャー経由でインデックスへ取り込むこと。欠損整理やメタ再抽出は含まない。
_Avoid_: 自動スキャン, リアルタイム同期（フル同期と混同しやすいため）

**Manual sync**:
ユーザーが Gallery 上の操作で Picture folder sync を明示的に開始すること。
_Avoid_: 手動スキャン, 更新ボタン（一覧再取得だけを指す場合があるため）
