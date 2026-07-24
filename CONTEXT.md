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

## Dashboard

起動・自分のプレゼンス変更・公式サービス健全性をまとめるホーム画面体験の用語。**Server status**（[Issue #10](https://github.com/JO3QMA/vrctweaker/issues/10)、[ADR 0009](docs/adr/0009-dashboard-server-status.md)）は grill-with-docs / grill-review-ready で合意済み。実装契約は ADR を正本とする。

### Language

**Dashboard**:
サイドバー先頭（`/`）のホーム画面体験。**Dashboard launch block**、自分のプレゼンス変更（Quick status 等）、**Server status** を置く。
_Avoid_: ダッシュボード画面, ホーム（他アプリのホームと混同しやすいため）

**Server status**:
[status.vrchat.com](https://status.vrchat.com/) が示す VRChat **サービス**の健全性（API / 認証 / リージョン別リアルタイムなど）。ログイン中ユーザー自身の join me / busy などのプレゼンスや、フレンドのオンライン状況とは別物。
_Avoid_: VRChat status, ステータス（Quick status や Cached VRChat user の status と混同しやすいため）, Server Status Page（取得元の俗称）

**Server status summary**:
Dashboard の Server status で常に示す全体の健全性の要約（例: 全システム正常 / 一部障害）。公式 `summary.json` のインジケータに相当。平常時はこれと外部リンク程度に畳む。
_Avoid_: Quick status, サマリー行（Activity 等の別サマリーと混同しやすいため）

**Server status detail**:
component ごとの健全性の内訳。**Abnormal server status** のときだけ Dashboard に展開する。平常時（全 component が operational）は出さない。v1 の展開内容は、(1) リーフ component の名前とステータス、(2) 未解決インシデントの見出し（あれば 1 件）、(3) 予定または進行中メンテナンスの見出し（あれば）とする。過去インシデント履歴や公式ページ相当の全文は載せない。
_Avoid_: component 一覧（常時表示を含意するため）, ステータス詳細（ユーザープロフィールと混同しやすいため）

**Abnormal server status**:
Server status detail を Dashboard に出す条件。少なくとも 1 つの component が `operational` 以外（`degraded_performance` / `partial_outage` / `major_outage` / `under_maintenance` など）のとき。未解決インシデントや予定メンテの有無は v1 では detail 展開条件に含めない（component 状態のみで判定）。
_Avoid_: 障害中, メンテ中（component 状態とインシデント文面を同一視しないため）

**Server status section**:
Dashboard 上 **Quick launch より上**に置く Server status の UI ブロック。Server status summary を常時示し、Abnormal server status のときだけ Server status detail を展開する。status.vrchat.com への外部リンクを含む。
_Avoid_: Quick status パネル, ステータスカード（個人プレゼンス変更と混同しやすいため）

**Server status refresh**:
Server status section のデータ再取得。Dashboard 表示時（`onMounted`）に 1 回行い、表示中は一定間隔（v1: 5 分）でバックグラウンドポーリングする。手動リフレッシュボタンは v1 では置かない。取得は Go バックエンド経由（フロントから status.vrchat.com へ直接 fetch しない）。
_Avoid_: Activity refresh, 同期ボタン（他画面の更新と混同しやすいため）

**Server status visibility**:
Server status section の表示条件。VRChat へのログイン状態に依存せず、Dashboard を開けば常に表示する（未ログインでも取得・表示する）。Quick status とは別で、公開の status.vrchat.com API のみを参照する。
_Avoid_: ログイン必須, 認証後のみ（Quick status の条件と混同しやすいため）

**Server status fetch failure**:
status.vrchat.com からの取得が失敗したときの扱い。Server status section は残し、取得できなかった旨だけを示す。次回の Server status refresh（ポーリング）で再試行する。最後に成功した値の stale 表示や、失敗時だけの手動リフレッシュは v1 では行わない。
_Avoid_: 非表示, オフライン（ネットワーク断と公式障害を同一視しないため）

**Server status labeling**:
Server status section の表示言語の扱い。component 名とインシデント／メンテ見出しは API 原文（英語）のまま示す。ステータス値（operational / under_maintenance 等）と Server status summary の文言は UI ロケール（i18n）で翻訳する。
_Avoid_: 全英語表示, 日本語 component 名（公式表記と照合しづらいため）

**Server status presentation**:
Server status section の視覚表現。他 Dashboard パネルと同様 `el-card` で枠を揃えつつ、健全性は status.vrchat.com に近い色分け（正常=緑、パフォーマンス低下=黄、部分障害=橙、重大障害=赤、メンテナンス=青系）で示す。平常時はコンパクトなサマリー行、Abnormal server status 時は Server status detail を同色ルールで展開する。
_Avoid_: Quick status ボタン色（個人プレゼンス用の独自パレットと混同しやすいため）, モノクロのみ（障害判別が弱いため）

**Server status v1 scope**:
Issue #10 で最初に届ける Server status の範囲。Dashboard の Server status section（取得・表示・外部リンク）に限定する。v1 では含めないもの: 障害時 OS 通知、Settings でのオンオフやポーリング間隔変更、リージョン絞り込み、Dashboard 以外への常設表示、取得結果のローカル履歴・グラフ。
_Avoid_: Server status（v1 機能全体を指すときは section とセットで書く）, 将来拡張（スコープ外リストの総称として曖昧なため）

**Quick status**:
Dashboard 上でログイン中ユーザー自身の VRChat プレゼンス（join me / active / ask me / busy）を API 経由で変更する操作群。Custom status・Templates と併置する。**Server status** とは無関係。
_Avoid_: ステータス, Server status, クイックステータス（インフラ健全性と混同しやすいため）

## Launcher

VRChat の起動引数を名前付きで保存し、起動に使うための用語。

### Language

**Launcher**:
Launch profile を一覧・作成・編集・保存する画面体験。主目的は起動引数の編集と保存であり、VRChat の起動（Profile launch）は副次の導線。
_Avoid_: ランチャー画面, 起動画面（Quick launch / Profile launch と混同しやすいため）

**Launch profile**:
Tweaker が保存する起動設定のまとまり。表示名、起動引数の文字列、既定かどうか（`isDefault`）を持つ。Launcher 画面で編集し、Dashboard の起動ブロック（Quick launch / Instance rejoin）や World join のベース引数になる。Launch profile が 1 件も無いときの初回シードでは、表示名 **Desktop**（`--no-vr`・Default）と **VR**（`-no-vr` なし）の 2 件を用意する。既に 1 件以上ある DB への後追い追加はしない。
_Avoid_: プロファイル（VRChat profile slot と混同するため）, preset 単体, Non-VR（シード表示名は Desktop）

**Draft launch profile**:
Launcher で新規作成し、まだ DB に保存していない Launch profile（`id` が空）。サイドバー一覧には行が無く、Unsaved launch profile edits がある間はエディタ上部バナーで示す。別 Launch profile への切り替えやルート離脱時は確認ダイアログを出し、破棄すればドラフトは消える。
_Avoid_: 新規プロファイル, 仮プロファイル（保存済みとの境界が曖昧なため）

**Default launch profile**:
`isDefault` が真の Launch profile。profile を指定しない World join が使う引数の出所。Dashboard の起動ブロックで Last launch profile が無効なときのフォールバック解決にも使う。同時に存在できるのは高々 1 件。削除や既定フラグの解除後、どの Launch profile も `isDefault` でない状態があり得る。初回シードでは Non-VR（`--no-vr`）の Launch profile が既定になる（VR 側は既定にしない）。
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
Dashboard の起動ブロック内で、共有 Launch profile セレクタの選択中 profile の引数で、部屋指定なしに VRChat を起動する操作。主な起動導線のひとつ。DB に保存済みの profile のみ参照し、Launcher 上の Unsaved launch profile edits は反映しない。起動プロセスの開始に成功したとき Last launch profile を更新する。
_Avoid_: 起動, Launch（Profile launch と区別できないため）

**Profile launch**:
Launcher から、選択中 Launch profile の引数で VRChat を起動する操作。Unsaved launch profile edits があっても保存を強制せず、その編集中の引数で起動してよい。起動プロセスの開始に成功したとき、選択中 Launch profile（保存済み profile ID）で Last launch profile を更新する。セカンダリ導線。
_Avoid_: このプロファイルで起動（UI 文言は可）, Quick launch（Default 固定ではないため）

**Rejoin target**:
Instance rejoin の対象となる Play session。VRChat instance key（`play_sessions.instance_id`）が空でない Play session のうち、開始時刻（`start_time`）が最も新しい 1 件。Open play session か終了済みかは問わない。複数 Log source で Open play session が同時にあっても、開始時刻が最も新しい 1 件だけを選ぶ。**Activity retention** により対象 Play session が削除された場合は Rejoin target は存在しない。
_Avoid_: 最後のセッション（Play session / VRChat クライアント起動 / ログインと混同しやすいため）, 最後のインスタンス（VRChat instance key 以外の意味を含みうるため）

**Last launch profile**:
直近に **Profile launch**、**Quick launch**、または **Instance rejoin** で VRChat 起動プロセスの開始に成功した Launch profile。`app_settings` に profile ID として永続化し、Dashboard 起動ブロックの Launch profile セレクタ初期選択に使う。Dashboard 上でセレクタだけ変更して起動していない場合は更新しない。参照先 profile が削除された場合は **Default launch profile** にフォールバックする。
_Avoid_: 既定プロファイル（Default launch profile と混同しやすいため）, 前回のプロファイル（Launcher の選択状態だけを指す印象）

**Dashboard launch block**:
Dashboard 上 **Server status section** 直下の起動 UI ブロック。共有の Launch profile セレクタ、**Quick launch** ボタン（汎用ラベル。例: VRChat を起動）、**Instance rejoin** ボタン（Rejoin target があるときのみ）を含む。**常に表示する。** 保存済み Launch profile が 1 件以上あるときはセレクタと Quick launch を有効にし、Rejoin target があるときだけ Instance rejoin ボタンも出す。profile が 0 件のときはセレクタとボタンを無効化し、**Launcher** で Launch profile を作成する旨の短い説明と **Launcher へのリンク（またはボタン）** を示す。Rejoin target が無いとき（profile はある）はセレクタと Quick launch だけ有効にし、Instance rejoin ボタンは出さない。Activity retention 等で Rejoin target が消えてもブロック自体は残す（説明は出さない）。**Dashboard launch block fetch failure** のときはブロックを残し取得失敗の旨をブロック内に示す（`ElMessage` は出さない）。`activity:encounters-changed` 等の再取得で復帰しうる。
_Avoid_: 起動パネル, Quick launch ボタン（ブロック全体と混同しやすいため）, Instance rejoin section（統合前の旧称）

**Dashboard launch profile**:
Dashboard launch block の共有 Launch profile セレクタで選ぶ Launch profile。**Quick launch** と **Instance rejoin** の両方がこの選択を使う。初期値は Last launch profile → Default launch profile → 保存済み Launch profile 一覧の先頭、の順で解決する。起動引数は profile に保存済みの内容（`-no-vr` 含む）をそのまま使い、Dashboard 上での起動時オーバーライドは行わない。Launcher 上の Unsaved launch profile edits は反映しない（保存済み profile の DB 内容を参照する）。
_Avoid_: Default launch profile（常に Default とは限らないため）, Display mode override（起動時だけ Desktop/VR を差し替える機能は持たない）, Instance rejoin launch profile（統合前の旧称）

**Instance rejoin**:
Dashboard launch block から Rejoin target の VRChat instance key を使い、共有セレクタで選択中の Launch profile の引数で VRChat を起動し、同じ部屋へ入る操作。起動 URL は Rejoin target の instance key 丸ごと（`vrchat://launch?id=<VRChat instance key>`）。ボタンラベルは Rejoin target 由来のワールド表示名（`world_info`）があるとき「{ワールド名} に参加」、無いとき汎用ラベル（例: 最後のインスタンスに参加）。`wrld_*` など技術 ID はボタンに出さない。World join（`world_id` のみで新規インスタンスになりうる）や Quick launch（部屋指定なし）とは別導線。起動プロセスの開始に成功したとき Last launch profile を更新する。満員・非公開などで入れない場合の成否は VRChat 側に委ねる。
_Avoid_: 最後のセッションに参加, Rejoin（Profile launch や World join と区別できないため）

## Activity

output_log から得た「誰と・どのワールドで会ったか」を振り返るための用語。

### Language

**Activity**:
遭遇ログを一覧・絞り込み・深掘りする画面体験。主目的は、同一インスタンスで重なった他ユーザーの滞在区間を追うこと。Encounter log を画面上部に置き、Play time chart はその下に副次セクションとして置く（既定は折りたたみ）。
_Avoid_: アクティビティ画面, ログ画面（output_log 生データやプレイ時間だけを指す語と混同しやすいため）

**User encounter**:
他ユーザーが同一 VRChat instance key にいたひと区間の記録。入室時刻（joined-at）から退室時刻（left-at）まで。Output log ingest 由来の行は属する Log source を持つ（UI には出さない）。退室が未観測のとき left-at は空（滞在中）。
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
VRChat の output_log を読み取り、User encounter・Play session・ワールド表示名など Activity の元データを更新すること。起動時の過去分取り込みと、稼働中の追記監視を含む。監視対象はログフォルダ（未設定時は既定の VRChat ログフォルダ）であり、フォルダ内の複数 Log source を並行 ingest する。
_Avoid_: ログ解析, ログ同期（checkpoint やファイル切替と混同しやすいため）, 単一ファイル監視, ファイル直接指定

**Log source**:
Output log ingest の単位。ひとつの VRChat クライアントが書き込む `output_log*.txt` 1 本に対応する。相関状態の分離と、プレイセッション・遭遇の finalize スコープの基準になる。識別子は正規化したログファイルの絶対パス。ログローテーションでパスが変わったら新パスは別 Log source とし、Log rotation handoff で旧 Log source を finalize する。新 Log source 側では相関状態をログ replay で再構築する（旧 Log source の状態は引き継がない）。
_Avoid_: インスタンス, instance_id（VRChat instance key と混同しやすいため）, プロセス

**Log rotation**:
稼働中の VRChat クライアントが新しい `output_log*.txt` へ切り替えること。新ファイルは新 Log source。旧ファイルは増加停止かつ別ファイルが増加開始した時点で **Log rotation handoff** として旧 Log source を finalize する（60 秒 stall を待たない）。
_Avoid_: ログ切替, ファイルローテーション（OS・一般ログのローテーションと混同しやすいため）

**Log rotation handoff**:
watch ディレクトリ内で、ある Log source のファイルが増加停止し、別の `output_log*.txt` が増加を始めたとき、旧 Log source の open 行を finalize して tail を止めること。同一クライアントのログローテーション向け。複数クライアントが同時に増加している場合は発火しない（両方とも tail 継続）。
_Avoid_: ログ切替, ファイルスイッチ（MultiOutputLogWatcher の実装語）

**Log source stall**:
ある Log source の `output_log` が一定時間（60 秒）サイズ増加しなくなった状態。tail の goroutine は停止し checkpoint を保存するが、**この時点では open な User encounter / Play session は finalize しない**（ワールド滞在中のログ沈黙による誤退室を避ける）。finalize は VRChat 全終了、Log rotation handoff、または当該 Log source 上の Joining / Left room など既存の相関ルールに委ねる。
_Avoid_: タイムアウト, アイドル切断（ネットワーク切断と混同しやすいため）

**Log replay**:
Output log ingest のうち、すでにディスク上にある行を offset から読み直して Activity の相関状態を再構築すること。起動時 bootstrap を含む。User encounter・Play session の更新のみ行い、Friend joined などの automation は発火しない（automation は追記監視の live tail に限る）。
_Avoid_: ログ再処理, catch-up ingest（live tail との境界が曖昧なため）, bootstrap（起動時だけを指す印象）

**VRChat instance key**:
ログ上の部屋識別子（例: `wrld_…:room~type`）。User encounter と Play session が「どのワールド／部屋か」を表すときに使う。複数 VRChat クライアントが同じ部屋に入ってもキーは同じになりうる。Log source とは別概念。
_Avoid_: instance_id（列名・実装語）, インスタンス ID（Log source と混同しやすいため）, インスタンス

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
ローカルユーザーが output_log 上でワールド／VRChat instance key に入ってから出るまでのひと区間。属する Log source と VRChat instance key を持つ（instance key は UI に出さない）。`Joining wrld_...` で始まり、`OnLeftRoom` / `Left room` / `Leaving room` で終わる。別ワールドへ移るたびに同一 Log source 内で前の区間を閉じて新しい区間を開く。退室が未観測のとき終了時刻は空（進行中）。
_Avoid_: VRChat セッション, ログイン時間（クライアント起動全体や認証と混同しやすいため）

**Open play session**:
終了時刻が未確定の Play session。ログ上まだ `Left room` 系が取れていない滞在。複数 Log source が同時に稼働すると、Log source ごとに Open play session が同時に存在しうる。日別 Play time では開始〜最後に観測した時刻までを暦日ごとに按分して含める（複数 open は合算する）。
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

**Listable friend**:
Friends のマスター一覧（オンライン／オフライン切替を含む）に載せてよい Cached VRChat user。表示名が空でないことが必須。VRChat API がフレンド関係を肯定していること（`isFriend=true`、または Friends REST 同期の一覧に含まれること）も必須。VRChat Pipeline のプレゼンスだけでは Listable friend にならない。一覧取得時の条件と、キャッシュ書き込み時の昇格条件の両方で守る。
_Avoid_: フレンド（Friends 画面体験全体）, user_kind=friend（DB 上の分類と混同しやすいため）

**Profile resolution**:
Cached VRChat user に表示名などプロフィールフィールドを埋めること。VRChat の Friends REST 同期（フレンド一覧取得）または単体ユーザー取得（`GET /users/{id}`）で行う。Pipeline のプレゼンスイベント単体では Profile resolution にならない。Pipeline 受信時は単体取得を試み、失敗時は Reconcile / RefreshFriends 側で再試行する（ハイブリッド）。
_Avoid_: フレンド同期, キャッシュ更新（プレゼンス更新だけを含意しないため）

**Unresolved friend presence**:
VRChat Pipeline の friend-* イベントで分かったプレゼンス（status / platform / location など）だが、当該時点で Profile resolution できなかった Cached VRChat user の状態。`user_kind=contact` としてプレゼンスだけ保持し、Listable friend にはしない。お気に入りフラグは付けない（降格時はクリアする）。Reconcile または後続の Profile resolution 成功時に `friend` へ昇格しうる。過去に誤って `user_kind=friend` かつ表示名空で保存された行は、アプリ起動時マイグレーションで `contact` に降格する。降格直後、ログイン済みセッションでは対象 ID へ Profile resolution を 1 回試行する。
_Avoid_: 無名フレンド（Listable friend に出してしまう現状バグの俗称）, 仮フレンド

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

**User tag**:
Cached VRChat user の `tagsJson`（およびアバタータグ）に含まれる VRChat タグ文字列。User detail では User tag chip として一覧表示する。API に載っているものだけを表示し、タグ ID が無いランクは合成しない。
_Avoid_: Trust rank（User tag の一部）, タグ ID（チップの内部識別子・生文字列と混同しやすいため）

**Trust rank tag**:
`system_trust_` で始まる User tag。VRChat の Trust rank（New User, User, Known User, Trusted User など）に対応する。チップの表示ラベルは VRChat クライアントと同様、UI ロケールに関係なく英語の短い名称（色の括弧付き suffix なし）とする。ツールチップの説明文は UI ロケールの翻訳のまま。deprecated な trust タグも同じラベル規則に含める。
_Avoid_: ソーシャルランク, Trust level（Visitor や色名だけを指す印象）

**Visitor**:
VRChat 上の最下位 Trust rank の表示名。trust rank タグを 1 つも持たないユーザーに対応するが、API の `tagsJson` にはタグ ID として現れない。VRCTweaker は User tag として合成表示しない。
_Avoid_: Visitor タグ, `system_trust_visitor`（存在しない ID）

**User tag chip**:
User detail で 1 つの User tag を示すチップ UI。ラベルとツールチップ（説明・deprecated 表示）を持つ。ツールチップにタグ ID 行は出さない。未知タグはラベルに生のタグ文字列を示し、ツールチップは不明旨のみ。
_Avoid_: バッジ, タグ一覧（行全体のラベル付きセクションと混同しやすいため）

## yt-dlp

VRChat の動画プレイヤーが裏で使う yt-dlp 向けの用語。**yt-dlp Tools replace maintain**（[Issue #9](https://github.com/JO3QMA/vrctweaker/issues/9)、[ADR 0008](docs/adr/0008-ytdlp-tools-replace-maintain.md)）は Accepted・製品実装済み。**yt-dlp Cookie linkage**（[Issue #8](https://github.com/JO3QMA/vrctweaker/issues/8)、[ADR 0007](docs/adr/0007-ytdlp-cookie-linkage.md)）は Accepted・製品実装済み（制限付き再生には Official yt-dlp cache 経由が前提。同梱版単体では Cookie オプション非対応）。起動前ワンショットの直置き試作は [PR #40](https://github.com/JO3QMA/vrctweaker/pull/40)（望む動作に未達）。

### Language

**VRChat-bundled yt-dlp**:
VRChat が Tools 配下に置く yt-dlp 実行ファイル。公式 yt-dlp を削った／独自オプション付きのビルドであり、調査時点では `--cookies` / `--cookies-from-browser` を受け付けない。起動やログインの過程で Tools 上の差し替えを同梱版へ戻しうることがある。
_Avoid_: 公式 yt-dlp, yt-dlp.exe（どちらを指すか曖昧なため）

**yt-dlp Tools replace**:
Tools 上の `yt-dlp.exe` を Official yt-dlp cache の実体を指すリンクとして載せ替える**一回の配置操作**。Tools ディレクトリ内への公式バイナリの直置きコピーは含めない。起動前のワンショットだけだと VRChat が同梱版へ戻しうる。読み取り専用で戻しを防ぐと再生不能になりうる。維持の仕組み全体は **yt-dlp Tools replace maintain**。
_Avoid_: yt-dlp 更新（ユーザー config や Cookie linkage と混同しやすいため）, バージョン管理（UI 見出しは可）, 維持モード（maintain を指すときは専用語を使う）, Tools への直置きコピー

**yt-dlp Tools replace maintain**:
ユーザーが有効化した、Official yt-dlp cache を Tools から参照し続ける desired 状態（オプトイン・既定オフ）。Tweaker 常駐中に VRChat 起動を検知して yt-dlp Tools replace と監視を行い、無効化時は監視だけ止めて Tools 上のファイルは触らない。v1 は Windows のみ（動画タブ）。製品方針の正本は ADR 0008（Accepted）。
_Avoid_: yt-dlp Tools replace（一回の配置操作）, Cookie linkage, 自動更新（明示更新と混同しやすいため）

**Tools replace effective state**:
Tools 上の `yt-dlp.exe` が Official yt-dlp cache を指しているかどうかで決まる実効状態。desired（maintain オン／オフ）とは別。動画タブは両方を示す。
_Avoid_: 維持オン（desired だけを指す印象）, 適用済み（監視中と混同しやすいため）, バイト一致（直置きコピー前提の印象）

**Tools replace risk acknowledgment**:
yt-dlp Tools replace maintain を初めて有効化する前に、同梱版を外すリスクと公式の差し替え非推奨をユーザーが確認したこと。一度行えば以降の有効化では再確認しない。画面上の常時警告文とは別。
_Avoid_: Cookie linkage risk acknowledgment（別機能）, 利用規約同意, 毎回確認

**Official yt-dlp cache**:
Tweaker が保持する公式 `yt-dlp.exe` のローカル控え。初回適用と明示の更新確認で取得し、以降の VRChat セッションではこの控えから Tools へ配置する。
_Avoid_: 最新版（キャッシュと GitHub latest を同一視する印象）, Tools 上の exe（effective 側）

**yt-dlp Cookie linkage**:
Tweaker が yt-dlp user config へ Cookie 参照オプションを書き込み／削除する設定体験（UI は動画タブ。製品実装済み。実装契約は ADR 0007 を正本とする）。Cookie 本体の取得・検証や動画の取得は行わない。VRChat の `config.json` を扱う Config 画面の対象ではない。有効化は Tools replace maintain にハード依存しないが、**Tools replace effective state** が偽のときは **Cookie linkage official hint** を出す（同梱 yt-dlp では Cookie オプションが効かないため）。
_Avoid_: Cookie 同期, ログイン連携（VRChat 認証と混同しやすいため）, yt-dlp 実行, Config（VRChat config.json 編集と混同しやすいため）, yt-dlp Tools replace / maintain（別問題）, Settings（Cookie は載せない）

**Cookie linkage official hint**:
Cookie linkage の UI 上で、Tools replace effective state が偽のときに出す案内。Official（Cookie 対応）exe が Tools から参照されていないと制限付き再生の目的を満たせない旨と、同一画面の Tools replace 操作への案内。有効化や yt-dlp user config への書き込みは妨げない。
_Avoid_: 必須ゲート, maintain 必須（ハード依存を連想させるため）, Cookie linkage risk acknowledgment（BAN 警告とは別）

**yt-dlp user config**:
yt-dlp が読むユーザー向け設定ファイル。Cookie 参照オプションの置き場。Windows での書き込み正本は `%APPDATA%\yt-dlp\config`（拡張子なし）。`config` が無く `%APPDATA%\yt-dlp\config.txt` だけがあるときは、そのファイルを Effective／upsert の対象とする（新規作成時は常に `config`）。VRChat の `config.json`（Config 画面の対象）とは別物。無ければ親ディレクトリごと作成してよい。Managed cookie options 削除後に他行が無く空ならファイル自体を削除してよい。
_Avoid_: VRChat config, config.json, yt-dlp 設定（対象ファイルが曖昧なため）, 常に config.txt（Issue 原文の俗称）

**Managed cookie options**:
yt-dlp Cookie linkage が yt-dlp user config 内で所有する Cookie 参照オプション（`--cookies-from-browser` / `--cookies` のみ）。有効時は Browser cookie source か Cookies file source のどちらか一方だけ（排他）。誰が書いたかに関わらず、同種の Cookie 参照行は Managed とみなし、書き込み時はそれらを upsert（置換）する。無効化時はこれらの行だけを削除する。ファイル内の他行（sleep 間隔など）は触らない。v1 で間隔オプションを Managed に含めない。
_Avoid_: yt-dlp 設定全体, config 全体（手書きオプションまで含意するため）, 設定ファイルの退避・リネーム（無効化の意味に含めない）, sleep／リクエスト間隔の自動挿入

**Browser cookie source**:
Cookie 参照方式のひとつ。指定したブラウザのログイン Cookie を yt-dlp に読ませる。v1 で選べるブラウザは chrome / edge / firefox の既定プロファイルのみ（プロファイルパス指定なし）。ブラウザ起動中は Cookie ストアがロックされ、読み込みに失敗しうる。
_Avoid_: ブラウザ連携, Chrome 連携（特定ブラウザに固定する印象）, プロファイル指定（v1 の範囲外）

**Cookies file source**:
Cookie 参照方式のひとつ。ユーザーが用意した cookies テキストファイルのパスを yt-dlp に読ませる。Browser cookie source のファイルロック回避手段。ファイルの作成・更新自体は Tweaker の責務外。動画タブではパスを **テキスト入力＋参照ボタン**（既存のファイル選択ダイアログ）で指定する。Managed cookie options への書き込み前に、指定パスにファイルが存在することを必須とする（空ファイルも可。形式の中身検証・非空チェックはしない）。
_Avoid_: Cookie エクスポート（Tweaker がファイルを作る印象）, cookies.txt（ファイル名に限定する印象）, 参照のみ／手入力不可

**Cookie linkage risk acknowledgment**:
yt-dlp Cookie linkage について、Tweaker が **初めて yt-dlp user config へ書き込む操作**（有効化・方式変更・無効化を含む）の前に、アカウント BAN リスクとサブアカウント利用の推奨をユーザーが確認したこと。一度行えば以降の書き込みでは再確認しない。他ツール等で既に Cookie linkage effective state が有効でも、閲覧だけなら ack 不要。画面上の常時警告文・**Cookie linkage official hint** とは別。
_Avoid_: 利用規約同意（アプリ全体の同意と混同しやすいため）, 毎回確認, Effective 表示のための必須ゲート

**Cookie linkage effective state**:
yt-dlp user config 上に Managed cookie options（Cookie 参照行）があるかどうかで決まる、いま実際に効いている有効／方式／参照先。ファイルが無いことも「Managed なし＝無効」として扱い、読み取りエラーにはしない。**Cookie linkage config read failure** のときは無効と偽らない。v1 が編集できない形（コンテナ指定・未対応ブラウザ名・`--cookies` と `--cookies-from-browser` の併記など）でも行があれば **有効**とし、方式は未対応として示す。その状態から v1 の方式を保存すると全 Managed 行を選んだ一方に置換し、無効化すると Cookie 参照行をすべて削除する。動画タブの表示はこれを正とする。書き込み失敗時はエラーを示し、表示を操作前の Effective state に戻す（試した値は Cookie linkage draft に残してよい）。
_Avoid_: アプリ内の有効フラグ（ファイルと食い違う下書きと混同しやすいため）, 未初期化（無効と別状態にしない）, 未対応＝無効（ファイルを隠す印象）

**Cookie linkage config read failure**:
yt-dlp user config のパスに何かあるが読めない状態（権限・ロック・ディレクトリ誤配置など）。ファイル不存在とは別。UI にエラーを示し、Effective を無効とみなさない。この間の書き込み操作は行わない（失敗として止める）。
_Avoid_: Effective＝無効, サイレントフォールバック, 自動修復

**Cookie linkage draft**:
動画タブ上で覚える、方式・ブラウザ・cookies ファイルパスなどの入力下書き、および Cookie linkage risk acknowledgment。無効中でも前回の選択を残してよい。有効時の変更は即時に yt-dlp user config へ書き込み、Cookie linkage effective state と揃える。Cookie ファイルの作成・エクスポート、ブラウザ起動中のロック自動検知、yt-dlp／動画再生の成否確認は含まない。
_Avoid_: 保存済み設定（未書き込みの下書きだけを指す印象）, 適用待ち（明示適用ボタン前提の印象）, Cookie エクスポート, ロック監視

**Cookie linkage unsupported form**:
Cookie linkage effective state が有効だが、v1 UI（chrome / edge / firefox 既定、または単純な cookies ファイルパス）では再現・編集できない Managed cookie options の形。表示上は未対応として示し、保存または無効化で v1 の単純形／削除へ寄せる。
_Avoid_: 壊れた設定, パースエラー（読み取り失敗と混同しやすいため）

## Config

VRChat の `config.json` を編集する画面体験の用語。Settings の DB メンテナンスや yt-dlp 用キャッシュとは別系統。**Asset cache clear**（[Issue #11](https://github.com/JO3QMA/vrctweaker/issues/11)、[ADR 0011](docs/adr/0011-asset-cache-clear.md)）は grill-with-docs で合意済み。実装契約は ADR を正本とする。

### Language

**VRChat asset cache**:
VRChat クライアントがワールド・アバター等のダウンロード済みアセットを置くディレクトリ。`config.json` の `cache_directory` で指定し、空ならプラットフォーム既定の **Cache-WindowsPlayer**（VRChat データディレクトリ配下。Picture folder の既定解決と同型のヘルパ）。容量上限は `cache_size`（GB）、有効期限は `cache_expiry_delay`（日）。**Cached VRChat user**（users_cache）や **Official yt-dlp cache** とは別物。データディレクトリ本体（config.json / Tools）とは別パス。
_Avoid_: キャッシュ（対象が曖昧なため）, フレンドキャッシュ, users_cache, yt-dlp キャッシュ

**Asset cache clear**:
解決済みの VRChat asset cache ディレクトリの**中身をすべて削除**する操作。ディレクトリ自体は残す。Picture folder・Tools・`config.json`・Tweaker DB（users_cache 等）は対象外。VRChat 本体の `cache_size` / `cache_expiry_delay` による自動間引きとは別導線。**VRChat クライアント起動中は実行しない**（拒否して終了を促す）。既存のプロセス検知（`VRChatRunning`）を用いる。操作導線は **Config** のキャッシュ設定カード内（`cache_directory` 編集の近く）。Settings の DB メンテナンスには置かない。削除対象パスの正本は**ディスク上に保存済みの `config.json`**（Config 画面の未保存入力は使わない）。`cache_directory` が空ならプラットフォーム既定パスへ解決する。実行前は **1 回の確認ダイアログ**（解決済み絶対パスと、次回以降の再ダウンロードが必要になる旨。所要時間が長くなりうることも示す）。Settings の DB クリア系と同型で、二段階確認やパス再入力は求めない。解決パスのディレクトリが**存在しない**ときはエラー（作成はしない）。**空ディレクトリ**は成功（削除対象なし）。削除の**途中失敗はエラーで止め、成功扱いにしない**。既に消えた分のロールバックはしない。残りはディスクに残り、再実行で続けて消せる。実行 UI は **同期 API + ボタン loading**（進捗％・キャンセルは v1 では持たない）。実行前に **ボリュームルート**・**非ディレクトリ**・**空パス**を拒否する。解決済みパスが **Picture folder**（保存済み `picture_output_folder`、空なら既定写真パス）と同一なら拒否する（スクリーンショット誤削除防止）。解決済みパスが **VRChat データディレクトリ**（config.json がある LocalLow/VRChat/VRChat 相当）と同一なら拒否する（config / Tools 誤削除防止）。成功時は**削除したエントリ数**を返す（Settings のクリア系と同型）。解放バイト数は出さない（使用量表示は #208 側）。実行中に Config から離れても**バックエンドの削除は完走**する。フロントは unmount 後に loading／トーストを更新しない。
_Avoid_: キャッシュ削除（対象が曖昧なため）, Clear friends cache（Settings の users_cache クリア）, VACUUM

**Asset cache clear v1 scope**:
Issue #11 で最初に届ける Asset cache clear の範囲。Config キャッシュ設定カードからの全中身削除（起動中拒否・パスガード・1 回確認・件数フィードバック）に限定する。v1 では含めないもの: 進捗％・キャンセル、解放バイト数表示、`cache_size` までの間引きや選択削除、Settings への二重導線、Official yt-dlp cache / Tweaker DB / Picture folder の削除、VRChat 起動中の強制クリア、ホーム／デスクトップ等のヒューリスティック拒否。
_Avoid_: Asset cache clear（v1 範囲を指すときは scope とセットで書く）, 将来拡張（スコープ外リストの総称として曖昧なため）

**Config**:
VRChat の `config.json` を閲覧・編集する画面体験。キャッシュ設定・写真出力などクライアント設定の編集が主目的。Tweaker 自身の Settings（ログイン・DB メンテナンス・パス設定）とは別。
_Avoid_: 設定画面（Settings と混同しやすいため）, VRChat 設定（対象ファイルが曖昧なため）

## Automation

Tweaker が監視するイベントや時刻に応じて、VRChat 操作や OS 操作などを自動実行する画面体験の用語。**Automation platform**（[ADR 0012](docs/adr/0012-automation-platform.md)、[Issue #225](https://github.com/JO3QMA/vrctweaker/issues/225)）は grill-with-docs で合意済み。実装契約は ADR を正本とする。

### Language

**Automation**:
サイドバーから開くオートメーション画面体験。**Automation item** を一覧・作成・編集し、有効／無効を切り替える。一般利用者向けの宣言的ルールと、エンジニア向けの Lua スクリプトを同じ一覧に載せ、同じイベント源とアクションカタログを共有する。**Automation rule** の編集は **Automation rule builder**（セクション型）を用いる。
_Avoid_: オートメーション画面, ルール画面（スクリプトを含意しないため）, マクロ（Launch profile や手動起動と混同しやすいため）

**Automation rule builder**:
Automation rule を GUI で組み立てる体験。v1 では「いつ（トリガー）」「もし（条件）」「したら（アクション列）」の**セクション型カード**とする（ドラッグ式ノードエディタは将来）。既存の単純フォーム要素は流用してよい。視覚的ブロック／ノードエディタは v1 スコープ外。
_Avoid_: ワークフローエディタ, ノードエディタ（v1 で含意しないため）, Automation script エディタ（Lua 側と混同しやすいため）

**Automation item**:
Automation に並べる自動実行の設定単位。GUI で作る **Automation rule** と Lua で書く **Automation script** の総称。一覧では種別を区別して示し、有効／無効は item 単位。1 item は rule **または** script のどちらか一方のみ（同居しない）。
_Avoid_: ルール（スクリプトを含意しないため）, ワークフロー（実装のグラフ構造を連想させるため）

**Automation rule**:
Automation item のひとつ（`kind: rule`）。トリガー・条件・アクションを GUI のフォーム（将来はビルダー）で宣言的に定義する。保存形式は構造化データ（JSON 等）。Lua コードは持たない。script への変換や同居は v1 では行わない。
_Avoid_: オートメーション（画面体験全体）, Automation script（Lua 側と混同しやすいため）

**Automation script**:
Automation item のひとつ（`kind: script`）。Lua でイベント購読・条件分岐・アクション呼び出しを記述する。rule で表現しきれない複合ロジック向けに**新規作成**する。同じイベント源とアクションカタログを使う。rule との双方向変換は v1 では持たない。
_Avoid_: プラグイン, 拡張機能（アプリ全体のモジュールと混同しやすいため）, Automation rule（宣言的ルールと混同しやすいため）

**Automation trigger**:
Automation rule が「いつ評価を走らせるか」を決める部分。評価の起点は **Automation event**（イベント駆動）。スケジュール（毎日 0:00 など）も event の一種として扱い、ログ由来・Pipeline 由来・時刻・プロセス状態変化などを同じイベントバスで配信する。
_Avoid_: 条件（真偽判定の本体と混同しやすいため）, トリガー＝スケジュールのみ（イベント駆動を含意しないため）

**Automation condition**:
Automation trigger により評価が始まったとき、**アクションを実行してよいか**を決める追加の真偽条件。複数あるときは AND（すべて真）が既定。例: スケジュール event のあと「VRChat 起動中」が真か。イベント payload のフィールド一致も条件の一種。
_Avoid_: トリガー（評価の起点と混同しやすいため）, フィルタ（Gallery や Activity の絞り込みと混同しやすいため）

**Automation event**:
Automation の評価を起こしうる Tweaker 内の出来事。ログ tail 由来（例: フレンド参加）、Pipeline 由来（例: フレンドのプレゼンス変化）、スケジュール tick、VRChat プロセス状態の変化など。**Log replay** 中は発火しないログ由来 event がある（Activity の **Log replay** 定義に従う）。
_Avoid_: ParsedEvent（実装型名）, トリガー（起点と同一視しない。event が流れ、rule が反応する）

**Automation action**:
条件を満たしたときに実行する操作。VRChat プレゼンス変更、電源プロファイル切り替えなど。Automation item あたり 1 件以上を持てるかは別途決める。
_Avoid_: Launch（手動起動と混同しやすいため）, Quick status（Dashboard の手動操作と混同しやすいため）

**Automation event catalog**:
Automation で購読・トリガーに指定できる event 種別と payload 形状の公式一覧。event 名は安定した識別子（例: `friend_joined`, `schedule.tick`, `vrchat.process`）とし、破壊的変更は避け追加で拡張する。GUI のトリガー選択肢と Lua の subscribe 可能一覧の正本。
_Avoid_: トリガー一覧（条件・アクションを含意しないため）, ParsedEvent 一覧（実装内部の型名）

**Automation event catalog v1 scope**:
最初に届ける Automation event の範囲。**ログ tail 由来**の `friend_joined`、**スケジュール** tick（**Schedule rule** に従う）、**VRChat プロセス**の起動／終了（状態変化）に限定する。Pipeline 由来・Play session・ログイン状態などはカタログ追加で後続。ログ由来は **Log replay** では発火しない。
_Avoid_: Automation v1（画面全体や Lua を含む総称として曖昧なため）, 全イベント対応（スコープ外の期待を招くため）

**Schedule rule**:
スケジュール系 Automation event をいつ発火させるかの宣言。v1 では**曜日（複数選択可）＋時刻（時・分）**を **端末ローカルタイムゾーン**で指定する（例: 平日 23:00、土日 2:00）。将来は cron 相当（5 フィールド）へ拡張するが、v1 では GUI は曜日＋時刻に限定する。
_Avoid_: cron（v1 でフル cron を含意しないため）, タイマー（単発遅延実行と混同しやすいため）

**Schedule tick**:
Schedule rule の条件を満たした瞬間に配信する **Automation event**（カタログ上 `schedule.tick`）。同一分に複数 item が該当しうる。評価順序は別途決める。
_Avoid_: ポーリング, バックグラウンドジョブ（実装の定期処理と混同しやすいため）

**Automation action catalog**:
Automation で実行可能な操作種別と payload 形状の公式一覧。action 名は安定した識別子（例: `change_status`, `set_power_plan`）とし、event catalog と同様に破壊的変更を避け追加で拡張する。GUI のアクション選択肢と Lua から呼べる API の正本。各 action は対応 **platform**（OS）を宣言し、非対応環境では選べない／実行時エラーとする。
_Avoid_: アクション（単発の実行インスタンスと混同しやすいため）, 機能一覧（Automation 外の画面操作を含意しうるため）

**Automation action catalog v1 scope**:
最初に届ける Automation action の範囲。既存の VRChat プレゼンス変更（`change_status`）と、Windows の**電源プラン切り替え**（`set_power_plan`）に限定する。音量・ディスプレイなどはカタログ追加で後続。
_Avoid_: Automation v1（event 側の scope と混同しやすいため）

**VRChat window resize**:
Automation trigger 発火時に、起動中の VRChat クライアントの**OS ウィンドウサイズ**を外部（ウィンドウ API）から変える Automation action（[Issue #12](https://github.com/JO3QMA/vrctweaker/issues/12)）。指定は**幅×高さのピクセル数値のみ**（位置は変えない。プリセットや「現在の半分」は持たない）。Launch profile の起動引数や SteamVR／GPU 側の制限とは別物。朝に戻す等は別 trigger／別数値の rule で足り、専用の restore は持たない。
_Avoid_: 解像度変更（ディスプレイや Launch profile と混同しやすいため）, FPS 制限, ウィンドウ操作（OS 全般と曖昧なため）, Window resize preset, restore point, 位置指定

**Exclusive fullscreen**:
ディスプレイをクライアントが占有する、いわゆる本物のフルスクリーン。外部からのウィンドウリサイズは原理的にできない。ボーダーレス（仮想フルスクリーン）や通常／最大化ウィンドウとは別。
_Avoid_: 仮想フルスクリーン, ボーダーレス, 最大化（リサイズ可能なことが多いため）

**VRChat window resize applicability**:
**VRChat window resize** の実行可否。VRChat が起動しており、主ウィンドウに対してサイズ変更が**可能なら行う**。**最大化中は最大化を解除してから**指定サイズへ変える。**Exclusive fullscreen** でサイズを変えられないときは失敗（サイレント成功にはしない。ボーダーレスと誤判定しやすいため）。可視の主ウィンドウを持つ VRChat プロセスが複数あるときは失敗。未起動や窓が特定できないときは失敗。
_Avoid_: 常に成功, Exclusive fullscreen の強制解除, 最大化をスキップ（省エネにならないため）, プロセス数だけで複数扱い（ランチャー等の副プロセスを誤検知しやすいため）

**VRChat window resize v1 scope**:
Issue #12 で最初に届ける範囲。Automation からの **VRChat window resize**（**Windows のみ**、幅×高さ数値、最大化は解除してから変更）に限定する。**Vendor graphics throttle**（[Issue #249](https://github.com/JO3QMA/vrctweaker/issues/249)）・Launch profile のスケジュール書き換え・ひな形 rule・専用 restore は含めない。
_Avoid_: VRChat window resize（v1 範囲を指すときは scope とセットで書く）

**Vendor graphics throttle**:
GPU ドライバの FPS 制限や SteamVR supersampling など、ベンダー依存の描画負荷下げ。[Issue #249](https://github.com/JO3QMA/vrctweaker/issues/249)。**VRChat window resize**（[Issue #12](https://github.com/JO3QMA/vrctweaker/issues/12)）とは別 Issue。
_Avoid_: VRChat window resize, set_power_plan

**Automation platform scope**:
Automation の event / action がどの OS で有効かの方針。v1 では評価基盤（rule・script・スケジュール・ログ event・`change_status`）は**クロスプラットフォーム**、`set_power_plan` および **VRChat window resize** は **Windows のみ**。カタログの platform 宣言に従い、GUI は非対応 action を出さないか無効化し、Lua は実行時に明示エラーとする。
_Avoid_: Windows 版 Automation（タブ全体が Windows 限定という意味ではない）, プラットフォーム設定（Settings のパス設定と混同しやすいため）

**Automation action sequence**:
1 つの Automation item が持つ複数 **Automation action** の実行順。v1 ではリスト順に**直列実行**する（前のアクション完了後に次へ）。いずれかが失敗したときは**既定で残りを実行しない**（item 全体を失敗扱い）。action ごとに `continue_on_error` を指定した場合のみ、失敗後も次へ進める。
_Avoid_: ワークフロー, パイプライン（分岐・並列を含意しうるため）

**Continue on error**:
**Automation action sequence** 内の 1 アクションが失敗したあと、後続アクションを実行するかどうかの指定。未指定時は偽（失敗で停止）。GUI と Lua で同じ payload フィールドを使う。
_Avoid_: スキップ, リトライ（失敗を成功扱いにしないため）

**Power plan**:
Windows の電源プラン（`powercfg` が扱うスキーム）。**Launch profile**（VRChat 起動引数）とは無関係。Automation の `set_power_plan` が切り替える対象。
_Avoid_: 電源プロファイル（Launch profile と混同しやすいため）, プロファイル（どちらの profile か曖昧なため）

**Power plan preset**:
GUI 向けの抽象ラベル（例: 省エネ・バランス・高パフォーマンス）。実行時に OS 上の実プランへ解決する。マシンに該当プランが無いときの扱いは別途決める。
_Avoid_: 電源モード（OS 設定全般と混同しやすいため）

**Detected power plan**:
その PC で `powercfg` 等により列挙した実電源プラン（表示名と GUID）。GUI の詳細選択と Lua の GUID 指定に使う。
_Avoid_: プリセット（抽象ラベルと混同しやすいため）

**Power plan selection**:
`set_power_plan` を GUI で指定するときの体験。既定は **Power plan preset** から選び、詳細では **Detected power plan** 一覧から選べる。Lua は preset キーまたは GUID を payload に渡せる。
_Avoid_: 電源設定画面（Windows コントロールパネルと混同しやすいため）

**Automation script API**:
Automation script から使える Lua 表面。v1 では (1) **Automation event** の `subscribe`、(2) **Automation action catalog** 経由の `actions.run`（OS 操作の直叩きは不可）、(3) Tweaker 状態の**読み取り専用** API（フレンド・Play session・ログイン状態など）に限定する。ファイル IO・任意 HTTP・シェル実行は v1 では許可しない。追加能力は action または読み取り API のカタログ拡張で届ける。
_Avoid_: プラグイン API（アプリ全体の任意拡張と混同しやすいため）, フル Lua（サンドボックスなしを含意するため）

**Automation run log**:
Automation item が評価・実行された結果の記録。v1 では Automation 画面に**直近 N 件**（目安 20〜50）を示す。成功／失敗、時刻、item 名、アクション完了数、関連する **VRChat 表示名**（例: `friend_joined` の対象）を載せてよい。**`usr_*` ユーザー ID は UI に載せない**。永続ページング・OS 通知・長期保管は v1 スコープ外。公開成果物（PR/Issue）では **Redacted reproduction** に従い表示名も抽象化する。
_Avoid_: Activity ログ, output_log（遭遇ログや生ログと混同しやすいため）, 監査ログ（v1 で長期保管を含意しないため）

**Automation runtime**:
Automation が評価・実行される前提条件。v1 では **VRCTweaker プロセスが起動している間のみ**有効（トレイ常駐を含む）。OS サービスやスケジューラ単体でのバックグラウンド実行は v1 スコープ外とし、UI では起動中のみ有効である旨を示す。将来のサービス化は別判断。
_Avoid_: 常時実行, バックグラウンドサービス（v1 で含意しないため）, Tweaker 終了後も動く（誤期待を招くため）

## Agent contribution

Issue・PR・コミットなど Git に残るテキストを書くときの用語。

### Language

**Public contribution artifact**:
Git 履歴や GitHub 上に公開される成果物。Pull request・Issue・コミットメッセージ・ブランチ名、および Agent がそれら向けに生成する下書きや `docs/ai_dlc/` の Issue メモを含む。VRChat 上の実在ユーザーを特定できる情報を載せない対象。
_Avoid_: 公開物, Git テキスト（スコープが曖昧なため）

**Redacted reproduction**:
Public contribution artifact に書くバグ再現・検証記述。VRChat 表示名・`usr_*` ID・プロフィール URL・ログイン username・インスタンス文字列内の user ID を使わず、件数・ステータス・手順の抽象（例: offline フレンドがキャッシュに無い）で述べる。詳細ルールは `docs/agents/redaction.md`。
_Avoid_: 匿名化, 個人情報マスク（置換手順まで含意しないため）
