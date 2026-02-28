---
applyTo: "**/*.xal"
---

# xaligo DSL (.xal) 仕様

## 概要

`.xal` ファイルは XML 構文を持つ Vue 風レイアウト DSL です。
ルートタグは必ず `<frame>` でなければなりません。
パーサーは `encoding/xml` を使用し、属性・入れ子タグ・テキストコンテンツを扱います。

## ルートタグ

```xml
<frame width="1440" height="900" class="pa-4">
  ...
</frame>
```

| 属性 | 型 | デフォルト | 説明 |
|---|---|---|---|
| `width` | float | `1280` | フレーム幅 (px) |
| `height` | float | `720` | フレーム高さ (px) |
| `class` | string | — | スペーシングクラス |
| `layout` | string | — | `"horizontal"` を指定すると子要素を横並びに配置する |
| `gap` | float | `16` | 子要素間の隙間 (px) |
| `item-size` | float | `48`（設定値）| このファイル内の全 `<item>` に適用するアイコン最大サイズ (px)。`app.yaml` の `item.icon_size` よりも優先される |

## レイアウトタグ

### `<container>`

子要素を **縦方向** に均等配置します (`frame` と同挙動)。`layout="horizontal"` を指定すると横並びになります。

```xml
<container class="pa-4" gap="16">
  ...
</container>
```

| 属性 | 型 | デフォルト | 説明 |
|---|---|---|---|
| `layout` | string | — | `"horizontal"` で子要素を横並び配置にする |
| `gap` | float | `16` | 子要素間の隙間 (px) |

### `<row>`

子要素を **横方向** に 12 カラムグリッドで配置します。

```xml
<row gap="20">
  <col span="8">...</col>
  <col span="4">...</col>
</row>
```

| 属性 | 型 | デフォルト | 説明 |
|---|---|---|---|
| `gap` | float | `16` | 列間の隙間 (px) |
| `border` | string | — | `"none"` を指定すると枠線を非表示にする |

### `<col>`

`<row>` 内に置く縦方向スタックコンテナ。`span` で占有カラム数を指定します。

| 属性 | 型 | デフォルト | 説明 |
|---|---|---|---|
| `span` | float | `12 / 列数` | 占有カラム数 (合計 12 基準) |
| `class` | string | — | スペーシングクラス |

## リーフタグ

`frame` / `container` / `row` / `col` / AWS グループタグ / `item` 以外のタグはすべてリーフ要素として扱われます。
Excalidraw 上では `rectangle + text` のペアとして描画されます。

```xml
<card title="Dashboard" />
<panel title="Main Chart" />
<text>任意のラベル</text>
```

| 属性 | 動作 |
|---|---|
| `title` | 表示ラベル (優先) |
| テキストコンテンツ | title がない場合のラベル |
| (なし) | タグ名をラベルとして使用 |
| `border` | `"none"` を指定すると枠線を非表示にする |
| `visible` | `"false"` を指定するとそのコンポーネント自身 (枠・アイコン・ラベル) のみを非表示にする。子要素は親の `visible` に関わらず個別に描画される。レイアウト上の占有スペースは維持される |

## `<item>` タグ

AWS サービスアイコンをコンテナ内に配置するリーフ要素です。
`service-catalog.csv` に記載された数値 ID を `id` 属性で指定します。
アイコンは指定サイズ (`item-size` の値) に収まるように描画されます。

```xml
<public-subnet title="Public Subnet">
  <item id="1178" />   <!-- アイコンあり -->
  <item />             <!-- スペーサー: アイコンなし、レイアウト上のスロットのみ確保 -->
  <item id="1189" />   <!-- アイコンあり -->
</public-subnet>
```

| 属性 | 型 | 必須 | 説明 |
|---|---|---|---|
| `id` | int | — | `service-catalog.csv` のサービス ID。省略または空の場合はスペーサーとして扱われる |

> `id` に対応するアイコンが見つからない場合は描画をスキップし、エラーにはなりません。

## `<connection>` タグ

`<item>` 要素間に **折れ線矢印 (elbowed arrow)** を引きます。
`<frame>` の直接子要素として、レイアウトタグの **外側** に記述します。
`src` / `dst` には `<item id="N">` と同じカタログ ID を指定します。

```xml
<frame width="1122" height="794" class="pa-4">
  <aws-cloud title="AWS Cloud">
    <public-subnet title="Public Subnet">
      <item id="1178" />
      <item id="1189" />
    </public-subnet>
  </aws-cloud>

  <!-- connection は frame の直接子として最後に記述する -->
  <connection src="1178" dst="1189" />
</frame>
```

| 属性 | 型 | 必須 | 説明 |
|---|---|---|---|
| `src` | int | ✓ | 矢印の始点アイコンのカタログ ID |
| `dst` | int | ✓ | 矢印の終点アイコンのカタログ ID |

**矢印仕様:**
- `elbowed: true` — 常に直角折れ線 (Excalidraw の "elbow connector")
- 終点のみ矢印 (`endArrowhead: "arrow"`, `startArrowhead: null`)
- 線色 `#1e1e1e`, 線幅 `2px`
- 始点・終点は要素の **辺中央** に接続される
  - 接続方向が **下向き** の場合 → ラベルテキスト要素 (`{id}-item-lbl`) の下辺
  - その他の方向 → アイコン画像要素 (`{id}-item`) の対応する辺
- `fixedPoint` で辺を正規化座標で固定するため Excalidraw がファイルを開いた際に矢印が正しくスナップされる
- 矢印 ID は `conn-{src}-{dst}-{index}` 形式
- バインドされた要素の `boundElements` に矢印 ID が登録される

**接続辺の選択ロジック:**

| 方向 (dst が src に対して) | 始点側の辺 | 終点側の辺 |
|---|---|---|
| 右 (dx ≥ dy) | right | left |
| 左 | left | right |
| 下 (dy > dx) | bottom (ラベル) | top |
| 上 | top | bottom (ラベル) |

> `src` / `dst` に対応するアイテムがレンダリングされていない場合は警告を出力してスキップします。

## AWS グループタグ

`container` と同様に子要素を縦スタック配置しつつ、**AWS アーキテクチャ図の Group ボーダースタイル**で描画されるタグ群です。
テンプレートは `etc/resources/aws/templates/excalidraw/` (`.excalidraw`) と `etc/resources/aws/templates/xal/` (`.xal`) に拡張子別で配置されています。
アイコン SVG は `etc/resources/aws/svg/Architecture-Group-Icons/` から参照します。

```xml
<aws-cloud title="Production Environment">
  <vpc title="vpc-0a1b2c3d">
    <private-subnet title="Private Subnet A">
      <card title="App Server" />
    </private-subnet>
  </vpc>
</aws-cloud>
```

| タグ | 表示名 | ボーダー色 | スタイル | アイコン |
|---|---|---|---|---|
| `<aws-cloud>` | AWS Cloud | `#000000` | solid | AWS-Cloud-logo_32.svg |
| `<aws-cloud-alt>` | AWS Cloud | `#000000` | solid | AWS-Cloud_32.svg |
| `<region>` | Region | `#00A1C9` | dashed | Region_32.svg |
| `<availability-zone>` | Availability Zone | `#00A1C9` | dashed | — |
| `<security-group>` | Security group | `#CC0000` | dashed | — |
| `<auto-scaling-group>` | Auto Scaling group | `#E7601B` | dashed | Auto-Scaling-group_32.svg |
| `<vpc>` | Virtual private cloud (VPC) | `#8C4FFF` | solid | Virtual-private-cloud-VPC_32.svg |
| `<private-subnet>` | Private subnet | `#00A1C9` | solid | Private-subnet_32.svg |
| `<public-subnet>` | Public subnet | `#3F8624` | solid | Public-subnet_32.svg |
| `<server-contents>` | Server contents | `#7A7C7F` | solid | Server-contents_32.svg |
| `<corporate-data-center>` | Corporate data center | `#7A7C7F` | solid | Corporate-data-center_32.svg |
| `<ec2-instance-contents>` | EC2 instance contents | `#E7601B` | solid | EC2-instance-contents_32.svg |
| `<spot-fleet>` | Spot Fleet | `#E7601B` | solid | Spot-Fleet_32.svg |
| `<aws-account>` | AWS account | `#E7008A` | solid | AWS-Account_32.svg |
| `<aws-iot-greengrass-deployment>` | AWS IoT Greengrass Deployment | `#3F8624` | solid | AWS-IoT-Greengrass-Deployment_32.svg |
| `<aws-iot-greengrass>` | AWS IoT Greengrass | `#3F8624` | solid | — |
| `<elastic-beanstalk-container>` | Elastic Beanstalk container | `#E7601B` | solid | — |
| `<aws-step-functions-workflow>` | AWS Step Functions workflow | `#E7008A` | solid | — |
| `<generic-group>` | Generic group | `#AAB7B8` | dashed | — |

属性は `container` と同じ (`title`, `class`, `gap` など)。

### レイアウト制御属性 (コンテナ共通)

`frame` / `container` / `col` および全 AWS グループタグに指定できます。

| 属性 | 指定値 | 説明 |
|---|---|---|
| `layout` | `"horizontal"` | 子要素を **横方向** に比率配置 (`col` 属性で幅比率を指定) |
| `layout` | `"staggered"` | 子要素を奥行きオフセット付きで重ねて配置 (AWS グループタグのみ有効) |
| `gap` | float | 子要素間隔 (px). デフォルト `16` |

### 子要素のサイズ比率属性

| 属性 | 使用方向 | 説明 |
|---|---|---|
| `row` | 縦 (`layoutStack`) | 子要素の **高さ比率** (flex-grow 相当). デフォルト `1.0` (均等) |
| `col` | 横 (`layout="horizontal"`) | 子要素の **幅比率** (flex-grow 相当). デフォルト `1.0` (均等) |

```xml
<!-- 横並び: 左2:右1 の幅比率 -->
<vpc title="VPC" layout="horizontal">
  <public-subnet title="Public" col="2" />
  <private-subnet title="Private" col="1" />
</vpc>

<!-- 縦並び: 上1:下2 の高さ比率 -->
<region title="Region">
  <vpc title="VPC A" row="1" />
  <vpc title="VPC B" row="2" />
</region>
```

## スペーシングクラス (`class` 属性)

Vuetify ライクな記法。**単位は `spacingUnit = 8px`**。

### 一括指定

| クラス | 意味 |
|---|---|
| `pa-{n}` | padding 全方向 = n × 8px |
| `ma-{n}` | margin 全方向 = n × 8px |

### 軸一括指定

| クラス | 意味 |
|---|---|
| `px-{n}` | padding 左右 = n × 8px |
| `py-{n}` | padding 上下 = n × 8px |
| `mx-{n}` | margin 左右 = n × 8px |
| `my-{n}` | margin 上下 = n × 8px |

### 個別方向指定

| クラス | 意味 |
|---|---|
| `pt-{n}` | padding-top |
| `pr-{n}` | padding-right |
| `pb-{n}` | padding-bottom |
| `pl-{n}` | padding-left |
| `mt-{n}` | margin-top |
| `mr-{n}` | margin-right |
| `mb-{n}` | margin-bottom |
| `ml-{n}` | margin-left |

複数クラスはスペース区切り: `class="pa-4 mt-2"`

### セマンティクス

| 種別 | 対象タグ | 動作 |
|---|---|---|
| `padding` | frame / container / col | box 内側の余白。子要素の配置開始点が pad 分だけ内側になる |
| `padding` | AWS グループタグ / 未知コンテナ | `defaultGroupTopInset(44)` / `defaultGroupSideInset(12)` に**加算**される。`pa-2` を指定するとヘッダー下に +16px の追加余白が生まれる |
| `margin` | 任意の子要素 | 親レイアウト (`layoutStack` / `layoutRow`) が事前に読み取り、sibling 間スペースとして割り当てる (CSS flex の margin に相当) |

## レイアウト計算ルール

1. `frame` / `container` / `col` → **縦スタック** (高さを `gap` を差し引いて均等分割)
2. `row` → **12 カラムグリッド** (`span` 属性で各列の幅を決定)
3. リーフ要素 → 親から受け取った `(x, y, w, h)` をそのまま使用
4. `margin` は要素自身の位置・サイズに影響し、`padding` は子要素の配置開始点に影響する

## サンプル

```xml
<frame width="1440" height="900" class="pa-4">
  <container class="pa-4">
    <row gap="20" class="mb-2">
      <col span="8" class="pa-2">
        <card title="Dashboard" />
      </col>
      <col span="4" class="pa-2">
        <card title="Summary" />
      </col>
    </row>

    <row gap="20">
      <col span="4" class="pa-2">
        <panel title="Filters" />
      </col>
      <col span="8" class="pa-2">
        <panel title="Main Chart" />
      </col>
    </row>
  </container>
</frame>
```

## 制約・注意事項

- ルートタグは `<frame>` 固定。それ以外はパースエラー。
- 自己クローズタグ (`<card title="..." />`) と通常タグ (`<card title="..."></card>`) の両方を使用可能。
- `row` 直下の子要素の `span` の合計は 12 以内を推奨 (超過した場合は右方向にはみ出す)。
- `.xal` ファイルは UTF-8 で保存すること。
