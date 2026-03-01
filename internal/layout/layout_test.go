package layout

import (
	"strings"
	"testing"

	"github.com/ryo-arima/xaligo/internal/model"
	"github.com/ryo-arima/xaligo/internal/parser"
)

func TestAttrFloat(t *testing.T) {
	tests := []struct {
		v        string
		fallback float64
		want     float64
	}{
		{"100", 0, 100},
		{"", 42, 42},
		{"  ", 42, 42},
		{"abc", 10, 10},
		{"3.14", 0, 3.14},
	}
	for _, tc := range tests {
		got := attrFloat(tc.v, tc.fallback)
		if got != tc.want {
			t.Errorf("attrFloat(%q, %v) = %v, want %v", tc.v, tc.fallback, got, tc.want)
		}
	}
}

func TestParseClassSpacing(t *testing.T) {
	tests := []struct {
		class   string
		wantPad Spacing
		wantMar Spacing
	}{
		{
			"pa-2",
			Spacing{16, 16, 16, 16},
			Spacing{},
		},
		{
			"ma-1",
			Spacing{},
			Spacing{8, 8, 8, 8},
		},
		{
			"px-3 py-1",
			Spacing{Top: 8, Bottom: 8, Left: 24, Right: 24},
			Spacing{},
		},
		{
			"pt-2 pb-1 pl-3 pr-4",
			Spacing{Top: 16, Bottom: 8, Left: 24, Right: 32},
			Spacing{},
		},
		{
			"mt-1 mb-2 ml-3 mr-4",
			Spacing{},
			Spacing{Top: 8, Bottom: 16, Left: 24, Right: 32},
		},
		{
			"mx-2 my-3",
			Spacing{},
			Spacing{Top: 24, Bottom: 24, Left: 16, Right: 16},
		},
		{
			"",
			Spacing{},
			Spacing{},
		},
	}
	for _, tc := range tests {
		pad, mar := parseClassSpacing(tc.class)
		if pad != tc.wantPad {
			t.Errorf("parseClassSpacing(%q) pad = %+v, want %+v", tc.class, pad, tc.wantPad)
		}
		if mar != tc.wantMar {
			t.Errorf("parseClassSpacing(%q) mar = %+v, want %+v", tc.class, mar, tc.wantMar)
		}
	}
}

func mustParse(t *testing.T, src string) model.Document {
	t.Helper()
	doc, err := parser.Parse(strings.NewReader(src))
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	return doc
}

func TestBuildFrameDimensions(t *testing.T) {
	doc := mustParse(t, `<frame width="1440" height="900"></frame>`)
	box, err := Build(doc)
	if err != nil {
		t.Fatalf("Build error: %v", err)
	}
	if box.W != 1440 || box.H != 900 {
		t.Errorf("frame dimensions = (%v, %v), want (1440, 900)", box.W, box.H)
	}
}

func TestBuildFrameDefaultDimensions(t *testing.T) {
	doc := mustParse(t, `<frame></frame>`)
	box, err := Build(doc)
	if err != nil {
		t.Fatalf("Build error: %v", err)
	}
	if box.W != 1280 || box.H != 720 {
		t.Errorf("default dimensions = (%v, %v), want (1280, 720)", box.W, box.H)
	}
}

func TestBuildNilRoot(t *testing.T) {
	_, err := Build(model.Document{Root: nil})
	if err == nil {
		t.Fatal("expected error for nil root")
	}
}

func TestBuildStackChildren(t *testing.T) {
	src := `<frame width="1000" height="600" class="pa-4">
		<container>
			<card title="A" />
			<card title="B" />
		</container>
	</frame>`
	doc := mustParse(t, src)
	box, err := Build(doc)
	if err != nil {
		t.Fatalf("Build error: %v", err)
	}
	if len(box.Children) != 1 {
		t.Fatalf("expected 1 child, got %d", len(box.Children))
	}
	container := box.Children[0]
	if len(container.Children) != 2 {
		t.Fatalf("expected 2 children in container, got %d", len(container.Children))
	}
	a := container.Children[0]
	b := container.Children[1]
	if a.Y >= b.Y {
		t.Errorf("stack layout: A.Y=%v should be less than B.Y=%v", a.Y, b.Y)
	}
}

func TestBuildRowLayout(t *testing.T) {
	src := `<frame width="1000" height="600">
		<row gap="0">
			<col span="6"><card title="Left" /></col>
			<col span="6"><card title="Right" /></col>
		</row>
	</frame>`
	doc := mustParse(t, src)
	box, err := Build(doc)
	if err != nil {
		t.Fatalf("Build error: %v", err)
	}
	row := box.Children[0]
	if len(row.Children) != 2 {
		t.Fatalf("expected 2 cols, got %d", len(row.Children))
	}
	left := row.Children[0]
	right := row.Children[1]
	if left.X >= right.X {
		t.Errorf("row layout: left.X=%v should be less than right.X=%v", left.X, right.X)
	}
}

func TestBuildHorizontalLayout(t *testing.T) {
	src := `<frame width="1000" height="600" layout="horizontal">
		<card title="A" />
		<card title="B" />
	</frame>`
	doc := mustParse(t, src)
	box, err := Build(doc)
	if err != nil {
		t.Fatalf("Build error: %v", err)
	}
	if len(box.Children) != 2 {
		t.Fatalf("expected 2 children, got %d", len(box.Children))
	}
	a := box.Children[0]
	b := box.Children[1]
	if a.X >= b.X {
		t.Errorf("horizontal layout: A.X=%v should be less than B.X=%v", a.X, b.X)
	}
}

func TestBuildPaddingReducesInner(t *testing.T) {
	src := `<frame width="1000" height="600" class="pa-4">
		<card title="Inner" />
	</frame>`
	doc := mustParse(t, src)
	box, err := Build(doc)
	if err != nil {
		t.Fatalf("Build error: %v", err)
	}
	pad := float64(4 * spacingUnit)
	child := box.Children[0]
	// child's X and Y should be >= frame's X/Y + padding
	if child.X < pad {
		t.Errorf("child.X=%v should be >= padding=%v", child.X, pad)
	}
	if child.Y < pad {
		t.Errorf("child.Y=%v should be >= padding=%v", child.Y, pad)
	}
}

func TestBuildConnectionFilteredFromLayout(t *testing.T) {
	src := `<frame width="1000" height="600">
		<card title="A" />
		<connection src="1" dst="2" />
	</frame>`
	doc := mustParse(t, src)
	box, err := Build(doc)
	if err != nil {
		t.Fatalf("Build error: %v", err)
	}
	// connection should not appear as layout child
	for _, ch := range box.Children {
		if ch.Tag == "connection" {
			t.Error("connection tag should not appear in layout children")
		}
	}
}

func TestBuildRowColProportion(t *testing.T) {
	src := `<frame width="1200" height="600">
		<row gap="0">
			<col span="4"><card title="A" /></col>
			<col span="8"><card title="B" /></col>
		</row>
	</frame>`
	doc := mustParse(t, src)
	box, err := Build(doc)
	if err != nil {
		t.Fatalf("Build error: %v", err)
	}
	row := box.Children[0]
	colA := row.Children[0]
	colB := row.Children[1]
	// colB should be approximately twice as wide as colA
	ratio := colB.W / colA.W
	if ratio < 1.9 || ratio > 2.1 {
		t.Errorf("col width ratio = %v, want ~2.0", ratio)
	}
}
