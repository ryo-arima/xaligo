# xaligo

Vue-like の独自DSLでレイアウトを記述し、Excalidrawに取り込める `.excalidraw` JSON を生成する Go CLI です。

## 目的

- Vuetify風の概念（`container`, `row`, `col`）でフレーム配置を記述
- `pa-*`, `ma-*`, `pt-*` などの余白ユーティリティで整列を簡潔に指定
- Excalidrawへそのまま読み込めるJSONを出力

## インストール

```bash
go mod tidy
go build -o .bin/xaligo ./cmd
```

## 使い方

```bash
# サンプルDSLを生成
.bin/xaligo init -o examples

# DSLをExcalidraw JSONへ変換
.bin/xaligo render examples/sample.xal -o examples/sample.excalidraw
```

Excalidrawで `examples/sample.excalidraw` をインポートしてください。

## DSL スケルトン

```xml
<frame width="1440" height="900" class="pa-4">
  <container class="pa-4">
    <row gap="20">
      <col span="8" class="pa-2">
        <card title="Dashboard" />
      </col>
      <col span="4" class="pa-2">
        <card title="Summary" />
      </col>
    </row>
  </container>
</frame>
```

## 現在のスコープ（スケルトン）

- CLI: `init`, `render`, `version`
- パーサ: XML互換のVueライクタグ構文をAST化
- レイアウト: `frame/container/row/col` の最低限配置
- 余白: `pa-*`, `ma-*`, `pt/pr/pb/pl`, `mt/mr/mb/ml`
- 出力: Excalidraw elements（rectangle + text）

## 今後の拡張候補

- ネストされたgridの精密レイアウト
- align/justify などのFlex風プロパティ
- コンポーネント定義と再利用
- テーマ/トークン化
- バリデーションとフォーマッタ
