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

## レイアウトタグ

### `<container>`

子要素を **縦方向** に均等配置します (`frame` と同挙動)。

```xml
<container class="pa-4" gap="16">
  ...
</container>
```

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

### `<col>`

`<row>` 内に置く縦方向スタックコンテナ。`span` で占有カラム数を指定します。

| 属性 | 型 | デフォルト | 説明 |
|---|---|---|---|
| `span` | float | `12 / 列数` | 占有カラム数 (合計 12 基準) |
| `class` | string | — | スペーシングクラス |

## リーフタグ

`frame` / `container` / `row` / `col` 以外のタグはすべてリーフ要素として扱われます。
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

## スペーシングクラス (`class` 属性)

Vuetify ライクな記法。**単位は `spacingUnit = 8px`**。

### 一括指定

| クラス | 意味 |
|---|---|
| `pa-{n}` | padding 全方向 = n × 8px |
| `ma-{n}` | margin 全方向 = n × 8px |

### 方向指定

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
