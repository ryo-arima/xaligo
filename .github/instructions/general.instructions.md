---
applyTo: "**"
---

# xaligo — 一般コーディング規約

## プロジェクト概要

`xaligo` は Vue 風の独自 DSL (.xal ファイル) を、Vuetify スタイルのグリッド / パディング / マージン / コンテナを使って Excalidraw の JSON ファイルへ変換する Go 製 CLI ツールです。
また AWS サービスアイコンを既存 `.excalidraw` ファイルへ追記する `add service` コマンドも備えます。

## モジュール情報

```
module: github.com/ryo-arima/xaligo
Go:     1.22
主要依存: github.com/spf13/cobra v1.8.1
         gopkg.in/yaml.v3 v3.0.1
```

## ディレクトリ構成

```
xaligo/
├── cmd/
│   └── main.go                   # エントリポイント
├── pkg/
│   ├── command.go                # Root cobra コマンド (サブコマンドを wiring)
│   └── controller/
│       ├── render.go             # xaligo render <input.xal> -o output.excalidraw
│       ├── init.go               # xaligo init -o <dir>  (sample.xal を生成)
│       ├── version.go            # xaligo version
│       └── add.go                # xaligo add service [flags]
├── internal/
│   ├── model/
│   │   └── ast.go                # DSL AST: Document, Node
│   ├── parser/
│   │   └── parser.go             # XML ベース DSL パーサー
│   ├── layout/
│   │   └── layout.go             # Vuetify スタイル レイアウトエンジン
│   ├── excalidraw/
│   │   └── scene.go              # Excalidraw JSON ビルダー (render コマンド用)
│   ├── entity/
│   │   ├── scene.go              # Scene 構造体 (add コマンド用)
│   │   └── service.go            # ServiceEntry 構造体
│   ├── repository/
│   │   ├── builder.go            # MakeText / MakeImage 要素ビルダー
│   │   ├── scene.go              # ReadScene / WriteScene
│   │   ├── icon.go               # SvgToDataURL / FileID / LoadFromCSV / SVGBGColor
│   │   └── service_list.go       # ReadServiceList (CSV/TXT パーサー)
│   └── config/
│       └── config.go             # Config 構造体 + findProjectRoot + etc/resources/aws/app.yaml 読み込み
├── examples/
│   └── sample.xal               # サンプル DSL ファイル
├── etc/
│   └── resources/
│       └── aws/
│           ├── app.yaml         # パス・凡例サイズ等の設定 (デフォルト値あり)
│           ├── service-catalog.csv  # 全 SVG アイコンカタログ
│           ├── svg/             # AWS アイコン SVG (Architecture-Service/Group/Resource/Category-Icons)
│           └── templates/
│               ├── excalidraw/  # AWS Group タグ別テンプレート (.excalidraw)
│               └── xal/         # AWS Group タグ別テンプレート (.xal)
├── scripts/
│   ├── gen_service_catalog.py   # service-catalog.csv 再生成
│   └── gen_group_templates.py   # etc/resources/aws/templates/{excalidraw,xal}/ 再生成
├── Makefile
├── go.mod / go.sum
└── README.md
```

## アーキテクチャ方針

- **cmd → pkg/command.go → pkg/controller/ → internal/**: 依存方向を一方向に保つ。
- `internal/` パッケージは `pkg/` からのみ参照し、`cmd/` から直接参照しない。
- 各 `controller` ファイルは `Init<Cmd>Cmd() *cobra.Command` というファクトリ関数を公開し、`pkg/command.go` でルートコマンドへ登録する。
- ビジネスロジックは `internal/` に閉じ込め、cobra のフラグ処理は `controller` 層で行う。

## コーディング規約

- Go 標準の `gofmt` / `golint` に従う。
- パッケージ名は小文字英単語 (例: `controller`, `repository`, `entity`)。
- エラーは `fmt.Errorf("<context>: %w", err)` でラップして呼び出し元へ返す。
- `panic` は使用しない。エラーは必ず `error` として返す。
- Excalidraw の要素は `map[string]interface{}` で表現する (既存フォーマットと互換を保つため)。

## 設定ファイル (etc/resources/aws/app.yaml)

プロジェクトルート (`go.mod` が存在するディレクトリ) の `etc/resources/aws/app.yaml` を読み込む。
不在時はすべてデフォルト値が使われるため、ファイルは必須ではない。

```yaml
paths:
  asset_package:        etc/resources/aws/svg
  service_catalog_csv:  etc/resources/aws/service-catalog.csv
  output_frames:        output/aws-frames
legend:
  offset_x:  120
  offset_y:  0
  icon_size: 32
  font_size: 12
```

## CLI コマンド一覧

| コマンド | 説明 |
|---|---|
| `xaligo render <file.xal> -o <out.excalidraw>` | DSL を Excalidraw JSON へ変換 |
| `xaligo init [-o <dir>]` | `sample.xal` を生成 |
| `xaligo version` | バージョン表示 |
| `xaligo add service --name <name> --file <file>` | AWS サービスアイコンを単体追加 |
| `xaligo add service --list <csv> --file <file>` | AWS サービスアイコンを一括追加 |
| `xaligo generate xal --clouds N --accounts N --regions N --azs N --az-layout grid\|staggered --subnets N --spacing vertical\|horizontal\|both --start top\|left --paper A4 --orientation portrait\|landscape -o out.xal` | AWS インフラ階層の .xal を生成 |
| `xaligo generate excalidraw --xal <file.xal> -o <out.excalidraw>` | .xal を .excalidraw へ変換 |

## テスト・ビルド

```bash
make build   # .bin/xaligo をビルド
make run     # examples/sample.xal → output/sample.excalidraw
make clean   # .bin/ と output/ を削除
go test ./...            # 全テスト実行
go build ./...           # ビルドエラー確認
```
