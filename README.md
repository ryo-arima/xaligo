# xaligo

Vue 風 DSL (`.xal`) を **Excalidraw JSON** へ変換する Go 製 CLI ツール。  
Vuetify スタイルのスペーシング・グリッドレイアウトと AWS アーキテクチャ図グループタグを内蔵しています。

## インストール

```bash
git clone https://github.com/ryo-arima/xaligo
cd xaligo
go mod tidy
make build        # .bin/xaligo を生成
```

## コマンド一覧

| コマンド | 説明 |
|---|---|
| `xaligo generate excalidraw --xal <file.xal> -o <out.excalidraw>` | .xal → .excalidraw 変換 |
| `xaligo generate xal [flags] -o <out.xal>` | AWS インフラ階層の .xal を自動生成 |
| `xaligo add service --name <name> --file <file>` | AWS サービスアイコンを単体追加 |
| `xaligo add service --list <csv> --file <file>` | AWS サービスアイコンを一括追加 |
| `xaligo version` | バージョン表示 |

### generate xal フラグ

```
--clouds N                         AWS Cloud ブロック数 (default 1)
--accounts N                       Account ブロック数 (default 1)
--regions N                        Region ブロック数 (default 1)
--azs N                            Availability Zone 数 (default 2)
--az-layout grid|staggered         AZ 配置スタイル (default grid)
--subnets N                        サブネット数 (default 2)
--spacing vertical|horizontal|both スペーシング方向 (default both)
--start top|left                   描画開始位置 (default top)
--paper A4                         用紙サイズ
--orientation portrait|landscape   用紙向き (default landscape)
-o <file>                          出力ファイルパス
```

## クイックスタート

```bash
# AWS 構成の .xal を自動生成
.bin/xaligo generate xal --clouds 1 --accounts 1 --regions 2 --azs 2 \
  --az-layout staggered -o output/infra.xal

# .xal を .excalidraw へ変換
.bin/xaligo generate excalidraw --xal output/infra.xal -o output/infra.excalidraw
```

Excalidraw で `output/infra.excalidraw` をインポートしてください。

## .xal DSL

### ルート構造

```xml
<frame width="1122" height="794" class="pa-4" item-size="48">
  <!-- ここに要素を並べる -->
</frame>
```

### レイアウトタグ

| タグ | 説明 |
|---|---|
| `<frame>` | ルートタグ。幅・高さ・余白を指定 |
| `<container>` | 縦スタックコンテナ (`layout="horizontal"` で横並び) |
| `<row>` | 12 カラムグリッド行 |
| `<col>` | `<row>` 内の列 (`span` で幅指定) |

### AWS グループタグ

AWS アーキテクチャ図のグループボーダースタイルで描画されるタグ群。

| タグ | 表示名 | ボーダー色 |
|---|---|---|
| `<aws-cloud>` | AWS Cloud | `#000000` |
| `<aws-account>` | AWS Account | `#E7008A` |
| `<region>` | Region | `#00A1C9` |
| `<availability-zone>` | Availability Zone | `#00A1C9` |
| `<vpc>` | VPC | `#8C4FFF` |
| `<public-subnet>` | Public Subnet | `#3F8624` |
| `<private-subnet>` | Private Subnet | `#00A1C9` |
| `<security-group>` | Security Group | `#CC0000` |
| `<auto-scaling-group>` | Auto Scaling Group | `#E7601B` |
| `<server-contents>` | Server Contents | `#7A7C7F` |
| `<corporate-data-center>` | Corporate Data Center | `#7A7C7F` |
| その他 | 詳細は xal-spec 参照 | — |

### `<item>` タグ

`service-catalog.csv` の ID を指定して AWS サービスアイコンを埋め込みます。
`id` を省略または空にすると、アイコンは描画されないスペーサーとして機能します。

```xml
<public-subnet title="Public Subnet">
  <item id="1178" />   <!-- アイコンあり -->
  <item />             <!-- スペーサー (空白スロット) -->
  <item id="1189" />   <!-- アイコンあり -->
</public-subnet>
```

### `<connection>` タグ

`<item>` 要素間に折れ線矢印を引きます。`<frame>` の直接子として記述します。

```xml
<frame width="1122" height="794" class="pa-4">
  <!-- ... レイアウト要素 ... -->

  <!-- frame の最後に接続を列挙 -->
  <connection src="1178" dst="1189" />
</frame>
```

| 属性 | 説明 |
|---|---|
| `src` | 矢印始点のカタログ ID |
| `dst` | 矢印終点のカタログ ID |

矢印は常に **elbowed (直角折れ線)** スタイルで描画されます。
始点・終点はアイコン画像またはラベルテキストの **辺中央** に接続されます。
接続方向が下向きの場合はラベルテキスト要素の辺、それ以外はアイコン画像の辺を使用します。
Excalidraw の `fixedPoint` バインディングで辺を固定するため、ファイルを開くと矢印が要素に正しくスナップされます。

### 主要属性

| 属性 | 対象 | 説明 |
|---|---|---|
| `title` | 任意 | 表示ラベル |
| `layout="horizontal"` | コンテナ系 | 子要素を横並びに配置 |
| `layout="staggered"` | AWS グループタグ | 奥行きオフセット付きで重ねて配置 |
| `row="N"` | 縦スタック内の子要素 | 高さ比率 (flex-grow 相当) |
| `col="N"` | `layout="horizontal"` 内の子要素 | 幅比率 (flex-grow 相当) |
| `gap="N"` | コンテナ系 | 子要素間隔 (px) |
| `border="none"` | 任意 | 枠線を非表示 |
| `visible="false"` | 任意 | 自コンポーネントのみ非表示 (子要素は描画される) |
| `item-size="N"` | `<frame>` | このファイル内の全 `<item>` アイコンサイズを上書き (px) |
| `class` | 任意 | Vuetify 風スペーシングクラス |

### スペーシングクラス

単位は `8px`。複数クラスはスペース区切り: `class="pa-4 ml-2"`

| パターン | 説明 |
|---|---|
| `pa-{n}` / `ma-{n}` | padding / margin 全方向 |
| `px-{n}` / `py-{n}` | padding 左右 / 上下 |
| `mx-{n}` / `my-{n}` | margin 左右 / 上下 |
| `pt/pr/pb/pl-{n}` | padding 各方向 |
| `mt/mr/mb/ml-{n}` | margin 各方向 |

## サンプル DSL

```xml
<frame width="1122" height="794" class="pa-4">
  <aws-cloud title="AWS Cloud 1">
    <aws-account title="Account 1">

      <!-- Region 1: 全体の 80% (row="4") -->
      <region title="Region 1" row="4">
        <vpc title="VPC 1" class="ml-2" layout="staggered">
          <availability-zone title="AZ 1">
            <public-subnet title="Public Subnet 1" row="2">
              <item id="1178" /><item id="1189" />
            </public-subnet>
            <!-- visible="false": 枠は非表示だがレイアウト占有は維持 -->
            <private-subnet title="Private Subnet 1" visible="false">
              <item id="1179" /><item id="1183" />
            </private-subnet>
          </availability-zone>
          <availability-zone title="AZ 2">
            <public-subnet title="Public Subnet 1" />
            <private-subnet title="Private Subnet 1" />
          </availability-zone>
        </vpc>
      </region>

      <!-- Region 2: 全体の 20% (row="1") — 横並びサブネット -->
      <region title="Region 2" row="1">
        <vpc title="VPC 2" class="ml-2" layout="horizontal" gap="12">
          <public-subnet title="Public Subnet 1" col="2" />
          <private-subnet title="Private Subnet 1" col="1" />
          <private-subnet title="Private Subnet 2" col="1" />
        </vpc>
      </region>

    </aws-account>
  </aws-cloud>
</frame>
```

## 設定ファイル

`etc/resources/aws/app.yaml` でパスやデフォルト値を変更できます（省略時はすべてデフォルト値）。

```yaml
paths:
  asset_package:       etc/resources/aws/svg
  service_catalog_csv: etc/resources/aws/service-catalog.csv
  output_frames:       output/aws-frames

legend:
  offset_x:  120
  offset_y:  0
  icon_size: 32
  font_size: 12

item:
  icon_size: 48   # <item> アイコンのデフォルトサイズ (px)
```

## ビルド・テスト

```bash
make build   # .bin/xaligo をビルド
make run     # examples/sample.xal → output/sample.excalidraw
make clean   # .bin/ と output/ を削除
go test ./...
```

## ライセンス

[MIT](LICENSE)
