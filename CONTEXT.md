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

起動・自分のプレゼンス変更・公式サービス健全性をまとめるホーム画面体験の用語。**Server status**（[Issue #10](https://github.com/JO3QMA/vrctweaker/issues/10)、[ADR 0009](docs/adr/0009-dashboard-server-status.md)）は grill-with-docs / grill-review-ready で合意済み。**Presence change section**（[Issue #174](https://github.com/JO3QMA/vrctweaker/issues/174)、[ADR 0014](docs/adr/0014-dashboard-presence-change.md)）は grill-with-docs で合意済み。実装契約は各 ADR を正本とする。

### Language

**Dashboard**:
サイドバー先頭（`/`）のホーム画面体験。**Dashboard launch block**、**Presence change section**、**Server status** を置く。
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

**Presence change section**:
Dashboard 上でログイン中ユーザー自身の VRChat プレゼンス（色＋ステータス文章）をまとめて編集・反映する UI ブロック（[Issue #174](https://github.com/JO3QMA/vrctweaker/issues/174)）。**Presence change draft**（色選択・文章入力）と反映操作を 1 か所に置く。旧 Quick status / Custom status / Templates の 3 カード構成は置き換える。**Server status** とは無関係。
_Avoid_: Quick status パネル（旧称・3 分割構成を含意するため）, ステータスカード（Server status section と混同しやすいため）, クイックステータス（インフラ健全性と混同しやすいため）

**Presence change draft**:
Presence change section 内の、反映前の編集中状態。選択中のプレゼンス色（join me / active / ask me / busy のいずれか）とステータス文章（`statusDescription`）の 2 フィールドのみ。Quick status・Custom status・Templates がそれぞれ独立した状態を持たない。色の選択操作は draft の色だけを更新し、**VRChat API には送らない**（即時反映しない）。
_Avoid_: クイックステータス, カスタムステータス, テンプレート状態（旧 UI の分割を含意するため）

**Status description history**:
Presence change section の文章入力で選べる、過去に **Presence change apply** で反映したカスタム文章（trim 後非空）の一覧。`app_settings` の JSON 配列（キー例: `presence_description_history`）に永続化する。完全一致は重複排除し、最新を先頭に、最大 20 件。色（`status`）は履歴に含めない。旧 Dashboard Templates（i18n 固定 3 件）は廃止し、初回シードもしない。
_Avoid_: テンプレート, Templates パネル（独立 UI・色付きボタン群を含意するため）, ステータス履歴（色＋文章のセットを含意しうるため）

**Status description suggestions**:
Status description history を Presence change draft の文章欄へ提示する UI。入力欄の候補ドロップダウン（例: `el-autocomplete`）で、フォーカス時または入力時に履歴を示す。候補を選んでも **draft の文章だけ** を更新し、色は変えない。反映は **Presence change apply** まで送らない。
_Avoid_: テンプレートボタン, クイック入力（色と文章を同時に適用する旧 Templates を含意するため）

**Presence change sync snapshot**:
Presence change section が「変更なし」とみなす基準となる、最後に draft に取り込んだ self 行の `status` と `statusDescription` の組。**Presence change prefill**（マウント時）と **Presence change apply** 成功後の再取得で更新する。draft がこのスナップショットと一致するとき **Presence change apply** は無効（disabled）。
_Avoid_: プリフィル（初回だけを指す印象）, 現行ステータス（編集中 draft と混同しやすいため）

**Presence change draft dirty**:
ユーザーが色または文章を操作して、**Presence change sync snapshot** から draft がずれた状態。dirty の間は Pipeline 等による self 行の外部更新で draft を上書きしない。dirty でなければ self 行の更新で draft とスナップショットを再同期してよい。
_Avoid_: Unsaved edits（Launcher や Config の未保存編集と混同しやすいため）, テンプレート選択中（旧 UI）

**Presence change labeling**:
Presence change section の UI 文言の i18n キー名前空間。`dashboard.presenceChange.*` に統一する。旧 `dashboard.quickStatus` / `customStatus` / `templates*` は削除し、全ロケールを同期する。プレゼンス色のラベルは VRChat クライアントに近い表記（Join Me / Active / Ask Me / Busy 等）をロケールごとに示す。
_Avoid_: dashboard.quickStatus（旧 UI・旧称）, ステータス（Server status と混同しやすいため）

**Presence change self-cache sync**:
Cached VRChat user（self 行）が **Presence change section** 外で更新されたとき、**Presence change draft dirty** でなければ **Presence change section fetch** で draft と **Presence change sync snapshot** を再同期する。契機はバックエンドが self 行を書き換えたとき（Pipeline の self `user-update`、**Presence change apply** 成功、**Self profile refresh** 等）に発火する Wails イベントをフロントが購読し、debounce して再取得する（目安 300ms）。dirty の間は再同期しない。
_Avoid_: ポーリング, vrchat:friends-changed（フレンド一覧更新と混同しやすいため）, 常時上書き（draft dirty 無視）

**Presence change section fetch**:
Presence change section の初期表示と再同期用データ取得。**1 回のバックエンド呼び出し**で、ログイン可否・self の `status` / `statusDescription`・**Status description history** をまとめて返す（**Dashboard launch block** の読み取りと同型）。infra 失敗時は **Presence change fetch failure**。
_Avoid_: isLoggedIn の単独問い合わせ＋履歴の別 API（呼び出しが分散するため）, GetCurrentUser（汎用プロフィール取得の俗称）

**Presence change apply**:
Presence change draft の色と文章を VRChat へ送る操作（反映ボタン）。バックエンドの **1 メソッド**で VRChat API 更新（`status` + `statusDescription`）と、成功時のみの **Status description history** 追記（trim 後非空・重複排除・最大 20 件）を行う。文章欄が空白（trim 後）のときは `statusDescription` に **空文字** を送り、VRChat クライアントの色別既定表示に任せる。成功後は返却または再取得で draft と **Presence change sync snapshot** を同期する。draft がスナップショットと同一のときは操作不可（disabled）。操作失敗時は **`ElMessage.error`**（Launch block の起動失敗と同型）。読み取り失敗とは別契約。
_Avoid_: setStatus / setStatusDescription の Dashboard からの分割呼び出し, フロントのみでの履歴永続化, デフォルト文言の明示送信

**Presence change fetch failure**:
ログイン済みなのに Cached VRChat user（self 行）の取得に失敗したときの扱い。Presence change section は残し、カード内に取得できなかった旨だけを示す（**Dashboard launch block** の loadError と同型）。loading 中は取得中メッセージを示す。失敗時はフォームを出さない（またはすべて disabled）。`ElMessage` は出さない。再試行は Dashboard 再表示（`onMounted`）やログイン／プロフィール更新後の再取得に委ね、v1 では手動リトライボタンは置かない。
_Avoid_: 空フォームでの編集, ElMessage のみ（他 Dashboard ブロックと表現がずれるため）, Server status fetch failure（別データ源）

**Presence change visibility**:
Presence change section の表示条件。Dashboard を開けば **常にセクションを表示**する（**Server status visibility** と同様、未ログインでも非表示にしない）。未ログイン時は色・文章・反映をすべて無効化し、カード内に短い説明と **Settings** への導線を示す。ログイン済みのときだけ **Presence change prefill** と操作を有効にする。
_Avoid_: ログイン時のみ表示（Server status との並びが崩れるため）, 非表示（機能の存在が分からないため）

**Presence change prefill**:
Presence change section 表示時に **Presence change draft** と **Presence change sync snapshot** を Cached VRChat user（self 行）の現状で初期化すること（ログイン済みのみ）。色は `status`、文章欄は `statusDescription` をそのまま示す。API 上 `statusDescription` が空のときは文章欄も空（VRChat 既定文言の表示は欄に埋めない）。反映成功後は self 行を再取得して draft とスナップショットを同期する。
_Avoid_: 空の下書き, テンプレート初期値（旧 Templates カードの固定 3 件を含意するため）

**Quick status**:
（旧称）Dashboard 上のプレゼンス即時変更ボタン群。Issue #174 以降は **Presence change section** に統合し、単独の用語としては使わない。
_Avoid_: ステータス, Server status, クイックステータス（インフラ健全性と混同しやすいため）

## Launcher

VRChat の起動引数を名前付きで保存し、起動に使うための用語。

### Language

**Launcher**:
Launch profile を一覧・作成・複製・編集・保存・削除する画面体験。主目的は起動引数の編集と保存であり、VRChat の起動（Profile launch）は副次の導線。
_Avoid_: ランチャー画面, 起動画面（Quick launch / Profile launch と混同しやすいため）

**Launch profile**:
Tweaker が保存する起動設定のまとまり。表示名、起動引数の文字列、既定かどうか（`isDefault`）を持つ。Launcher 画面で編集し、Dashboard の起動ブロック（Quick launch / Instance rejoin）や World join のベース引数になる。Launch profile が 1 件も無いときの初回シードでは、表示名 **Desktop**（`--no-vr`・Default）と **VR**（`-no-vr` なし）の 2 件を用意する。既に 1 件以上ある DB への後追い追加はしない。
_Avoid_: プロファイル（VRChat profile slot と混同するため）, preset 単体, Non-VR（シード表示名は Desktop）

**Draft launch profile**:
Launcher 上でまだ DB に保存していない Launch profile（`id` が空）。サイドバー一覧には行が無く、Unsaved launch profile edits がある間はエディタ上部バナーで示す。別 Launch profile への切り替えやルート離脱時は確認ダイアログを出し、破棄すればドラフトは消える。**「+ 新規プロファイル」はこの状態を作らない**（即保存の Launch profile を作る。作成＝保存）。Draft 自体は別導線・将来用の概念として残す（現行 Launcher に空 `id` 編集 UI は置かない）。
_Avoid_: 新規プロファイル, 仮プロファイル（保存済みとの境界が曖昧なため）

**Launch profile create**:
Launcher の「+ 新規プロファイル」押下で、既定の表示名と起動引数をもつ **Launch profile を即保存する**こと（[ADR 0013](docs/adr/0013-launch-profile-create.md)、[Issue #210](https://github.com/JO3QMA/vrctweaker/issues/210)）。作成直後は一覧に現れ選択状態になり、Unsaved launch profile edits／Draft launch profile にはならない。作成後の名前や引数の変更は、既存 Launch profile の編集と同じく明示保存。既定の表示名はロケールの既定文字列（例: 「新しいプロファイル」）を使い、同名が既にあれば「… 2」「… 3」…と連番を付けてぶつからない名前にする（1 件目に番号は付けない。空きは既定名 → `既定名 + " " + n`（n≥2）の最小）。名前の一意制約は設けない（手動リネームでの同名は許す）。起動引数の初期値は空／GUI 既定（Primary・Advanced ほぼオフ）とし、Default launch profile や選択中 profile の複製はしない（引き継ぎは **Launch profile duplicate**）。保存に失敗したときは作成されなかったものとして扱い、一覧に新行は出さず選択も変えない（エラーはユーザーに示す）。保存成功時のトーストは出さない。保存済み Launch profile が 0 件のときは、作成する 1 件目を **Default launch profile**（`isDefault` 真）にする。既に 1 件以上あるときの作成では既定フラグは付けない。
_Avoid_: Draft launch profile（未保存新規）, 新規ドラフト, create 経路での複製（**Launch profile duplicate** と混同しない）

**Launch profile duplicate**:
Launcher で選択中の **Launch profile** の設定を引き継ぎ、新しい Launch profile として即保存すること（[ADR 0015](docs/adr/0015-launch-profile-duplicate.md)、[Issue #209](https://github.com/JO3QMA/vrctweaker/issues/209)）。**Launch profile create**（空／GUI 既定からの新規）とは別操作。複製直後は一覧に現れ選択状態になり、編集可能な状態になる。Unsaved launch profile edits／Draft launch profile にはならない（作成＝保存と同型）。**Unsaved launch profile edits** があるときは **Launch profile create** や切替と同じく、先に保存／破棄／キャンセルを求める。キャンセルなら複製しない。解決後は選択中 profile の**保存済み内容**（表示名・起動引数＝Primary / Advanced）を引き継ぐ。表示名は元の表示名を基点にし、**Launch profile create** と同じ連番ルール（基点が空いていればそのまま、埋まっていれば `基点名 + " " + n`、n≥2 の最小）でぶつからない名前にする。複製先の `isDefault` は**常に偽**（元が Default launch profile でも既定は移さない・付けない）。削除のような追加の確認ダイアログは出さない（未保存ガード以外）。保存成功時のトーストは出さない（一覧出現・選択がフィードバック）。保存に失敗したときは複製されなかったものとして扱い、一覧に新行は出さず選択も変えない。エラー表示は **Launch profile create**／明示保存と同じ（`ElMessage.error`）。
_Avoid_: Launch profile create, コピー（表示名の接尾辞だけを指す印象）, プロファイル複製（create との境界が曖昧なため）, 未保存のまま複製（create／切替と未保存ガードがずれるため）, 複製で Default を付け替える, 複製の確認ダイアログ（削除と混同しやすいため）

**Launch profile editor actions**:
選択中の Launch profile があるとき、エディタ上部に並べるラベル付き操作。左から **Profile launch**・明示保存・**Launch profile duplicate**・削除。削除は破壊操作として区別して示す。「⋯」オーバーフローは使わない（削除・複製を隠さない）。
_Avoid_: ⋯ メニュー, more actions（削除・複製を隠す導線）

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
Launcher エディタの折りたたみ内にまとめる起動引数。解像度、モニター、FPS、優先度、VRChat profile slot、デバッグ・MIDI、**IK 2.0 launch options** など。Primary に含まれないものはすべてここに属する。
_Avoid_: 詳細設定, すべてのオプション（UI ラベルは可。ドメインでは Advanced と書く）

**IK 2.0 launch options**:
Launch profile の起動引数のうち、VRChat 公式が IK 2.0 向けに列挙する起動オプション（[Issue #17](https://github.com/JO3QMA/vrctweaker/issues/17)、[公式ドキュメント](https://docs.vrchat.com/docs/ik-20-features-and-options)）。grill-with-docs / grill-review-ready で合意済み。対象は `--custom-arm-ratio`・`--disable-shoulder-tracking`・`--enable-ik-debug-logging`・`--calibration-range`・`--freeze-tracking-on-disconnect` の 5 本。クライアント内の IK Legacy／2.0 切替・Avatar Measurement・Lock Types・Locomotion Animation は起動引数ではないため含めない。既存の `--ignore-trackers`（Ignore trackers）は別オプションのまま据え置く。既知の IK 引数が Custom 欄にあれば、他の既知引数と同様に Parse で GUI フィールドへ吸い上げ、Custom からは外す（二重出力しない）。Parse で既知プレフィックスだが値が不正なトークンは、既存の他引数と同様フィールドにも Custom にも残さず捨てる。
_Avoid_: IK 設定（クライアント内トグルを含意するため）, Full-body 設定, ignore-trackers（別の既存 Advanced 項目）, Custom に残したまま GUI にも出す

**IK launch options section**:
Launcher の **Advanced launch options** 内に置く第 3 ブロック（「表示・性能」「デバッグ・上級」に続く）。見出しは日英とも **「IK 2.0」**（`IK 2.0`）。**IK 2.0 launch options** の 5 本だけをまとめる。並びは公式ドキュメントの「New Launch Options」順（`--custom-arm-ratio` → `--disable-shoulder-tracking` → `--enable-ik-debug-logging` → `--calibration-range` → `--freeze-tracking-on-disconnect`）。Ignore trackers は「デバッグ・上級」に残し、このブロックへ移さない。
_Avoid_: トラッキング設定（Ignore trackers やクライアント内 IK を含意するため）, デバッグ・上級への混在, IK / トラッキング, IK 起動オプション（見出しとしては冗長・Ignore trackers を連想させやすい）, 値付き優先や常用推測の並び

**IK value launch option**:
**IK 2.0 launch options** のうち値が付くもの（`--custom-arm-ratio`・`--calibration-range`）。Launcher では既存の FPS／MIDI と同型で、有効チェックがオフのときは引数自体を出さない。オンにしたときの初期値は公式ドキュメントの既定（arm-ratio `0.4537`／calibration-range `0.6`）をシードする。フラグのみの 3 本（`--disable-shoulder-tracking`・`--enable-ik-debug-logging`・`--freeze-tracking-on-disconnect`）はチェックボックスのみ。有効時の値は**正の実数**（`> 0`）のみ許可し、公式に無い上下限は設けない。有効なのに値が空・0・負・非数のときは、保存を止めず公式既定へ戻す（FPS と同型）。Merge（保存・起動用文字列）の正規形はクォート無しの `=値`（例: `--custom-arm-ratio=0.4537`）。Parse 時は公式ドキュメント形のクォート付きも受理し、GUI フィールドへ取り込む。未設定の内部表現は FPS／解像度と同型（専用 sentinel や文字列保持にはしない）。
_Avoid_: 常時数値欄, 省略不可の固定値, 空欄のまま引数出力, 推測の固定レンジ（例: 0.01–1.0）, 常にクォート付きで書き出す, 読み込み形のまま非正規化で保持, 不正値で保存拒否, 不正値で有効チェック自動 OFF

**IK 2.0 launch options v1 scope**:
Issue #17 で最初に届ける範囲。上記 5 本を **IK launch options section** の GUI で編集・保存できるようにすることに限定する。クライアント内 IK 設定の再現、Ignore trackers の改変・移動、Primary への昇格は含めない。Desktop（`--no-vr`）プロファイルでもセクションは常に表示・編集可（他 Advanced と同型。VRChat 側の無視はユーザー責任）。
_Avoid_: IK 2.0 launch options（v1 範囲を指すときは scope とセットで書く）, 将来拡張（スコープ外リストの総称として曖昧なため）, no-vr 時の非表示・無効化

**Unsaved launch profile edits**:
Launcher エディタで、最後の保存または読み込み以降に加えた Launch profile の変更（名前、既定フラグ、Primary / Advanced の各引数）。保存前はサイドバーの未保存表示とエディタ上部バナーで示す。
_Avoid_: dirty 状態, 未保存（他画面の編集と混同しやすいため）

**Discard launch profile edits**:
Unsaved launch profile edits を保存せず、直前に保存または読み込みした内容に戻すこと。別 Launch profile への切り替え、**Launch profile create**、**Launch profile duplicate**、Launcher 以外の画面への移動の前に確認できる。
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

**Encounter friend mark**:
Encounter log 上で、当該 User encounter の相手が **Encounter log の再取得時点**で Listable friend であることを示す印。遭遇当時のフレンド関係のスナップショットではない（解除後は過去行からも消え、後から Listable friend になった相手は過去行にも付く）。Friends 同期だけでは Encounter log を再取得せず、マークも更新しない。未ログイン・キャッシュ未同期・users_cache に行が無いとき、および判定不能のときは付かない。印は表示名テキストを書き換えない（列の表示名は User encounter＝ログ由来のまま）。表示名の左に固定幅のスロットを取り、非 Listable friend 行は空にして名前の左端を揃える。v1 では Encounter log のみ（Encounter history には付けない）。
_Avoid_: フレンドマーク（Friends のお気に入り★と混同しやすいため）, 遭遇時フレンド（当時スナップショットを含意するため）, フレンドアイコン（表示形式の俗称）, リアルタイムフレンド印（Friends 同期即時反映を含意するため）

**Display name filter**:
Encounter log 上の唯一の絞り込み。表示名の部分一致のみ（クライアント側）。ワールドや期間での絞り込みは Encounter history 側に任せる。
_Avoid_: 検索, フィルタ（Gallery の World search や Date range filter と混同しやすいため）

**Activity retention**:
Output log 由来の Activity データの保存上限。設定の保存期間（日）を過ぎた User encounter・Play session・**Video playback attempt** は自動削除される。Activity 画面ではページ全体（タイトル付近）に 1 回だけ期間を示すヒント文を置き、空状態だけに頼らない。Video playback history では履歴カード内に同じ日数の短いヒントを 1 行置く。
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
Output log ingest のうち、すでにディスク上にある行を offset から読み直して Activity の相関状態を再構築すること。起動時 bootstrap を含む。User encounter・Play session・**Video playback attempt** の更新を行い、Friend joined などの automation は発火しない（automation は追記監視の live tail に限る）。Video playback history 向けの変更通知も replay / bootstrap 中は抑制してよい（完了後の再取得に委ねる）。
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

**Video playback attempt**:
output_log の `[Video Playback]` 由来で、ある URL の resolve を 1 回試みた記録（[Issue #176](https://github.com/JO3QMA/vrctweaker/issues/176)）。`Attempting to resolve URL` / `Resolving URL` で始まり、同じ URL に対する成功（`resolved to`）または失敗（`ERROR`）で結果が付く。結果の正は **Video playback outcome**。結果が未観測の間は Open。試行時点に Open play session があればそのワールドを文脈として付与し、無ければワールドは空（試行自体は残す）。直近の終了済み Play session には遡らない。属する Log source を持つ（UI には出さない）。User encounter や Play session と同じく Output log ingest の対象であり、Activity 画面には出さない。
_Avoid_: 再生履歴（一覧全体を指す語）, VideoPlaybackEvent（実装型名）, 再生セッション（Play session と混同しやすいため）, 動画ログ

**Open video playback attempt**:
結果（成功・失敗）が未確定の Video playback attempt。一覧では結果列に未解決ラベルで示す（成功／失敗／欠損の空表示とは区別する）。Open encounter / Open play session と同型。Log rotation handoff やクライアント終了だけでは自動で失敗扱いにしない（結果行が来るまで、または Activity retention で削除されるまで Open のまま）。
_Avoid_: 進行中再生, pending（実装語）, タイムアウト失敗（未観測を失敗と同一視しないため）

**Video playback world context**:
Video playback attempt に付く、試行時点の Open play session 由来のワールド情報。一覧の主表示はワールド表示名（`world_info` 等）。`wrld_*` や VRChat instance key は Encounter log と同様、一覧の主列には出さない。Open play session が無いときの空欄は「不明」ではなく文脈なし。
_Avoid_: ワールド文脈（曖昧なままの Issue 俗称）, 最後のワールド（終了済み session へのフォールバックを含意するため）, インスタンス（Log source や instance key と混同しやすいため）

**Video playback outcome**:
Video playback attempt の結果区分。成功・失敗・未解決（Open）の三値。同じ試行 URL に対し `[Video Playback] ERROR:` が一度でも付いたら失敗とし、その後の `URL '…' resolved to '…'`（入力と同一 URL へのフォールバックを含む）では成功に上書きしない。ERROR が無く `resolved to` だけなら成功。失敗時の理由文は ERROR 行の文言。成功時は解決先 URL を保持してよい。
_Avoid_: 再生結果（プレイヤー再生完了と混同しやすいため）, resolved＝常に成功（ERROR 後の同 URL resolve を成功と誤るため）

**Video playback failure reason**:
失敗した Video playback attempt に付く、ログ上の ERROR 文言。`[Video Playback] ERROR:` 以降のテキスト。**全文を保持し、UI にも全文を出す**（折り返しやツールチップは可。先頭省略や i18n 分類は v1 ではしない）。Public contribution artifact には載せない。
_Avoid_: エラーコード, yt-dlp 終了コード（NativeProcess 行は Video playback attempt の正本にしないため）

**Video playback correlation**:
同一 Log source 内で、試行行（`Attempting` / `Resolving`）と結果行（`ERROR` / `resolved to`）を URL で対応づけること。結果は同じ URL の Open video playback attempt のうち最も古い 1 件に付与する（FIFO）。対応する Open が無い結果行は捨て、結果だけの行は作らない。別 URL の Open は互いに取り違えない。相関状態は Log source ごとに分離し、Open play session と同じ session 相関の寿命を共有する（[ADR 0016](docs/adr/0016-video-playback-history.md)）。
_Avoid_: pending URL（実装語だけ）, セッション相関（Play session / encounter 全体と混同しやすいため）

**Video playback history**:
動画タブ上の、Video playback attempt の時系列一覧。画面見出しは「再生履歴」など。列は時刻・試行 URL・Video playback outcome・Video playback failure reason・Video playback world context（表示名）。**全 Log source の行を試行時刻で混ぜて 1 一覧**とし、Log source 列は出さない（Encounter log と同型）。**Activity retention 内を一括取得して表示**する（Encounter log と同様、v1 ではサーバ側ページングは持たない）。既定は試行時刻の新しい順。ページ上はタイトル直下に置き、yt-dlp Tools replace maintain / Cookie linkage より上（診断を先に見せる）。履歴カード内に Activity retention と同じ保存日数を示す短いヒントを 1 行置く。行からの操作は **試行 URL のコピーのみ**（再試行・外部ブラウザ起動・解決先 URL コピーは持たない）。Activity 画面には置かない。
_Avoid_: 動画ログ, Encounter log（Activity の遭遇一覧と混同しやすいため）, 再生セッション一覧

**Video playback history refresh**:
Video playback history の再取得。Output log ingest が Video playback attempt を更新したあと、専用の変更通知（Encounter 用の `activity:encounters-changed` とは別）で、動画タブ表示中のみ debounce して行う。v1 では手動更新ボタンは置かない（マウント時取得＋上記の自動再取得）。
_Avoid_: Activity refresh（Encounter log / Play time chart 向けと混同しやすいため）, encounters-changed 相乗り

**Video playback history fetch failure**:
Video playback history の一覧取得に失敗したときの扱い。履歴カードは残し、カード内に取得できなかった旨だけを示す（Encounter log / Dashboard launch block と同型）。`ElMessage` は出さない。Output log ingest 中の単一行の永続化失敗は、ログに残してその行だけスキップし、ingest 全体や他種別（User encounter 等）は止めない。
_Avoid_: 履歴非表示, ingest 全体停止, ElMessage のみ

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

VRChat の動画プレイヤーが裏で使う yt-dlp 向けの用語。**yt-dlp Tools replace maintain**（[Issue #9](https://github.com/JO3QMA/vrctweaker/issues/9)、[ADR 0008](docs/adr/0008-ytdlp-tools-replace-maintain.md)）は Accepted・製品実装済み。**yt-dlp Cookie linkage**（[Issue #8](https://github.com/JO3QMA/vrctweaker/issues/8)、[ADR 0007](docs/adr/0007-ytdlp-cookie-linkage.md)）は Accepted・製品実装済み（制限付き再生には Official yt-dlp cache 経由が前提。同梱版単体では Cookie オプション非対応）。起動前ワンショットの直置き試作は [PR #40](https://github.com/JO3QMA/vrctweaker/pull/40)（望む動作に未達）。動画タブ上の両機能の UI 括りは **yt-dlp experimental features**（[Issue #220](https://github.com/JO3QMA/vrctweaker/issues/220)。用語・レイアウト合意は本 CONTEXT。新規 ADR は作らない）。

### Language

**yt-dlp experimental features**:
動画タブ上で **yt-dlp Tools replace maintain** と **yt-dlp Cookie linkage** をまとめる 1 つの UI カード（外側のみ `el-card`。内側はネストしたカードにせず、`<section>` ＋小見出しで区切る）。外側の見出しは日本語「yt-dlp (実験的機能)」／英語 `yt-dlp (Experimental features)`（他ロケールも同趣旨）。カード内は上から **置換（小見出し: 日本語「yt-dlp の置換」／英語 `yt-dlp replacement`）→ Cookie（小見出し: 日本語「Cookie を利用する」／英語 `Use cookies`）** の順。置換ブロックのスイッチ横には機能名ラベル（旧 `replaceLabel`＝「yt-dlp の置換」）を出さない（小見出しと重複させない。effective 状態表示は残す）。初期ロードはブロック単位: 置換の loading は置換ブロックだけを覆い、Cookie は独立に取得・表示する（カード全体の共通 loading や API 結合はしない）。旧カード見出し「yt-dlp を置き換える」「yt-dlp Cookie 連携」／英語旧 `Replace yt-dlp`・`yt-dlp Cookie linkage` は使わない。表示条件は現状どおり: 外側カードは動画タブで常に出し、置換ブロックは非対応時も警告を示し、Cookie ブロックは対応環境（または操作エラー表示が必要なとき）だけ出す。各ブロック直下の常時警告（置換の公式差し替え非推奨文・Cookie の BAN／捨て垢文）は統合せず現状どおり残す。各機能の意味・操作・risk acknowledgment・effective state・**Cookie linkage official hint** の振る舞いは変えない（レイアウトと小見出し文言の直し）。
_Avoid_: 実験的機能（単独・対象が曖昧なため）, yt-dlp section（実験性を落とすため）, Settings / Config（載せない）, yt-dlp を置き換える / yt-dlp Cookie 連携（旧小見出し）, Cookie を上・置換を下（official hint の「上の」前提が崩れるため）, 内側のネスト `el-card`（二重枠）, 両方未対応で外側ごと隠す（置換の非対応案内が消えるため）, カード先頭への警告統合（機能別リスクが曖昧になるため）, 小見出しとスイッチ横の「yt-dlp の置換」二重表示, 外側カード全体を置換 loading で隠す（Cookie 表示を遅らせるため）

**yt-dlp experimental features v1 scope**:
[Issue #220](https://github.com/JO3QMA/vrctweaker/issues/220) で届ける範囲。動画タブの **yt-dlp experimental features** 1 カード化（小見出し・i18n・hint 呼び揃え・スイッチ横機能名ラベル削除・Vitest／Storybook／E2E 追随・死キー整理）に限定する。v1 では含めないもの: Go／usecase／Wails 契約の変更、risk acknowledgment や effective state の意味変更、Settings／Config への移設、常時警告の統合、loading／API の共通化、機能の追加・削除、新規 ADR、ADR 0007／0008 の製品方針書き換え。
_Avoid_: yt-dlp experimental features（v1 範囲を指すときは scope とセットで書く）, 将来拡張（スコープ外リストの総称として曖昧なため）

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
Cookie linkage の UI 上で、Tools replace effective state が偽のときに出す案内。Official（Cookie 対応）exe が Tools から参照されていないと制限付き再生の目的を満たせない旨と、同一 **yt-dlp experimental features** カード内の置換ブロック（小見出し「yt-dlp の置換」／英語 `yt-dlp replacement`）への案内。有効化や yt-dlp user config への書き込みは妨げない。hint 文言内の呼び方は当該小見出しに揃える（英語旧 `Tools replace` 呼びは使わない）。
_Avoid_: 必須ゲート, maintain 必須（ハード依存を連想させるため）, Cookie linkage risk acknowledgment（BAN 警告とは別）, Tools replace（hint 内の旧呼び）

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
動画タブ上で覚える、方式・ブラウザ・cookies ファイルパスなどの入力下書き、および Cookie linkage risk acknowledgment。無効中でも前回の選択を残してよい。有効時の変更は原則 **即時**に yt-dlp user config へ書き込み、Cookie linkage effective state と揃える。例外: **Cookie linkage effective state** が有効のとき **Browser cookie source** と **Cookies file source** のラジオを切り替え、切替先がまだ書き込めない（典型: Cookies file source へ切替だがパス未指定／不存在）場合は、黙って **Disable**（Managed 行削除）して Effective を無効にし、Draft は切替先の方式のまま、トグルは OFF のままとする（追加の案内文は出さない）。切替先が書き込み可能なら即 upsert（Cookies file → Browser、Browser → Cookies file で有効パスあり、など）。Disable が失敗したときはセクション内エラー、ラジオは操作前の Effective へ戻し、トグルは ON のまま。トグル OFF の間の参照は Draft のパス更新のみ（自動再有効化しない）。同一方式内（ブラウザ変更・パス欄編集）はこの例外に含めず、従来どおり即書き込み（失敗時はエラー＋Effective へ戻す）。Cookie ファイルの作成・エクスポート、ブラウザ起動中のロック自動検知、yt-dlp／動画再生の成否確認は含まない（[Issue #219](https://github.com/JO3QMA/vrctweaker/issues/219)）。
_Avoid_: 保存済み設定（未書き込みの下書きだけを指す印象）, 適用待ち（明示適用ボタン前提の印象）, Cookie エクスポート, ロック監視, 方式切替の常時 Disable（書き込み可能な切替先まで巻き込む印象）

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

## Design system

ボタン様式・余白・色・タイポグラフィ・フォーム入力部品などアプリ横断 UI 部品の用語。**VtButton** と 4 **Button variant**（[ADR 0017](docs/adr/0017-button-design-system-vtbutton.md)）は grill-with-docs で合意済み。**Spacing**（余白ルール）（[ADR 0018](docs/adr/0018-spacing-design-system.md)）は grill-with-docs で合意済み。**Color**（カラーパレット）（[ADR 0019](docs/adr/0019-color-design-system.md)）は grill-with-docs で合意済み。**Typography**（タイポグラフィ）（[ADR 0020](docs/adr/0020-typography-design-system.md)）は grill-with-docs で合意済み。**Form control** ラッパー（VtInput / VtSelect / VtCheckbox / VtSwitch）は grill-with-docs で合意済み（[ADR 0021](docs/adr/0021-form-control-design-system.md)）。**Feedback & Badges**（Alert / Toast / Badge / Loading の横断チャネル）は grill-with-docs で合意済み（[ADR 0022](docs/adr/0022-feedback-badges-design-system.md)）。実装契約は各 ADR を正本とする。色コード・hex 値・個別 props など実装詳細はここに書かない。

### Language

**VtButton**:
VRCTweaker の標準操作ボタンコンポーネント。画面に置く **Primary / Secondary / Tertiary / Danger** は VtButton の `variant` で指定する（**必須**。省略時の暗黙既定は持たない）。`el-button` の直接利用は新規・改修時に VtButton へ寄せる。**Semantic button** は VtButton の variant に含めない（専用 UI を別に持つ）。
_Avoid_: ボタン（`<button>` 素要素や `el-button` 直呼びと混同しやすいため）, Button component（汎用 React/Vue の俗称）

**Button variant**:
VtButton で選ぶ 4 様式のいずれか: `primary` / `secondary` / `tertiary` / `danger`。**Primary button** 等の用語と 1 対 1 で対応する。
_Avoid_: type（Element Plus の `type` prop と混同しやすいため）, スタイル（CSS クラスだけを指す印象）

**VtButton adoption**:
VtButton への移行方針。新規画面・改修で触ったファイルでは **VtButton を使う**。**Semantic button** は専用 UI のまま移行対象外。未着手の既存 `el-button` 直書きは一括置換しない（触ったところから順次）。遵守は `.cursor/rules/` で案内し、移行期は ESLint では強制しない。見た目の正本は **VtButton Storybook catalog**。
_Avoid_: 全面置換, el-button 禁止（即時一括・CI 強制を含意するため）

**VtButton Storybook catalog**:
VtButton の Storybook 一覧。4 **Button variant** それぞれについて通常・**Button disabled state**・**Button loading state** を載せ、Action group の並び例（例: Primary + Secondary、Danger 確認 + Secondary キャンセル）も 1 ストーリー以上含める。
_Avoid_: ボタン一覧（Semantic button を含意しうるため）, 全画面 Storybook（コンポーネントカタログに限定）

**Dialog confirm button**:
`ElMessageBox` 等、テンプレートに **VtButton** を置けない確認ダイアログのボタン。**Danger button** / **Secondary button** の見た目に、VtButton 実装と共有する class 定数で揃える。**VtButton adoption** の対象外。
_Avoid_: モーダルボタン（将来の自作ダイアログ置換を含意しうるため）, MessageBox（実装 API 名だけを指す印象）

**Action group**:
画面上でひとまとまりに並ぶ操作ボタンの群。ダイアログのフッター、フォーム下部のボタン列、ツールバー、カード内のアクション行など。**Primary button** の「高々 1 つ」は Action group 単位で数える。
_Avoid_: 画面全体, ツールバー（Action group の一例に過ぎないため）

**Primary button**:
Action group 内で主として進めたい操作を示すボタン様式。同一 Action group には高々 1 つだけ置く。見た目は強調色の塗り。
_Avoid_: メインボタン（色だけを指す印象）, CTA（マーケ用語）

**Secondary button**:
Action group 内の実行操作を示すボタン様式。Primary より優先度は低いが、押すと状態が変わる・処理が走る操作に使う（起動、更新、複製、参照、編集を破棄しないキャンセルなど）。見た目は中立色の塗り。
_Avoid_: サブボタン（見た目だけを指す印象）, デフォルトボタン（実装語）

**Tertiary button**:
補助・低優先の操作を示すボタン様式。テキストリンクに近い軽さでよい。一覧内の軽操作、ログの再読み込み、編集パネルを閉じるだけの操作、**Draft row removal** など。**Danger button** に当たる削除は含めない。見た目は背景なし（テキストリンクに近い）。
_Avoid_: ゴーストボタン（見た目だけを指す印象）, リンク（`router-link` や `<a>` と混同しやすいため）

**Danger button**:
**永続データ**に対する取り消しが難しい破壊操作を示すボタン様式。プロファイル削除、ルール削除、ログアウト、キャッシュ全消去、DB クリアなど。破壊を主目的とする確認ダイアログでは、確認側を Danger としキャンセルを Secondary とする。その Action group には **Primary button** を置かない（Danger が強調）。見た目は警告色の塗り。同一 Action group に **Primary button** があるときは枠線のみの弱い強調も可。
_Avoid_: 警告ボタン（非破壊の注意喚起と混同しやすいため）, 赤ボタン（色だけを指す印象）

**Draft row removal**:
保存前の下書き（フォーム編集中の一覧行など）から 1 行を外す操作。永続ストアからはまだ消えない。**Tertiary button** でよい。確認ダイアログは原則不要。
_Avoid_: 削除（**Danger button** の永続削除と混同しやすいため）, 行削除（永続か下書きかが分からないため）

**Button disabled state**:
操作できない／させたくないボタンに付ける状態。ボタン様式（Primary / Secondary / Tertiary / Danger）ではなく、いずれの様式にも重ねる。実装では標準の `disabled` 属性（または同等）で表す。未ログイン、変更なし、対象なしなど。無効理由はラベル・ツールチップ・周辺文言で伝える。
_Avoid_: Disabled ボタン（第 5 の様式名）, Disabled 様式

**Button loading state**:
非同期処理中に、**押したボタンだけ**に付ける一時状態。処理完了までそのボタンは操作できなくなる（**Button disabled state** と同型の扱い）。別様式にはしない。Action group 内の他ボタンまで一括で無効化するのは原則としない。
_Avoid_: Loading ボタン（様式名）, 画面全体ロック（ボタン状態の話と混同しやすいため）

**Semantic button**:
色や形状そのものがドメイン上の意味を伝えるボタン。4 様式（Primary / Secondary / Tertiary / Danger）の配置ルールは当てはめない。**Button disabled state** とアクセシビリティ（ラベル・`aria-label` 等）は当てはめる。例: **Presence change section** のプレゼンス色ボタン（Join Me / Active / Ask Me / Busy）。お気に入りトグルやアイコンのみの追加ボタンは Semantic button にせず、Tertiary または Secondary に寄せる。
_Avoid_: カラーボタン（装飾だけを指す印象）, 例外ボタン（ルール逃れの総称）

**Spacing**:
アプリ横断の余白ルール。`margin`・`padding`・`gap`（Flex/Grid）に使う数値スケールと、セクション間・カード内・フォーム項目間など繰り返し現れるレイアウト間隔の命名パターンを含む。アイコン寸法・アバターサイズ・行高・最小タップ領域などコンテンツ固有の寸法（Sizing）や角丸（**Border radius**）は含めない。
_Avoid_: マージン（CSS プロパティ名だけを指す印象）, レイアウト（グリッド列数・ブレークポイント等の総称）

**Spacing scale**:
Spacing で使ってよい余白の段階列。最小単位は 4px。許容値は **4, 8, 12, 16, 24, 32, 48, 64**（px）のみ。20 など中間値は持たず、最寄りの段階へ寄せる。新規・改修ではこの列以外の任意 px / rem 余白を増やさない（**Spacing adoption**）。
_Avoid_: 8 の倍数ルール（4 と 12 を含むため）, 連続スケール（任意の 4px 刻みすべてを許す印象）

**Spacing token**:
Spacing scale の各段階を表す CSS カスタムプロパティ。値は **px リテラル**（例: `--space-16: 16px`）。`rem` や `em` でスケールを表現しない。トークン名は段階の px 数と一致させる（`--space-4` … `--space-64`）。アプリの `font-size`（現行 14px ルート）を Spacing のために変更しない。
_Avoid_: rem トークン（14px ルートと噛み合わないため）, spacing-md（数値と名前がずれるため）

**Spacing pattern**:
繰り返し現れるレイアウト間隔に付ける意味別の名前。**Spacing token** のセマンティック・エイリアスとして定義し、実体は必ず `--space-*` へ委譲する（二重の px 値を持たない）。新規・改修では用途に合う pattern があれば数値トークンの直書きより pattern を優先する。
_Avoid_: ユーティリティクラス（`.gap-8` 等のクラス体系を含意するため）, マジックナンバー（pattern なしの任意余白）

**Spacing pattern catalog**:
v1 で定義する **Spacing pattern** の公式一覧（いずれも CSS 変数）。`--space-inline-tight` → 4px、`--space-action-group` → 8px、`--space-form-field` → 12px、`--space-block` → 16px、`--space-section` → 24px、`--space-page` → 32px。48px / 64px は `--space-48` / `--space-64` の数値トークンのみ（pattern 名なし）。
_Avoid_: gap ユーティリティ, 全間隔の pattern 化（低頻度の 48 / 64 は v1 で名前を付けない）

**Spacing adoption**:
Spacing トークン・pattern への移行方針。新規画面・改修で触ったファイルでは **Spacing token** または **Spacing pattern catalog** を使う。未着手の既存 `rem` / 任意 px 余白は一括置換しない（触ったところから順次）。共有スタイル（`style.css` の `.page-title`、`.section-card` 等）は v1 でトークンへ寄せる。遵守は `.cursor/rules/` で案内し、移行期は ESLint では強制しない。
_Avoid_: 全面置換, rem 禁止の CI 強制（即時一括・CI 強制を含意するため）

**Spacing v1 scope**:
Spacing の最初の届け範囲。`style.css` への **Spacing token**・**Spacing pattern catalog** 定義、共有クラスのトークン化、`.cursor/rules/`、ADR、**Spacing Storybook catalog** に限定する。含めないもの: 全 View 一括トークン化、Element Plus 内部余白の上書き、ESLint 強制、Sizing / **Border radius** の再設計、48 / 64 の pattern 名付け。
_Avoid_: Spacing（v1 範囲を指すときは scope とセットで書く）, 将来拡張（スコープ外リストの総称として曖昧なため）

**Spacing Storybook catalog**:
Spacing の Storybook 一覧。数値スケール（4〜64）・**Spacing pattern catalog** の対応表・レイアウト使用例を載せ、見た目と使い方の正本とする（**VtButton Storybook catalog** と同型）。
_Avoid_: 全画面 Storybook, デザイントークン一覧（色・タイポ等の総称）

**Spacing migration rounding**:
既存の `rem` や任意 px 余白を **Spacing token** へ置き換えるとき、スケールに無い値は **四捨五入で最寄りの段階**へ寄せる（例: 14px → 16、21px → 24、10.5px → 12、7px → 8）。Sizing 用寸法・意図的な負の margin・非余白の配置は対象外。
_Avoid_: 目視のみ（置換ごとに判断がぶれるため）, 切り上げ／切り捨て固定（文脈で最適が変わる中間値は四捨五入が既定）

**Border radius**:
角丸の半径。既存の `--radius` トークンで扱う。**Spacing** スケールとは別カテゴリ。v1 では Spacing 策定と同時に値や命名を変えない。
_Avoid_: Spacing（余白と混同しやすいため）, 角丸ルール（Border radius 用語と重複）

**Color**:
アプリ横断の色ルール。**Brand color**、**Neutral color**、**Semantic color** などのカテゴリ、**Color token**、**Element Plus color mapping** を含む。**Spacing**・**Border radius** とは別カテゴリ。
_Avoid_: パレット（具体的な色一覧だけを指す印象）, テーマ（タイポ・余白まで含む総称）

**Color token**:
Color の正本となる CSS カスタムプロパティ（`--color-*` 接頭辞）。`frontend/src/assets/style.css` の `:root` に定義する。**App color token** が唯一の色の出所であり、Element Plus の `--el-color-*` 等は **Element Plus color mapping** でここから委譲する（Spacing の **Spacing token** が正本で pattern が委譲するのと同型）。新規・改修では hex 直書きやレガシー名（`--accent` 等）を増やさない（**Color adoption**）。
_Avoid_: --el-color-primary（EP 変数を正本にしない）, accent（レガシー俗称）

**App color token**:
**Color token** のうち、アプリ UI の意味論（Brand / Neutral / Semantic 等）に沿って名前付けされたもの。画面・共有スタイルはこれを参照し、同じ意味の色を二重に hex 定義しない。
_Avoid_: デザイントークン（Sizing・Spacing と混同しやすいため）, ブランドカラー（英語以外の俗称だけ）

**Element Plus color mapping**:
**App color token** から Element Plus の `--el-*` 色変数へ値を渡す層。`html.dark` ブロック内で定義する。`el-button` 等の EP テーマ整合が目的。コンポーネントの独自スタイルは原則 **App color token** を参照し、`--el-*` 直参照は EP 上書きブロックと既存 EP 利用箇所に限定する。
_Avoid_: EP 正本（出所を Element Plus に置く案）, 二重カタログ（App と EP を別々に hex 管理する案）

**Color v1 scope**:
Color の最初の届け範囲。**ダークテーマのみ**の **App color token** 定義、**Element Plus color mapping**、**Color adoption**、**Color Storybook catalog**、ADR、`.cursor/rules/` に限定する。含めないもの: ライトテーマ用トークン、`prefers-color-scheme` 切替、テーマ切替 UI、全 View の hex 一括置換、Element Plus 内部色の全面再設計、Neutral border の多段トークン化、ESLint 強制。

**Color v1 deliverables**:
Issue／PR で最初に届ける具体物。(1) `frontend/src/assets/style.css` の全カテゴリ **App color token** と **Element Plus color mapping**、レガシー **Color alias**、(2) `frontend/src/design/colorTokens.ts`（Spacing の `spacingTokens.ts` 同型）、(3) **Server status color**・**Presence color** の既知 2 コンポーネントをトークン参照へ、(4) 共有スタイル（`.page-title`、`.section-card` 等）の Neutral 置換、(5) **Color Storybook catalog**、ADR。各 View の scoped hex は **Color adoption** に従い触ったところだけ。
_Avoid_: Color v1 scope（届け物リストを指すときは deliverables とセットで書く）, 全画面置換（adoption と矛盾するため）

**Color adoption**:
**App color token** への移行方針。新規画面・改修で触ったファイルでは **Color token** を使い、hex 直書きやレガシー名を増やさない。未着手の既存 View・scoped style は一括置換しない。**Color v1 deliverables** の共有スタイルと既知 Domain コンポーネントは v1 で寄せる。遵守は `.cursor/rules/` で案内し、移行期は ESLint では強制しない。見た目の正本は **Color Storybook catalog**。
_Avoid_: 全面置換, hex 禁止の CI 強制（即時一括・CI 強制を含意するため）

**Color alias**:
移行期に残す、旧 CSS 変数名から新 **App color token** へのエイリアス（例: `--accent` → `--color-brand`、`--bg-primary` → `--color-bg-base`）。v1 では **Color adoption** の破壊を避けるため `:root` に置く。削除時期は別判断（v1 では削除しない）。
_Avoid_: 永久互換の約束（削除時期を v1 で固定しないため）, 二重 hex（エイリアスは委譲のみ。別値を持たない）

**Color Storybook catalog**:
Color の Storybook 一覧。**Brand color**、**Neutral color token** 各段、**Semantic color catalog**、**Server status color**、**Presence color** のスウォッチとトークン名対応表、**Element Plus color mapping** の要点を載せ、見た目と使い方の正本とする（**Spacing Storybook catalog** と同型）。
_Avoid_: 全画面 Storybook, デザイントークン一覧（Sizing・Spacing の総称）

**Brand color**:
アプリのメインアクセント色（リンク、フォーカスリング、**Primary button** の塗り、サイドバーアクティブ項目など）。**Color** カテゴリのひとつ。**Primary button**（操作優先度の様式名）とは別概念。「Primary color」「アクセント色」としてパレット用語にしない。
_Avoid_: Primary color（**Primary button** と混同するため）, accent（レガシー `--accent` の俗称）, メインカラー（様式と色の区別が曖昧なため）

**Brand color token**:
**Brand color** を表す **App color token**（例: 本体・hover のペア）。**Element Plus color mapping** では `--el-color-primary` 系へ委譲する。レガシーの `--accent` / `--accent-hover` は v1 で **Brand color token** へ寄せ、移行期はエイリアス可。
_Avoid_: --color-primary（Primary button 用語と接頭辞が衝突するため）, --el-color-primary（正本ではなく mapping 先）

**Semantic color**:
UI の状態・結果を伝える **Color** カテゴリ。**Danger color**、**Success color**、**Warning color**、**Info color** の 4 種（**Semantic color catalog**）。`el-tag`・`ElMessage`・フォーム検証・**Danger button** の警告色など横断フィードバックに使う。**Semantic button**（Presence 色ボタン等のドメイン UI）とは別カテゴリ。
_Avoid_: Semantic button（ボタン分類の用語と混同するため）, 状態色（Server status のドメイン色と混同しやすいため）

**Semantic color catalog**:
v1 の **Semantic color** 公式一覧: danger（危険・エラー・破壊）、success（成功・正常）、warning（注意・要確認）、info（中立通知・補足）。Element Plus の `danger` / `success` / `warning` / `info` と 1 対 1 で **Element Plus color mapping** する。`error` は danger と同値でよい（EP 互換）。
_Avoid_: 3 色のみ（info を落とすと EP `type="info"` とずれる）, Server status 5 色（**Domain color** 側）

**Domain color**:
特定画面・ドメインだけが持つ意味付きの色。**Semantic color** とは別カタログ。全体のフィードバック色として流用しない。v1 では **Server status color** と **Presence color** を扱う。他ドメインは必要になったらカタログ追加。
_Avoid_: Semantic color（横断フィードバックと混同するため）, テーマカラー（**Brand color** と混同しやすいため）

**Server status color**:
**Server status presentation** 向けの **Domain color**。status.vrchat.com に近い 5 段（operational / degraded / partial / major / maintenance）＋ unknown。**Semantic color** の success / warning / danger とは別トークン（メンテナンス青・partial 橙などを無理に Semantic に畳まない）。
_Avoid_: Semantic color, プレゼンス色

**Presence color**:
**Presence change section** の **Semantic button**（Join Me / Active / Ask Me / Busy）向けの **Domain color**。VRChat クライアントに近い 4 色。**Semantic color catalog** や **Brand color** の流用はしない。
_Avoid_: Semantic color, クイックステータス色（旧称）

**Neutral color**:
背景・テキスト・枠線など、意味を持たない骨格色の **Color** カテゴリ。**Neutral surface**、**Neutral text**、**Neutral border** に分ける。**Brand color** や **Semantic color** の代替ではない。
_Avoid_: グレー（色相の指定を含意しないため）, テーマ色（**Brand color** と混同しやすいため）

**Neutral surface**:
画面の層（奥行き）を表す **Neutral color**。v1 は 3 段: **base**（ページ地）、**elevated**（カード・パネル）、**muted**（入力欄・ホバー・第 3 層）。**Neutral surface token** で表す。
_Avoid_: bg-primary（レガシー名）, 背景色（段数が分からないため）

**Neutral text**:
本文・ラベル・補足の **Neutral color**。v1 は 3 段: **primary**（本文）、**secondary**（ラベル・補足）、**muted**（プレースホルダ・無効表示）。**Neutral text token** で表す。`placeholder` / `disabled` は **Element Plus color mapping** で **muted** へ委譲してよい。
_Avoid_: text-secondary だけ（段数が足りないため）, プレースホルダ色（単独トークンにせず muted に含める）

**Neutral border**:
区切り線・カード枠の **Neutral color**。v1 は 1 段（既定の枠線色）。細い階層（light / lighter）は v1 では **Element Plus color mapping** 内の導出のみとし、Neutral カタログには載せない。
_Avoid_: ボーダー色（段数が分からないため）, EP border-light 列の全面トークン化（v1 スコープ外）

**Neutral color token**:
**Neutral surface** / **Neutral text** / **Neutral border** を表す **App color token**（`--color-bg-*`、`--color-text-*`、`--color-border`）。レガシーの `--bg-primary` / `--text-secondary` / `--border` は移行期エイリアス可。
_Avoid_: --el-bg-color（正本ではなく mapping 先）, 数値段階名のみ（base 等の役割名と対応が分からないため）

**Semantic color token**:
**Semantic color catalog** の各色 1 つずつの **App color token**（`--color-danger` 等）。hover や EP の `light-*` / `dark-*` 派生は **App color token** に含めない（**Element Plus color derivative** として mapping 内のみ）。
_Avoid_: --color-semantic-danger（category 接頭辞の重複）, light-3（EP 派生段階を App 正本にしない）

**Element Plus color derivative**:
**Element Plus color mapping** 内だけで定義する、EP コンポーネント向けの濃淡（`--el-color-primary-light-3`、`--el-color-danger-light-5` 等）。**App color token** の委譲先であり、画面・共有スタイルから直接参照しない。v1 では tokenize 時に **既存 hex を維持**し、見た目を変えない。
_Avoid_: Color token（App 正本と混同するため）, 派生色カタログ（v1 で App 側に持たない）

**Color v1 visual policy**:
Color v1 では **hex 値・見た目を変えない**。現行 `style.css`・**Server status color**・**Presence color** の色をそのまま **App color token** に移し、トークン名と参照の整理のみ行う。コントラスト改善・色相の揃え込み・Brand のリデザインは v1 スコープ外（別 PR／Issue）。
_Avoid_: リデザイン（v1 で見た目変更を含意するため）, トークン化 PR での微調整（回帰レビューと混ぜないため）

**Typography**:
アプリ横断の文字組ルール。UI 用 **Font family**、**Typography scale**（サイズ）、**Line height**、**Font weight**、用途別 **Text style**（見出し・本文・キャプション等）を含む。**Spacing** の Sizing・**Border radius** とは別カテゴリ。**Color** の **Neutral text**（色）とは別で、サイズ・行高・太さを扱う。
_Avoid_: フォント（ウェイト・行高を含意しない俗称）, テーマ（色・余白まで含む総称）

**Typography v1 scope**:
Typography の最初の届け範囲。**ダークテーマのみ**の **Typography token**・**Text style catalog**、**Typography adoption**、**Typography Storybook catalog**、ADR、`.cursor/rules/` に限定する。含める: アプリ作者が書くテキスト（`style.css` 共有クラス、新規・改修 View）、**Line height** の正本化（**Spacing** では除外済み）。含めない: Element Plus コンポーネント内部タイポの全面上書き、**Sizing**（アイコン・アバター寸法）、TitleBar 等ウィンドウクローム固有サイズ、**Code font**（モノスペース）カタログ、ライトテーマ・テーマ切替 UI、全 View 一括トークン化、ESLint 強制。
_Avoid_: Typography（v1 範囲を指すときは scope とセットで書く）, 将来拡張（スコープ外リストの総称として曖昧なため）

**Typography v1 deliverables**:
Issue／PR で最初に届ける具体物。(1) `frontend/src/assets/style.css` の **Typography token** 全種、**Text style catalog** 共有クラス、`body` / `.page-title` のトークン化、**Element Plus typography mapping**、(2) `frontend/src/design/typographyTokens.ts`（Spacing の `spacingTokens.ts` 同型）+ 単体テスト、(3) **Typography Storybook catalog**、(4) `.cursor/rules/typography-design-system.mdc`、(5) `docs/adr/0020-typography-design-system.md` + `CONTEXT.md`。各 View の scoped `rem` / 任意 px は **Typography adoption** に従い触ったところだけ。
_Avoid_: Typography v1 scope（届け物リストを指すときは deliverables とセットで書く）, 全画面置換（adoption と矛盾するため）

**Typography v1 visual policy**:
Typography v1 では **見た目を変えない**（**Color v1 visual policy** と同型）。現行の computed サイズ・行高・ウェイトをそのまま **Typography token** / **Text style** に移し、トークン名と参照の整理のみ行う。タイポグラフィのリデザイン・スケールの整理による寄せは v1 スコープ外（別 PR／Issue）。
_Avoid_: リデザイン（v1 で見た目変更を含意するため）, Spacing 型の四捨五入寄せを v1 deliverables に含める（回帰レビューと混ぜないため）

**Typography token**:
Typography の正本となる CSS カスタムプロパティ。`frontend/src/assets/style.css` の `:root` に定義する。サイズは **px リテラル**の **Font size token**（`--font-size-*`）、行高は **Line height token**、太さは **Font weight token**、用途別まとまりは **Text style**（セマンティック・エイリアスまたは共有クラス）。`rem` / `em` でスケールを表現しない。`html { font-size: 14px }` は Typography v1 では変更しない（**Spacing** ADR と同様）。`body` の既定サイズは **Font size token** の 14px 段へ委譲する。
_Avoid_: rem トークン（14px ルートと噛み合わないため）, font-md（数値と名前がずれるため）

**Font size token**:
**Typography scale** の各段階を表す **Typography token**。値は **px リテラル**（例: `--font-size-14: 14px`）。トークン名の数値は実 px と一致させる（`--font-size-14` は 14px）。新規・改修ではスケール外の任意 `px` / `rem` 直書きを増やさない（**Typography adoption**）。
_Avoid_: --text-14（サイズと色の接頭辞が衝突しうるため）, 1rem（単位混在）

**Typography scale**:
Typography で使ってよいフォントサイズの段階列。許容値は **10, 12, 14, 16, 18, 20, 24**（px）のみ。新規・改修ではこの列以外の任意サイズを増やさない（**Typography adoption**）。`body` 既定は **14** 段。
_Avoid_: 8 の倍数ルール（10 を含むため）, 連続スケール（任意の 1px 刻みすべてを許す印象）, Spacing scale との同一列（用途が別のため）

**Typography migration rounding**:
既存の `rem` や任意 px を **Font size token** へ置き換えるとき（**Typography adoption**）、スケールに無い値は **四捨五入で最寄りの段階**へ寄せる（例: 12.6px → 12 または 13 → **12**、15.4px → **16**、19.6px → **20**）。**Typography v1 deliverables** では寄せない。スケール外でも現行 computed と完全一致が必要な共有スタイルは **Text style** で `calc(var(--font-size-14) * 倍率)` 等を使い見た目を維持する。
_Avoid_: 目視のみ（置換ごとに判断がぶれるため）, v1 deliverables での四捨五入（**Typography v1 visual policy** と矛盾）

**Font family token**:
UI 本文スタックを表す **Typography token**（`--font-family-ui`）。モノスペース（**Code font**）は v1 では含めない。Web フォントの新規読み込みは v1 スコープ外。`button` / `input` / `textarea` は `inherit` で **Font family token** に追従する。
_Avoid_: --font-sans（Tailwind 俗称）, 日本語フォント（v1 でスタック変更を含意するため）

**UI font**:
**Font family token** が指す、本文・見出し・ラベル等に使うサンセリフスタック。v1 では現行 `body` と同型（Segoe UI → system UI → sans-serif）。**Typography v1 visual policy** に従いスタックも見た目も変えない。
_Avoid_: 本文フォント（**Code font** を含意しうるため）, Element Plus フォント（EP 内部は v1 で全面上書きしない）

**Line height token**:
行高を表す **Typography token**。値は **無単位の比率**（`px` ではない）。フォントサイズに追従する。**Line height scale** の各段階または **Line height pattern** 名で参照する。
_Avoid_: px 行高（テキストの拡大縮小に弱いため）, line-height-150（数値名と実値の対応が曖昧なため）

**Line height scale**:
Typography で使ってよい行高の段階列（無単位）。v1 は **1.25, 1.5, 1.75** の 3 段のみ。`body` 既定は **1.5**（`--line-height-normal`）。メニュー行高など固定 `px` の Sizing は含めない。
_Avoid_: 1.4（スケール外の中間値）, 連続スケール（任意の小数すべてを許す印象）

**Line height pattern**:
繰り返し現れる行高に付ける意味別の名前。**Line height token** のセマンティック・エイリアスとして定義し、実体は **Line height scale** へ委譲する。v1: `--line-height-tight` → 1.25、`--line-height-normal` → 1.5、`--line-height-relaxed` → 1.75。
_Avoid_: ユーティリティクラス（`.leading-normal` 等のクラス体系を含意するため）

**Font weight token**:
フォントウェイトを表す **Typography token**。トークン名の数値は CSS の `font-weight` 値と一致させる（例: `--font-weight-600: 600`）。**Font weight scale** 以外の任意ウェイトを新規・改修で増やさない（**Typography adoption**）。
_Avoid_: semibold（数値と名前がずれるため）, bold トークン（`700` と 1 対 1 でない印象）

**Font weight scale**:
Typography で使ってよいフォントウェイトの段階列。v1 は **400, 500, 600, 700** の 4 段のみ。`body` 既定は **400**。非標準値（`450` / `550` 等）は **Typography adoption** で最寄り段階へ四捨五入（`450` → `500`、`550` → `600`）。**Typography v1 deliverables** では寄せない。
_Avoid_: 300（現行 UI に無いため）, 800 以上（現行 UI に無いため）

**Text style**:
用途別の標準文字組（フォントサイズ・行高・ウェイトの組み合わせ）。**Text style catalog** で定義する。**Neutral text**（色）は **Color** の責務のため Text style には含めない。実装は `style.css` の共有クラス（`.text-h1` 等）。色が必要なときは呼び出し側または親で **Neutral text token** を指定する。
_Avoid_: タイポグラフィ（トークン全体の総称）, テキスト色（**Color** と混同しやすいため）

**Text style catalog**:
v1 の **Text style** 公式一覧（いずれも共有クラス + Storybook 参照）。

| Style | font-size | line-height | font-weight | クラス |
|-------|-----------|-------------|-------------|--------|
| Heading 1 | `calc(var(--font-size-14) * 1.4)` | tight | 600 | `.text-h1`（既存 `.page-title` は同型 alias） |
| Heading 2 | `--font-size-18` | tight | 600 | `.text-h2` |
| Heading 3 | `--font-size-16` | tight | 600 | `.text-h3` |
| Heading 4 | `--font-size-14` | normal | 600 | `.text-h4` |
| Body | `--font-size-14` | normal | 400 | `.text-body`（`body` 既定と同型） |
| Body small | `--font-size-12` | normal | 400 | `.text-body-sm` |
| Caption | `--font-size-10` | normal | 400 | `.text-caption` |

**Typography v1 visual policy** に従い Heading 1 は `calc` で現行 19.6px を維持する。将来リデザインで `--font-size-20` へ寄せてよい。
_Avoid_: h1 要素（HTML セマンティクスと 1 対 1 でない。見出しレベルは文脈で選ぶ）, 全スタイルの色込み定義

**Page title style**:
画面最上部のページ名向け **Text style**。**Heading 1** と同じ font-size / font-weight（`--font-size-h1` / 600）だが、line-height は **v1 visual policy** のため body の `--line-height-normal` を継承する（`.text-h1` の tight とは意図的に異なる）。v1 では既存クラス名 `.page-title` を残す。
_Avoid_: Heading 1（HTML の `<h1>` 限定を含意するため）, ウィンドウタイトル（TitleBar は **Typography v1 scope** 外）, `.text-h1` と完全同一（line-height が異なる）

**Element Plus typography mapping**:
**Font size token** から Element Plus の `--el-font-size-*` へ値を渡す層。`html.dark` 内で定義する。`el-form` / `el-table` 等の EP テーマ整合が目的。コンポーネントの独自スタイルは原則 **Typography token** または **Text style** を参照し、`--el-font-size-*` 直参照は EP 上書きブロックと既存 EP 利用箇所に限定する（**Element Plus color mapping** と同型）。
_Avoid_: EP 正本（出所を Element Plus に置く案）, 二重カタログ（App と EP を別々に管理する案）

**Element Plus typography derivative**:
**Element Plus typography mapping** 内だけで定義する、EP コンポーネント向けのスケール外サイズ（例: `--el-font-size-small: 13px`）。**Font size token** の委譲先ではなく、**Typography scale** に無い値を v1 では mapping 内に残して見た目を維持する。画面・共有スタイルから直接参照しない。
_Avoid_: Typography token（App 正本と混同するため）, 13px の App token 化（v1 でスケールを壊すため）

**Typography adoption**:
**Typography token** / **Text style** への移行方針。新規画面・改修で触ったファイルでは **Typography token** または **Text style catalog** を使い、`rem`・スケール外の任意 `px`・非標準 `font-weight` を増やさない。未着手の既存 View は一括置換しない（触ったところから順次）。**Typography v1 deliverables** の共有スタイルと **Element Plus typography mapping** は v1 で寄せる。置換時は **Typography migration rounding** を適用する（**Typography v1 deliverables** 自体では寄せない）。遵守は `.cursor/rules/` で案内し、移行期は ESLint では強制しない。見た目の正本は **Typography Storybook catalog**。
_Avoid_: 全面置換, rem 禁止の CI 強制（即時一括・CI 強制を含意するため）

**Typography Storybook catalog**:
Typography の Storybook 一覧。**Typography scale**・**Line height scale**・**Font weight scale**・**Text style catalog** の対応表と使用例、**Element Plus typography mapping** の要点を載せ、見た目と使い方の正本とする（**Spacing Storybook catalog** / **Color Storybook catalog** と同型）。
_Avoid_: 全画面 Storybook, デザイントークン一覧（Color・Spacing の総称）

**Form control**:
テキスト入力・選択・チェック・トグルなど、ユーザーが値を編集する単体 UI 部品。**VtButton** と同型で **VtInput** / **VtSelect** / **VtCheckbox** / **VtSwitch** ラッパーを正とし、`el-input` / `el-select` / `el-checkbox` / `el-switch` の直接利用は **Form control adoption** の移行期のみ併存可。見た目の色は既存の **Element Plus color mapping**（`--el-input-*` / `--el-switch-*`）を維持する。
_Avoid_: フォーム（**Form layout** 全体と混同しやすいため）, 入力欄（素の `<input>` や path 専用パターンまで含意しうるため）

**VtInput**:
テキスト系入力の標準ラッパー。内部は `el-input`（`type` 透過で password 等も可）。`textarea` は v1 では別ラッパーにせず、必要時は `el-input type="textarea"` を規約で扱うか後続で **VtTextarea** を検討する。
_Avoid_: el-input（移行完了後の直呼び）, path 入力行（**Path input row** は専用パターン）

**VtSelect**:
単一選択ドロップダウンの標準ラッパー。内部は `el-select` + `el-option`。フィルタ可能・複数選択・リモート検索は v1 スコープ外（必要になったら拡張）。
_Avoid_: el-select, セレクトボックス（HTML `<select>` と混同しやすいため）

**VtCheckbox**:
真偽を選ぶチェックボックスの標準ラッパー。内部は `el-checkbox`。**Form layout** 内で他フィールドと**同じ保存・反映タイミング**の真偽に使う（起動オプション、ルール条件、continue on error、既定フラグなど）。複数選択は `el-checkbox-group`（v1 ではグループラッパーなし）。即時 API 保存の単発設定や一覧の有効フラグには使わない（**VtSwitch**）。
_Avoid_: el-checkbox, トグル（**VtSwitch** と混同しやすいため）, 即時反映設定（**Immediate setting toggle** と混同しやすいため）

**VtSwitch**:
オン／オフを切り替えるトグルスイッチの標準ラッパー。内部は `el-switch`。**Immediate setting toggle**（**Setting row**）、**List enable toggle** に使う。`el-switch` の内蔵ラベル（`active-text` / `inline-prompt` 等）は v1 では使わない（ラベルは外側の **Form layout** または **Setting row** で付ける）。**Form layout** 内で他項目とまとめて保存する真偽には使わない（**VtCheckbox**）。
_Avoid_: el-switch, チェックボックス（**VtCheckbox** と混同しやすいため）, UI モード切替（**UI mode toggle** は別パターン）

**Immediate setting toggle**:
変更と同時に（またはほぼ同時に）永続化・API 反映する単発のオン／オフ設定。**VtSwitch** + 多くは **Setting row**。例: スリープ抑制、yt-dlp 実験機能の有効化。編集フォームの「保存ボタンまで送らない」真偽（ルール条件など）とは別。
_Avoid_: 有効フラグ（**List enable toggle** と混同しやすいため）, チェックボックス設定

**List enable toggle**:
一覧・カード行上で、その item の有効／無効を**即切替**する **VtSwitch**。編集フォームの保存タイミングとは独立。例: Automation item の `isEnabled`。`size="small"` を使ってよい。
_Avoid_: チェックボックス, フォーム内トグル

**UI mode toggle**:
2 状態の**表示モード**を切り替える UI（設定の永続化が主目的ではない）。**VtSwitch** / **VtCheckbox** の使い分けルールの対象外。**toggle-group** + ラジオ、またはラベル付きモード表示（Friends の Online/Offline 型）を **Semantic 例外**として維持。Form control ラッパーへの統合は v1 では行わない。
_Avoid_: VtSwitch, 設定スイッチ（**Immediate setting toggle** と混同しやすいため）, フィルタ（Gallery の絞り込みと混同しやすいため）

**Form control checkbox–switch rule**:
**VtCheckbox** と **VtSwitch** の使い分け正本。編集フォーム内の真偽（他フィールドと同じ保存・反映）→ **VtCheckbox**。即時反映の単発設定 → **VtSwitch** + **Setting row**。一覧・カードの有効フラグ即切替 → **VtSwitch**（small 可）。UI 表示モード切替 → **UI mode toggle**（Semantic 例外）。複数選択 → **VtCheckbox** + group。
_Avoid_: トグル全般, 真偽入力（4 コントロール以外の radio 等を含意しうるため）

**Form control size**:
Form control の寸法。**省略時は常に `default`**（暗黙の `small` / `large` は持たない）。`small` は **List enable toggle** の **VtSwitch**、Launcher Advanced など密集レイアウトで**明示指定したときのみ**。`large` は v1 不使用（ラッパーでも公開しない）。編集フォームは **`el-form size="default"`** で子へ継承し、個別 Vt* に毎回 `size` を付けない。
_Avoid_: small 既定, large ボタン（**VtButton** の size と混同しやすいため）

**Form control disabled state**:
操作できない Form control に付ける状態。ボタン様式ではなくいずれのコントロールにも重ねる。実装は標準 `disabled`（または同等）の**透過**。無効理由はラベル近くの hint・周辺文言で伝える（**Button disabled state** と同型）。`readonly` は v1 の Form control では使わず編集不可は `disabled` に統一。**VtInput** / **VtSelect** には **Button loading state** 相当の loading 様式は設けない（ロード中はフォーム全体 `disabled` または操作ボタン loading）。
_Avoid_: Readonly 様式, Loading 入力（ボタン loading と混同しやすいため）, Disabled 様式（第 5 の見た目名）

**Form field validation error**:
保存前にクライアントで判定できる不正（空必須・数値範囲など）の示し方。**Form layout** 内では **`el-form-item` のフィールド直下**（`:error` または将来 `:rules`）に **i18n 固定文**のみ示す。Vt* ラッパー内にエラー UI は持たない。バックエンドの生エラー文字列をラベルやフィールドエラーに載せない。
_Avoid_: ElMessage（一括保存失敗と混同しやすいため）, バリデーション（サーバー応答のマッピングまで含意しうるため）

**Form save failure feedback**:
編集フォームの一括保存 API が失敗したときの示し方。既存どおり **showToast.error**（i18n + `formatBackendError` 等）。フィールド別に原因が分からない失敗を `el-form-item` へ割り当てない。
_Avoid_: フィールドエラー, セクション alert（即時反映失敗と混同しやすいため）

**Immediate toggle failure feedback**:
**Immediate setting toggle** または **List enable toggle** の API 失敗の示し方。**そのブロック内**の **VtAlert** またはブロック直下のエラー行（Cookie linkage 型）。**i18n 分類済みの短い文**のみ。**showToast** だけに頼らない（一覧トグルと Setting row を同型にする）。失敗後はトグルを操作前の Effective 状態へ戻す（Cookie linkage draft 型）。
_Avoid_: Form save failure feedback, 生のバックエンドエラー, Server status fetch failure（読み取り失敗と混同しやすいため）

**Form fetch failure**:
フォーム表示用データの取得失敗（ログイン済みだが self 行を読めない等）。Form control のバリデーションとは別契約。カード内メッセージ＋コントロール `disabled` または非表示（Presence change / Dashboard launch block と同型）。`ElMessage` は出さない。
_Avoid_: Form field validation error, 空フォームでの編集

**Form control error v1 scope**:
Form control まわりのエラー表示 v1 範囲。上記 3 層（フィールド検証・一括保存・即時トグル）と **Form fetch failure** の文書化に限定。**含めない**: 全フォーム `:rules` 一括導入、`VtFormItem` ラッパー、サーバーのフィールド別エラー自動マッピング、Automation 一覧トグルの ElMessage 即時統一（触ったファイルから **Immediate toggle failure feedback** へ寄せる）。
_Avoid_: Form control error（v1 範囲を指すときは scope とセットで書く）

**Path input row**:
ファイルまたはフォルダのパスを入力する**複合レイアウトパターン**。**Form control**（単体）・**Form layout**・**Setting row** とは別の第三の型。`.path-input-group` で **VtInput** と参照 **VtButton**（**Secondary button**）を横並べ。Settings の path-row はラベル上・入力下で、参照に加え存在確認など**複数アクション**を並べてよい。v1 では **VtPathInput** ラッパーは作らない。`path-input-pattern.mdc` の素 `<input>` 例は **VtInput** へ更新する。**Cookies file source** のパス欄も同パターン。
_Avoid_: VtInput 単体（browse 必須の文脈）, Path input（行全体のレイアウトを含意しないため）, 解像度プリセット（**resolution-selection-ui** は別パターン）

**Path input row adoption**:
Path input row への揃え方針。新規・改修のパス欄は `.path-input-group` + **VtInput** + **VtButton**（参照は Secondary）。既存 `el-input` / `el-button` 直書きは触ったところから。**VtPathInput** 統合は v1 では行わない。
_Avoid_: 全面 Path ラッパー化, 素 input 新規追加（`element-plus-ui` と矛盾するため）

**Form control Storybook catalog**:
Form control の Storybook 正本。**VtInput** / **VtSelect** / **VtCheckbox** / **VtSwitch** それぞれ独立の `*.stories.ts`（通常・**Form control disabled state**・代表 props）。加えて **Form layout patterns story** 1 本（**Form layout**・**Setting row**・**Path input row** の並び例）。見た目と使い方は **VtButton Storybook catalog** と同型。
_Avoid_: 統合 catalog 1 本のみ（コンポーネント別の正本が薄れるため）, 全画面 Storybook

**Form control v1 deliverables**:
Issue／PR で最初に届ける具体物。(1) 4 ラッパー + 各単体テスト、(2) **Form control Storybook catalog**、(3) **Form layout patterns story**、(4) `.cursor/rules/form-design-system.mdc`、(5) `docs/adr/0021-form-control-design-system.md`、(6) `path-input-pattern.mdc` の Vt* 更新、(7) `CONTEXT.md` 同期。色・タイポは既存 **Element Plus color mapping** / **Element Plus typography mapping** を維持（見た目変更なし）。各 View の `el-*` 直書きは **Form control adoption** に従い触ったところだけ。
_Avoid_: Form control v1 scope（届け物リストを指すときは deliverables とセットで書く）, 全画面置換

**Form control adoption**:
Form control ラッパーへの移行方針。新規画面・改修で触ったファイルでは **VtInput** / **VtSelect** / **VtCheckbox** / **VtSwitch** を使う。未着手の既存 `el-*` 直書きは一括置換しない（触ったところから順次）。遵守は `.cursor/rules/` で案内し、移行期は ESLint では強制しない。見た目の正本は各 **Form control Storybook catalog**。
_Avoid_: 全面置換, el-input 禁止の CI 強制（即時一括・CI 強制を含意するため）

**Form control v1 scope**:
Form & Input Controls 策定の最初の届け範囲。4 ラッパー + Storybook + `.cursor/rules/` + ADR + `CONTEXT.md` に限定する。含める: 上記 4 種、`disabled` / `loading`（該当時）・`data-testid` の透過、**Element Plus color mapping** の維持。含めない: `el-input-number` / `el-radio` / `el-autocomplete` のラッパー、**Path input row** の統合、全 View 一括置換、ESLint 強制、バリデーション UI の全面再設計。
_Avoid_: Form control（v1 範囲を指すときは scope とセットで書く）, 将来拡張（スコープ外リストの総称として曖昧なため）

**Form layout**:
Form control を並べるレイアウト規約。**複数項目の編集**は **`el-form` + `label-position="top"`** を正とする（Launcher・Automation・Settings ログイン等）。項目間の縦間隔は **Spacing pattern** の `--space-form-field`（12px）。`el-form-item` 内の入力・選択は**原則幅 100%**（フル幅）。ラベルは `el-form-item` の `:label`（または同等）で付け、**Body small** 相当の **Neutral text** で示す。補足文（hint）は入力の直下に `el-text` 等で置き、ラベルと入力の間には入れない。
_Avoid_: setting-row（**Setting row** は単発設定行の別パターン）, 横ラベルフォーム（編集フォーム全体を指すとき）

**Setting row**:
`el-card` 内などで **1 行に説明とコントロールを横並べ**するレイアウトパターン。左にラベル＋任意の hint、右に **VtSwitch** や **VtInput** 等。Settings の電源スイッチ・保存期間入力が代表例。共有クラス `.setting-row` を正とし、v1 では **VtSettingRow** ラッパーは必須にしない。複数項目の編集ブロックは **Form layout**（top ラベル `el-form`）を使い、Setting row に寄せない。
_Avoid_: フォーム行（**Form layout** 全体と混同しやすいため）, 設定行（Launcher の profile 行など別 UI と混同しやすいため）

**Form layout adoption**:
Form layout / Setting row への揃え方針。新規・改修で複数項目編集は **Form layout**、カード内単発の横並び設定は **Setting row**。既存のばらつき（`setting-row` 内 `el-input-number` 等）は一括置換しない（触ったところから）。Spacing は **Spacing token** / **Spacing pattern** を使う。
_Avoid_: 全面レイアウト統一, el-form 強制の CI（即時一括を含意するため）

**Feedback & Badges**:
アプリ横断のフィードバック表示ルール。**Feedback channel**（Alert / Toast / Badge / Loading）の使い分けと配置の正本。**Form control** ADR のエラー 3 層（フィールド検証・一括保存・即時トグル・取得失敗）はフィールド文脈の契約として残し、**どのチャネルを使うか**はここで結線する（上書き・再定義はしない）。色は **Semantic color catalog**（[ADR 0019](docs/adr/0019-color-design-system.md)）を参照する。
_Avoid_: エラー表示（Form 専用の 3 層まで含む総称）, 通知（OS 通知や `ElNotification` を含意しうるため）

**Feedback channel**:
ユーザーへ状態・結果・注意を伝える表示経路の総称。v1 で扱う 4 種: **Section alert**（インライン `el-alert`）、**Toast**（`ElMessage`）、**Status badge**（`el-tag` 等のラベル）、**Loading feedback**（**Button loading state** 以外の待機表示を含む）。チャネル選択の正本は **Feedback & Badges** ADR（予定）。
_Avoid_: フィードバック（Form バリデーションだけを指す印象）, ElNotification（v1 では未採用の別 API）

**Section alert**:
**Feedback channel** のひとつ。画面上の特定**セクション・カード・ブロック内**に置くインラインアラート（v1 実装は **VtAlert**、内部 `el-alert`）。**条件が解消するまで**（または画面を離れるまで）残す状態・注意・ブロック内操作失敗を示す。`show-icon`・`:closable="false"` を既定とする（v1）。**Fetch failure**（Presence / Launch block / Gallery loadError / Video playback history 等）、**Immediate toggle failure feedback**、常時リスク警告（Cookie BAN 文、Launcher 未保存バナー）に使う。バックグラウンド refresh 失敗も、対象ブロックの状態としてここに示す（Toast にしない）。
_Avoid_: Toast, トースト, ElMessage（一時通知と混同しやすいため）, モーダル（`ElMessageBox` と混同しやすいため）

**VtAlert**:
**Section alert** の標準ラッパー。内部は `el-alert`。`variant`: `success` | `warning` | `danger` | `info`（**必須**。暗黙既定なし）。**Semantic color catalog** と 1 対 1。既定で `show-icon`・`:closable="false"`。`title`・`description`・`data-testid` 等は透過。**VtAlert adoption**（触ったところから `el-alert` 直書きを置換）。
_Avoid_: el-alert（移行完了後の直呼び）, Toast

**VtAlert adoption**:
VtAlert への移行方針。新規・改修で **Section alert** には **VtAlert** を使う。未着手の `el-alert` 直書きは一括置換しない。ESLint 強制は v1 では入れない。見た目の正本は **VtAlert Storybook catalog**。
_Avoid_: 全面置換, el-alert 禁止の CI 強制

**VtAlert Storybook catalog**:
VtAlert の Storybook 一覧。4 `variant` それぞれ（通常・長文 `description` あり）と、常時 warning / ブロック内 error の使用例を載せる（**VtTag Storybook catalog** と同型）。
_Avoid_: 全画面 Storybook, Section alert（配置ルールまで含む総称）

**Toast**:
**Feedback channel** のひとつ。**showToast** 経由の画面端一時通知（内部は `ElMessage`）。**ユーザーが明示操作した直後**の結果フィードバック（保存・起動・反映・削除の成否、件数付き成功メッセージ等）。数秒で自動消滅し、消えても画面の主要 UI は読める。**Fetch failure** やブロック読み取り失敗には使わない（**Section alert** またはカード内文言）。`success` / `error` / `warning` / `info` は **Semantic color catalog** と 1 対 1。
_Avoid_: Section alert, 通知バナー（常時表示を含意するため）, ElNotification（v1 未採用 API）, ElMessage 直叩き（新規・改修の正本は **showToast**）

**showToast**:
**Toast** の薄いユーティリティ（`frontend/src/utils/showToast.ts` 等）。`showToast.success(message)` / `.error` / `.warning` / `.info` で `ElMessage` に委譲。引数は **i18n 済み string のみ**（生 `Error`・バックエンド生文字列は渡さない）。例外は呼び出し側で `formatError` / `formatBackendError` してから渡す。表示時間・`grouping` 等の既定はここに集約。**showToast adoption**（触ったところから `ElMessage.*` 直叩きを置換）。DOM ラッパー（VtToast）は v1 では作らない。
_Avoid_: VtToast, errorFromUnknown（例外吸収の魔法 API。v1 では呼び出し側で format して明示）, ElMessage（実装 API 名だけの直呼び正本）

**showToast adoption**:
showToast への移行方針。新規・改修で **Toast** には **showToast** を使う。未着手の `ElMessage.*` 直叩きは一括置換しない。ESLint 強制は v1 では入れない。
_Avoid_: 全面置換, ElMessage 禁止の CI 強制

**Feedback channel selection**:
Alert と Toast の切り分け正本。**主軸は持続性**（解消まで残す → Section alert、操作直後 → Toast）。**補足は文脈の近さ**（特定ブロックの状態 → Section alert、画面横断の操作結果 → Toast）。**Form save failure feedback** は Toast（**showToast.error**）。**Form fetch failure** は Section alert 型（Toast 禁止）。**Immediate toggle failure feedback** は Section alert 型。
_Avoid_: 重要度だけで機械分岐（success は常に Toast 等。常時 warning は Section alert のため）

**Status badge**:
**Feedback channel** のひとつ。短いラベルで状態・結果・属性を示すチップ UI（主に `el-tag`）。v1 では **Semantic status badge**・**Neutral badge**・**Domain badge** の 3 分類。**VtTag** が Semantic / Neutral の標準ラッパー。**Domain badge** は専用コンポーネントのまま **VtTag adoption** 対象外。
_Avoid_: User tag chip（**Domain badge** の VrcUserTagChip と混同しやすいため）, バッジ（通知バッジの OS/UI 俗称）

**Semantic status badge**:
操作結果・有効状態・ログイン済みなど、**良し悪し・注意度**を **Semantic color catalog**（success / warning / danger / info）で示す **Status badge**。Automation 実行ログの成否、Settings のログイン済み表示など。**VtTag** の `variant`（または EP `type`）で 4 色と 1 対 1。**Domain badge** や **Neutral badge** に Semantic 色を流用しない。
_Avoid_: primary ラベル（Brand 寄り属性と混同しやすいため）, プレゼンス色（**Domain badge**）

**Neutral badge**:
意味論色を持たない **Status badge**。件数・カテゴリ・技術メタのラベル（Licenses 件数、Automation の kind 表示など）。**VtTag** の `variant="neutral"`（`type` 省略・EP 既定の中立灰）。**Launch profile default label** は例外（下記）。
_Avoid_: Semantic status badge, info（補足通知色と混同しやすいため）

**Launch profile default label**:
Launcher サイドバー一覧で **Default launch profile** を示す **Status badge**。**Semantic status badge** ではない。**Neutral badge** でもなく、**Brand color** 寄りの `primary`（**VtTag** `variant="primary"` または EP `type="primary"`）を**この 1 用途だけ**許可する。成功・ログイン済み等の Semantic 表示と混同しない。
_Avoid_: 既定タグ（Neutral への統一を含意するため）, success バッジ（状態の良し悪しと混同しやすいため）

**Domain badge**:
特定ドメインだけが持つ意味付きチップ。**Semantic color catalog** を横断フィードバックとして流用しない（**VtButton** の **Semantic button** と同型の例外）。v1: **VrcStatusTag**（プレゼンス文字列）、**VrcUserTagChip**（User tag）。専用 Storybook・色ルールは各ドメイン／**Domain color** ADR 側。**VtTag** には寄せない。
_Avoid_: Semantic status badge, Status badge（3 分類全体と混同しやすいため）

**VtTag**:
**Semantic status badge**・**Neutral badge**・**Launch profile default label** の標準ラッパー。内部は `el-tag`。`variant`: `success` | `warning` | `danger` | `info` | `neutral` | `primary`（**必須**。暗黙既定なし）。`primary` は **Launch profile default label** のみ（**Brand color mapping**）。`size`・`data-testid` 等は透過。**VtTag adoption**（触ったところから `el-tag` 直書きを置換）。**Domain badge** は対象外。
_Avoid_: el-tag（移行完了後の直呼び）, VrcStatusTag（**Domain badge**）

**VtTag adoption**:
VtTag への移行方針。新規・改修で Semantic / Neutral の **Status badge** には **VtTag** を使う。**Domain badge** は専用 UI のまま。未着手の `el-tag` 直書きは一括置換しない。ESLint 強制は v1 では入れない。見た目の正本は **VtTag Storybook catalog**。
_Avoid_: 全面置換, Domain badge の VtTag 化

**Loading feedback**:
**Feedback channel** のひとつ。非同期処理の待機を示す表示。v1 は **Action loading** と **Section loading** の 2 層のみ。**Button loading state**（[ADR 0017](docs/adr/0017-button-design-system-vtbutton.md)）を操作待ちの正本とする。`v-loading` オーバーレイ・スケルトン・画面全体ロック・進捗％は v1 スコープ外。
_Avoid_: Button loading state（第 2 の様式名。操作ボタン待ちは ADR 0017 を正本とする）, スピナー（実装部品の総称）

**Action loading**:
ユーザーが押した**操作ボタン**に付ける待機表示。**Button loading state** と同義。VtButton / `el-button` の `:loading`。Action group 内の他ボタンまで一括無効化は原則しない。取得中は関連入力を `disabled` してよい（Gallery の `loading || scanning` と同型）。
_Avoid_: Section loading, 画面ロック

**Section loading**:
ブロック・カード・一覧の**初回または再取得中**に、その領域内でコンテンツの代わりに示す待機表示。中立テキスト 1 行（`common.loading` またはブロック専用 i18n、**Neutral text**）。Presence change・Activity・Video 履歴が代表例。loading 中は **Toast** を出さない。**Section alert**（loadError）と排他（loading → error → content）。一覧が空でも loading 中は empty メッセージを出さない。
_Avoid_: v-loading, スケルトン, Toast, Action loading（押下ボタンだけの待ち）

**Feedback & Badges v1 scope**:
Feedback & Badges 策定の最初の届け範囲。**VtAlert**・**VtTag**・**showToast**、各 Storybook catalog（Alert / Tag）、単体テスト、`.cursor/rules/feedback-design-system.mdc`、ADR 0022、`CONTEXT.md`、**Feedback channel selection** 参照表（Form ADR 3 層との結線）に限定する。含めないもの: `ElNotification`・OS 通知、`v-loading`・スケルトン・進捗％、VtToast、全 View 一括置換、ESLint 強制、Dashboard refresh 警告や Automation トグル失敗の v1 強制移行、**Domain badge** の変更。
_Avoid_: Feedback & Badges（v1 範囲を指すときは scope とセットで書く）, 将来拡張（スコープ外リストの総称として曖昧なため）

**Feedback & Badges v1 deliverables**:
Issue／PR で最初に届ける具体物。(1) **VtAlert** + テスト + **VtAlert Storybook catalog**、(2) **VtTag**（semantic 4 + neutral + primary）+ テスト + **VtTag Storybook catalog**、(3) **showToast** + テスト、(4) `.cursor/rules/feedback-design-system.mdc`、(5) `docs/adr/0022-feedback-badges-design-system.md`、(6) `CONTEXT.md` 同期。移行は **VtAlert adoption** / **VtTag adoption** / **showToast adoption** のみ（触ったところから）。
_Avoid_: Feedback & Badges v1 scope（届け物リストを指すときは deliverables とセットで書く）, 全画面置換（adoption と矛盾するため）

## Agent contribution

Issue・PR・コミットなど Git に残るテキストを書くときの用語。

### Language

**Public contribution artifact**:
Git 履歴や GitHub 上に公開される成果物。Pull request・Issue・コミットメッセージ・ブランチ名、および Agent がそれら向けに生成する下書きや `docs/ai_dlc/` の Issue メモを含む。VRChat 上の実在ユーザーを特定できる情報を載せない対象。
_Avoid_: 公開物, Git テキスト（スコープが曖昧なため）

**Redacted reproduction**:
Public contribution artifact に書くバグ再現・検証記述。VRChat 表示名・`usr_*` ID・プロフィール URL・ログイン username・インスタンス文字列内の user ID を使わず、件数・ステータス・手順の抽象（例: offline フレンドがキャッシュに無い）で述べる。詳細ルールは `docs/agents/redaction.md`。
_Avoid_: 匿名化, 個人情報マスク（置換手順まで含意しないため）
