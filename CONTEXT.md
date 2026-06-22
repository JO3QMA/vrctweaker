# VRChat Tweaker — Domain Language

アプリ横断のドメイン用語。実装詳細はここに書かない。

## Gallery

VRChat のスクリーンショットを閲覧・検索するための用語。

### Language

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
Gallery の詳細から、Screenshot に紐づくワールドへ VRChat を起動して入る操作。`world_id` が無い Screenshot では行えない。起動は Default launch profile を用いる。
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

## Launcher

VRChat の起動引数を名前付きで保存し、起動に使うための用語。

### Language

**Launcher**:
Launch profile を一覧・作成・編集・保存する画面体験。主目的は起動引数の編集と保存であり、VRChat の起動（Profile launch）は副次の導線。
_Avoid_: ランチャー画面, 起動画面（Quick launch / Profile launch と混同しやすいため）

**Launch profile**:
Tweaker が保存する起動設定のまとまり。表示名、起動引数の文字列、既定かどうか（`isDefault`）を持つ。Launcher 画面で編集し、Dashboard の Quick launch や World join のベース引数になる。
_Avoid_: プロファイル（VRChat profile slot と混同するため）, preset 単体

**Draft launch profile**:
Launcher で新規作成し、まだ DB に保存していない Launch profile（`id` が空）。サイドバー一覧には行が無く、Unsaved launch profile edits がある間はエディタ上部バナーで示す。別 Launch profile への切り替えやルート離脱時は確認ダイアログを出し、破棄すればドラフトは消える。
_Avoid_: 新規プロファイル, 仮プロファイル（保存済みとの境界が曖昧なため）

**Default launch profile**:
`isDefault` が真の Launch profile。Dashboard の Quick launch と、profile を指定しない World join が使う引数の出所。同時に存在できるのは高々 1 件。削除や既定フラグの解除後、どの Launch profile も `isDefault` でない状態があり得る（その間 Quick launch は利用できない）。
_Avoid_: 既定プロファイル（UI 表示は可。ドメイン文脈では Launch profile とセットで書く）

**VRChat profile slot**:
VRChat 起動引数 `--profile=N` で指定する、Unity 側のプロファイル番号（0 始まりのスロット）。Launch profile とは無関係。
_Avoid_: プロファイル, profile（Launch profile と混同するため）

**Primary launch options**:
Launcher エディタで常時表示する起動引数のまとまり。デスクトップモード、表示モード、カスタム引数文字列。Launch profile の名前や既定フラグは含まない。
_Avoid_: 基本設定, 日常設定（Launch profile 属性と混同しやすいため）

**Advanced launch options**:
Launcher エディタの折りたたみ内にまとめる起動引数。解像度、モニター、FPS、優先度、VRChat profile slot、デバッグ・MIDI など。Primary に含まれないものはすべてここに属する。
_Avoid_: 詳細設定, すべてのオプション（UI ラベルは可。ドメインでは Advanced と書く）

**Unsaved launch profile edits**:
Launcher エディタで、最後の保存または読み込み以降に加えた Launch profile の変更（名前、既定フラグ、Primary / Advanced の各引数）。保存前はサイドバーの未保存表示とエディタ上部バナーで示す。
_Avoid_: dirty 状態, 未保存（他画面の編集と混同しやすいため）

**Discard launch profile edits**:
Unsaved launch profile edits を保存せず、直前に保存または読み込みした内容に戻すこと。別 Launch profile への切り替え、新規作成、Launcher 以外の画面への移動の前に確認できる。
_Avoid_: リセット, クリア（カスタム引数フィールドの空欄化と混同しやすいため）

**Quick launch**:
Dashboard から Default launch profile の引数で VRChat を起動する操作。主な起動導線。常に DB に保存済みの Default launch profile を参照し、Launcher 上の Unsaved launch profile edits は反映しない。
_Avoid_: 起動, Launch（Profile launch と区別できないため）

**Profile launch**:
Launcher から、選択中 Launch profile の引数で VRChat を起動する操作。Unsaved launch profile edits があっても保存を強制せず、その編集中の引数で起動してよい。セカンダリ導線。
_Avoid_: このプロファイルで起動（UI 文言は可）, Quick launch（Default 固定ではないため）

## Activity

output_log から得た「誰と・どのワールドで会ったか」を振り返るための用語。

### Language

**Activity**:
遭遇ログを一覧・絞り込み・深掘りする画面体験。主目的は、同一インスタンスで重なった他ユーザーの滞在区間を追うこと。Encounter log を画面上部に置き、Play time chart はその下に副次セクションとして置く（既定は折りたたみ）。
_Avoid_: アクティビティ画面, ログ画面（output_log 生データやプレイ時間だけを指す語と混同しやすいため）

**User encounter**:
他ユーザーが同一インスタンスにいたひと区間の記録。入室時刻（joined-at）から退室時刻（left-at）まで。退室が未観測のとき left-at は空（滞在中）。
_Avoid_: 遭遇, 出会い（単発イベントの印象を与えるため）, タイムライン行

**Open encounter**:
left-at が未確定の User encounter。ログ上まだ退室が取れていない滞在。Encounter log では退室列に「滞在中」ラベルで示す（欠損の `—` とは区別する）。
_Avoid_: 未完了, アクティブ遭遇（実装状態と混同しやすいため）

**Unidentified encounter**:
VRC user ID が取れなかった User encounter。表示名は Encounter log に載せるが、プロフィールや Encounter history へのリンクは出さない（薄色テキスト）。
_Avoid_: 匿名ユーザー, 不明ユーザー（VRChat の匿名インスタンス設定と混同しやすいため）

**Encounter log**:
Activity に並べる User encounter の時系列一覧。画面上の見出しは「遭遇ログ」。入室・退室・表示名・ワールド名の4列（インスタンス ID は含めない）。入室時刻の新しい順が既定。表示名での絞り込みと、ユーザー・ワールド別の深掘りへの導線を持つ。
_Avoid_: 遭遇履歴（ユーザー／ワールド別の絞り込み画面全体を指す場合があるため）, ログ, タイムライン

**Display name filter**:
Encounter log 上の唯一の絞り込み。表示名の部分一致のみ（クライアント側）。ワールドや期間での絞り込みは Encounter history 側に任せる。
_Avoid_: 検索, フィルタ（Gallery の World search や Date range filter と混同しやすいため）

**Activity retention**:
Output log 由来の Activity データの保存上限。設定の保存期間（日）を過ぎた User encounter と Play session は自動削除される。Activity 画面ではページ全体（タイトル付近）に 1 回だけ期間を示すヒント文を置き、空状態だけに頼らない。
_Avoid_: Encounter retention（User encounter だけを指す印象）, ログ保持, データ削除（プレイ時間やスクリーンショットと混同しやすいため）

**Output log ingest**:
VRChat の output_log.txt を読み取り、User encounter・Play session・ワールド表示名など Activity の元データを更新すること。起動時の過去分取り込みと、稼働中の追記監視を含む。
_Avoid_: ログ解析, ログ同期（checkpoint やファイル切替と混同しやすいため）

**Activity refresh**:
Activity 画面の遭遇ログ一覧と Play time chart 用データの再取得。Output log ingest の後は自動で行う。画面上の手動更新は遭遇ログと Play time chart の両方を対象とし、取り込み漏れや不整合時にユーザーが再取得できる。
_Avoid_: Encounter log refresh（遭遇ログだけを指す印象）, 同期, リロード（画面全体の再読み込みと混同しやすいため）

**Encounter user navigation**:
Encounter log で識別済みユーザー（VRC user ID あり）の表示名を選んだときの遷移。対象がログイン中の自分なら Self profile へ。フレンドなら Friends へ。それ以外は User profile へ。遭遇の深掘りはプロフィール内や Encounter history から行う。
_Avoid_: プロフィール遷移, ユーザー詳細（Friends と区別できないため）

**Encounter world navigation**:
Encounter log でワールド名を選んだときの遷移。Encounter history（ワールド別）へ進み、そのワールドでの User encounter 一覧を見せる。VRChat への Join は行わない。
_Avoid_: ワールド Join, ワールド起動（Gallery や Launcher の導線と混同しやすいため）

**Encounter history**:
特定のユーザーまたはワールドに絞った User encounter の一覧。Activity の表から遷移するか、ユーザープロフィールなど別導線から開く。Activity 本体とは画面を分ける。
_Avoid_: 遭遇ログ（Activity 上の全体一覧と混同しやすいため）, 履歴画面

**Play session**:
ローカルユーザーが output_log 上でワールド／インスタンスに入ってから出るまでのひと区間。`Joining wrld_...` で始まり、`OnLeftRoom` / `Left room` / `Leaving room` で終わる。別ワールドへ移るたびに前の区間を閉じて新しい区間を開く。退室が未観測のとき終了時刻は空（進行中）。
_Avoid_: VRChat セッション, ログイン時間（クライアント起動全体や認証と混同しやすいため）

**Open play session**:
終了時刻が未確定の Play session。ログ上まだ `Left room` 系が取れていない滞在。日別 Play time では開始〜最後に観測した時刻までを暦日ごとに按分して含める。
_Avoid_: 未完了, アクティブセッション（実装状態と混同しやすいため）

**Play time**:
ローカルユーザーの Play session の滞在時間の合計。日別プレイ時間は端末ローカルタイムゾーンの暦日（0:00〜23:59）ごとに区間を割り当てて秒数を足したもの。Open play session も、開始〜 Output log ingest で最後に処理した行の時刻までを按分して含める。ワールド別の内訳は持たない（日別合計のみ）。
_Avoid_: プレイ時間（UI セクション名だけを指すとき）, 滞在時間（User encounter と混同しやすいため）, ワールド別プレイ時間（Encounter history や将来機能と混同しやすいため）

**Play time chart**:
Activity 上の副次セクション。Play time の日別合計を棒グラフで示す。表示する暦日数は 14 日と Activity retention の日数の小さい方（保存期間が 14 日未満のときは軸も短くする）。見出しもその日数（例: 直近7日）を反映する。遭遇ログの補助情報であり、Activity の主目的ではない。既定では折りたたみ、遭遇ログより下に置く。
_Avoid_: プレイ時間画面, アクティビティ統計（遭遇ログ全体を指す語と混同しやすいため）

## User detail

VRChat 上の人物（自分・フレンド・非フレンド）のプロフィールを閲覧する共通体験の用語。

### Language

**User detail**:
VRChat ユーザーのキャッシュ済みプロフィールを閲覧する共通体験。ヒーロー（バナー・アバター）、詳細タブ、遭遇履歴タブなどを含む。Friends の詳細ペイン、User profile 画面、Self profile で同じ表面を使う。
_Avoid_: ユーザープロフィール（User profile 画面体験と混同しやすいため）, プロフィール画面（Launch profile と混同しやすいため）

**Friends**:
フレンド一覧と User detail のマスター／ディテール画面体験。サイドバーから開く。一覧でユーザーを選ぶと右ペインに User detail を示す。
_Avoid_: フレンド画面, ユーザー一覧（Activity の遭遇ログ一覧と混同しやすいため）

**User profile**:
フレンド以外のユーザーを `vrcUserId` で開く単独画面体験。User detail を主コンテンツとして全面に示す。Activity の Encounter user navigation や外部導線から遷移する。
_Avoid_: ユーザープロフィール画面（User detail 全体と混同しやすいため）, プロフィール詳細

**Self profile**:
ログイン中のローカルユーザー自身の User detail。他ユーザーと同じ閲覧表面を使うが、お気に入りと遭遇履歴タブは出さない。詳細タブに Self profile refresh を置く。専用ルート `/me` で全面表示する。サイドバーに常時表示する項目があり、未ログインでもクリックで `/me` のログイン必要空状態へ進める。Settings profile summary の「詳細を見る」からも開ける。未ログインで `/me` を直接開いたときも Settings へリダイレクトせず、同じ空状態と Settings 導線を示す。表示データの正は Cached VRChat user（`users_cache` の self 行）。Settings のログイン確認用要約も同じ self 行の一部フィールドから派生する。
_Avoid_: 自分のアカウント, マイプロフィール（Dashboard や VRChat profile slot と混同しやすいため）

**Settings profile summary**:
Settings のログイン済みブロックに示す、Self profile の要約。アバター・表示名・ユーザー名・ステータスなど最小限の確認用情報。Cached VRChat user（self 行）の投影であり、User detail の代替ではない。「詳細を見る」で Self profile へ進む。
_Avoid_: 自己プロフィール（Self profile 本体と混同しやすいため）, プロフィールカード（User detail 全体と混同しやすいため）

**Self profile refresh**:
Self profile の詳細タブから、VRChat API 経由で Cached VRChat user（self 行）を再取得・更新する操作。Settings のプロフィール更新と同等の効果。Self profile 上で完結し、User detail 共通表面の自己向け差分として置く。
_Avoid_: プロフィール同期, 再読み込み（画面全体のリロードと混同しやすいため）

**Cached VRChat user**:
User detail の表示元となる、Tweaker が保持する VRChat ユーザー情報のスナップショット。表示名、ステータス、バイオ、ロケーション、お気に入りフラグなど。API 取得後に users_cache に保存される。
_Avoid_: UserCache, DTO（実装型名）, フレンド（Friends 画面体験と混同しやすいため）

**Self profile navigation**:
`vrcUserId` でユーザーを開く導線（Encounter user navigation、Friends の deep link、User profile への直リンクなど）のうち、対象がログイン中の自分のとき Self profile（`/me`）へ進めること。Friends や User profile にはフォールバックしない。
_Avoid_: Encounter user navigation（Activity 上の表示名クリックに限定した印象）, マイページ遷移

**Self profile nav**:
サイドバーで `/me` を開く常設項目の表示ラベル。i18n キー `nav.me` を用い、日本語は「自分」、英語は「Me」などロケールごとに短い呼び方にする。Friends や Settings の項目名とは別キーとする。
_Avoid_: プロフィール（Launch profile・User profile と混同しやすいため）, マイプロフィール（Self profile 画面体験の Avoid 語と重なるため）
