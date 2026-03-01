# xaligo

A Go CLI tool that converts a Vue-style DSL (`.xal`) into **Excalidraw JSON**.  
It bundles Vuetify-style spacing/grid layout and AWS architecture diagram group tags.

## Installation

```bash
git clone https://github.com/ryo-arima/xaligo
cd xaligo
go mod tidy
make build        # produces .bin/xaligo
```

## Commands

| Command | Description |
|---|---|
| `xaligo generate excalidraw --xal <file.xal> -o <out.excalidraw>` | Convert .xal → .excalidraw |
| `xaligo generate xal [flags] -o <out.xal>` | Auto-generate a .xal for an AWS infrastructure hierarchy |
| `xaligo add service --name <name> --file <file>` | Add a single AWS service icon |
| `xaligo add service --list <csv> --file <file>` | Add AWS service icons in bulk |
| `xaligo version` | Print version |

### generate xal flags

```
--clouds N                         Number of AWS Cloud blocks (default 1)
--accounts N                       Number of Account blocks (default 1)
--regions N                        Number of Region blocks (default 1)
--azs N                            Number of Availability Zones (default 2)
--az-layout grid|staggered         AZ placement style (default grid)
--subnets N                        Number of subnets (default 2)
--spacing vertical|horizontal|both Spacing direction (default both)
--start top|left                   Drawing start position (default top)
--paper A4                         Paper size
--orientation portrait|landscape   Paper orientation (default landscape)
-o <file>                          Output file path
```

## Quick Start

```bash
# Auto-generate a .xal for an AWS setup
.bin/xaligo generate xal --clouds 1 --accounts 1 --regions 2 --azs 2 \
  --az-layout staggered -o output/infra.xal

# Convert the .xal to .excalidraw
.bin/xaligo generate excalidraw --xal output/infra.xal -o output/infra.excalidraw
```

Import `output/infra.excalidraw` into Excalidraw.

## .xal DSL

### Root structure

```xml
<frame width="1122" height="794" class="pa-4" item-size="48">
  <!-- place elements here -->
</frame>
```

### Layout tags

| Tag | Description |
|---|---|
| `<frame>` | Root tag. Sets width, height, and padding |
| `<container>` | Vertical stack container (`layout="horizontal"` for horizontal) |
| `<row>` | 12-column grid row |
| `<col>` | Column inside `<row>` (width set via `span`) |

### AWS group tags

Tags that render with the group border style used in AWS architecture diagrams.

| Tag | Display name | Border color |
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
| others | See xal-spec for full list | — |

### `<item>` tag

Embeds an AWS service icon by specifying an ID from `service-catalog.csv`.
Omitting or leaving `id` empty makes the element act as a spacer (no icon rendered).

```xml
<public-subnet title="Public Subnet">
  <item id="1178" />   <!-- icon rendered -->
  <item />             <!-- spacer (empty slot) -->
  <item id="1189" />   <!-- icon rendered -->
</public-subnet>
```

### `<connection>` tag

Draws an elbowed arrow between `<item>` elements. Write it as a direct child of `<frame>`.

```xml
<frame width="1122" height="794" class="pa-4">
  <!-- ... layout elements ... -->

  <!-- list connections at the end of frame -->
  <connection src="1178" dst="1189" />
</frame>
```

| Attribute | Description |
|---|---|
| `src` | Catalog ID of the arrow's start element |
| `dst` | Catalog ID of the arrow's end element |

Arrows are always rendered in **elbowed (right-angle)** style.
The start and end points connect to the **midpoint of the edge** of the icon image or label text.
When the connection direction is downward the label text element's edge is used; otherwise the icon image edge is used.
`fixedPoint` binding in Excalidraw pins the edge so arrows snap correctly when the file is opened.

### Key attributes

| Attribute | Target | Description |
|---|---|---|
| `title` | any | Display label |
| `layout="horizontal"` | container-like | Lay out children horizontally |
| `layout="staggered"` | AWS group tags | Stack children with depth offset |
| `row="N"` | child in vertical stack | Height ratio (flex-grow equivalent) |
| `col="N"` | child inside `layout="horizontal"` | Width ratio (flex-grow equivalent) |
| `gap="N"` | container-like | Gap between children (px) |
| `border="none"` | any | Hide border |
| `visible="false"` | any | Hide only this component (children are still rendered) |
| `item-size="N"` | `<frame>` | Override icon size for all `<item>` elements in this file (px) |
| `class` | any | Vuetify-style spacing classes |

### Spacing classes

Unit is `8px`. Multiple classes are space-separated: `class="pa-4 ml-2"`

| Pattern | Description |
|---|---|
| `pa-{n}` / `ma-{n}` | padding / margin on all sides |
| `px-{n}` / `py-{n}` | padding left+right / top+bottom |
| `mx-{n}` / `my-{n}` | margin left+right / top+bottom |
| `pt/pr/pb/pl-{n}` | padding per side |
| `mt/mr/mb/ml-{n}` | margin per side |

## Sample DSL

```xml
<frame width="1122" height="794" class="pa-4">
  <aws-cloud title="AWS Cloud 1">
    <aws-account title="Account 1">

      <!-- Region 1: 80% of total height (row="4") -->
      <region title="Region 1" row="4">
        <vpc title="VPC 1" class="ml-2" layout="staggered">
          <availability-zone title="AZ 1">
            <public-subnet title="Public Subnet 1" row="2">
              <item id="1178" /><item id="1189" />
            </public-subnet>
            <!-- visible="false": border hidden but layout space is preserved -->
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

      <!-- Region 2: 20% of total height (row="1") — horizontal subnets -->
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

## Configuration

You can override paths and default values in `etc/resources/aws/app.yaml` (all values fall back to defaults when omitted).

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
  icon_size: 48   # default icon size for <item> elements (px)
```

## Build & Test

```bash
make build   # build .bin/xaligo
make run     # examples/sample.xal → output/sample.excalidraw
make clean   # remove .bin/ and output/
go test ./...
```

## License

[MIT](LICENSE)
