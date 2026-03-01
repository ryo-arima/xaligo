package layout

import (
	"strings"
	"testing"

	"github.com/ryo-arima/xaligo/internal/model"
	"github.com/ryo-arima/xaligo/internal/parser"
)

// ─── parseClassSpacing ────────────────────────────────────────────────────────

func TestParseClassSpacing_PaAll(t *testing.T) {
	pad, mar := parseClassSpacing("pa-2")
	want := float64(2 * spacingUnit)
	if pad.Top != want || pad.Right != want || pad.Bottom != want || pad.Left != want {
		t.Errorf("pa-2: pad = %+v, want all %v", pad, want)
	}
	if mar.Top != 0 || mar.Right != 0 || mar.Bottom != 0 || mar.Left != 0 {
		t.Errorf("pa-2: unexpected margin %+v", mar)
	}
}

func TestParseClassSpacing_MaAll(t *testing.T) {
	pad, mar := parseClassSpacing("ma-3")
	want := float64(3 * spacingUnit)
	if mar.Top != want || mar.Right != want || mar.Bottom != want || mar.Left != want {
		t.Errorf("ma-3: mar = %+v, want all %v", mar, want)
	}
	if pad.Top != 0 {
		t.Errorf("ma-3: unexpected padding")
	}
}

func TestParseClassSpacing_AxesAndDirections(t *testing.T) {
	cases := []struct {
		class    string
		padTop   float64
		padRight float64
		padBot   float64
		padLeft  float64
		marTop   float64
		marBot   float64
	}{
		{"px-1 py-2", 16, 8, 16, 8, 0, 0},
		{"pt-1 pr-2 pb-3 pl-4", 8, 16, 24, 32, 0, 0},
		{"mt-1 mb-2", 0, 0, 0, 0, 8, 16},
	}
	for _, tc := range cases {
		pad, mar := parseClassSpacing(tc.class)
		if pad.Top != tc.padTop {
			t.Errorf("[%s] pad.Top = %v, want %v", tc.class, pad.Top, tc.padTop)
		}
		if pad.Right != tc.padRight {
			t.Errorf("[%s] pad.Right = %v, want %v", tc.class, pad.Right, tc.padRight)
		}
		if pad.Bottom != tc.padBot {
			t.Errorf("[%s] pad.Bottom = %v, want %v", tc.class, pad.Bottom, tc.padBot)
		}
		if pad.Left != tc.padLeft {
			t.Errorf("[%s] pad.Left = %v, want %v", tc.class, pad.Left, tc.padLeft)
		}
		if mar.Top != tc.marTop {
			t.Errorf("[%s] mar.Top = %v, want %v", tc.class, mar.Top, tc.marTop)
		}
		if mar.Bottom != tc.marBot {
			t.Errorf("[%s] mar.Bottom = %v, want %v", tc.class, mar.Bottom, tc.marBot)
		}
	}
}

func TestParseClassSpacing_Empty(t *testing.T) {
	pad, mar := parseClassSpacing("")
	if pad != (Spacing{}) || mar != (Spacing{}) {
		t.Errorf("empty class: expected zero spacing, got pad=%+v mar=%+v", pad, mar)
	}
}

// ─── Build helpers ────────────────────────────────────────────────────────────

func mustParse(t *testing.T, src string) model.Document {
	t.Helper()
	doc, err := parser.Parse(strings.NewReader(src))
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	return doc
}

// ─── Build ────────────────────────────────────────────────────────────────────

func TestBuild_FrameDimensions(t *testing.T) {
	doc := mustParse(t, `<frame width="800" height="600"></frame>`)
	box, err := Build(doc)
	if err != nil {
		t.Fatalf("Build error: %v", err)
	}
	if box.W != 800 || box.H != 600 {
		t.Errorf("frame dimensions = %vx%v, want 800x600", box.W, box.H)
	}
	if box.X != 0 || box.Y != 0 {
		t.Errorf("frame origin = (%v,%v), want (0,0)", box.X, box.Y)
	}
}

func TestBuild_DefaultFrameDimensions(t *testing.T) {
	doc := mustParse(t, `<frame></frame>`)
	box, err := Build(doc)
	if err != nil {
		t.Fatalf("Build error: %v", err)
	}
	if box.W != 1280 || box.H != 720 {
		t.Errorf("default dimensions = %vx%v, want 1280x720", box.W, box.H)
	}
}

func TestBuild_NilRootError(t *testing.T) {
	_, err := Build(model.Document{Root: nil})
	if err == nil {
		t.Fatal("expected error for nil root")
	}
}

func TestBuild_SingleLeafChild(t *testing.T) {
	src := `<frame width="400" height="200"><card title="A" /></frame>`
	doc := mustParse(t, src)
	box, err := Build(doc)
	if err != nil {
		t.Fatalf("Build error: %v", err)
	}
	if len(box.Children) != 1 {
		t.Fatalf("expected 1 child, got %d", len(box.Children))
	}
	child := box.Children[0]
	if child.Tag != "card" {
		t.Errorf("child tag = %q, want card", child.Tag)
	}
	// leaf child should span the full inner width
	if child.W <= 0 {
		t.Errorf("child W should be positive, got %v", child.W)
	}
}

func TestBuild_VerticalStack(t *testing.T) {
	src := `<frame width="400" height="200">
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
	a, b := box.Children[0], box.Children[1]
	if a.Y >= b.Y {
		t.Errorf("expected A above B: A.Y=%v B.Y=%v", a.Y, b.Y)
	}
}

func TestBuild_HorizontalLayout(t *testing.T) {
	src := `<frame width="400" height="200" layout="horizontal">
		<card title="Left" />
		<card title="Right" />
	</frame>`
	doc := mustParse(t, src)
	box, err := Build(doc)
	if err != nil {
		t.Fatalf("Build error: %v", err)
	}
	if len(box.Children) != 2 {
		t.Fatalf("expected 2 children, got %d", len(box.Children))
	}
	left, right := box.Children[0], box.Children[1]
	if left.X >= right.X {
		t.Errorf("expected Left before Right: Left.X=%v Right.X=%v", left.X, right.X)
	}
}

func TestBuild_RowLayout(t *testing.T) {
	src := `<frame width="400" height="200">
		<row>
			<col span="8"><card title="Main" /></col>
			<col span="4"><card title="Side" /></col>
		</row>
	</frame>`
	doc := mustParse(t, src)
	box, err := Build(doc)
	if err != nil {
		t.Fatalf("Build error: %v", err)
	}
	if len(box.Children) != 1 {
		t.Fatalf("expected 1 row child, got %d", len(box.Children))
	}
	row := box.Children[0]
	if len(row.Children) != 2 {
		t.Fatalf("expected 2 col children, got %d", len(row.Children))
	}
	main, side := row.Children[0], row.Children[1]
	if main.X >= side.X {
		t.Errorf("main col should be left of side col: main.X=%v side.X=%v", main.X, side.X)
	}
	// span 8 col should be wider than span 4 col
	if main.W <= side.W {
		t.Errorf("span-8 col should be wider than span-4: main.W=%v side.W=%v", main.W, side.W)
	}
}

func TestBuild_PaddingShrinksInnerArea(t *testing.T) {
	srcNoPad := `<frame width="400" height="200"><card title="A" /></frame>`
	srcPad := `<frame width="400" height="200" class="pa-4"><card title="A" /></frame>`

	docNoPad := mustParse(t, srcNoPad)
	docPad := mustParse(t, srcPad)

	boxNoPad, _ := Build(docNoPad)
	boxPad, _ := Build(docPad)

	wNoPad := boxNoPad.Children[0].W
	wPad := boxPad.Children[0].W
	if wPad >= wNoPad {
		t.Errorf("padding should shrink child width: no-pad W=%v, pad W=%v", wNoPad, wPad)
	}
}

func TestBuild_ConnectionNotIncludedInChildren(t *testing.T) {
	src := `<frame width="400" height="200">
		<card title="A" />
		<connection src="1" dst="2" />
	</frame>`
	doc := mustParse(t, src)
	box, err := Build(doc)
	if err != nil {
		t.Fatalf("Build error: %v", err)
	}
	// connection must not appear as a layout box child
	for _, ch := range box.Children {
		if ch.Tag == "connection" {
			t.Error("connection should be excluded from layout children")
		}
	}
}

func TestBuild_StaggeredLayout(t *testing.T) {
	src := `<frame width="600" height="400">
		<vpc title="VPC" layout="staggered">
			<private-subnet title="A" />
			<private-subnet title="B" />
		</vpc>
	</frame>`
	doc := mustParse(t, src)
	box, err := Build(doc)
	if err != nil {
		t.Fatalf("Build error: %v", err)
	}
	vpc := box.Children[0]
	if len(vpc.Children) != 2 {
		t.Fatalf("expected 2 staggered children, got %d", len(vpc.Children))
	}
	// at least one child should be flagged as stagger background
	hasBg := false
	for _, ch := range vpc.Children {
		if ch.IsStaggerBg {
			hasBg = true
		}
	}
	if !hasBg {
		t.Error("expected at least one stagger background child")
	}
}
