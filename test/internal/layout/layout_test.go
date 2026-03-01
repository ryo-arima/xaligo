package layout_test

import (
	"strings"
	"testing"

	"github.com/ryo-arima/xaligo/internal/layout"
	"github.com/ryo-arima/xaligo/internal/model"
	"github.com/ryo-arima/xaligo/internal/parser"
)

func parseXAL(t *testing.T, xal string) *layout.Box {
	t.Helper()
	doc, err := parser.Parse(strings.NewReader(xal))
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	box, err := layout.Build(doc)
	if err != nil {
		t.Fatalf("build error: %v", err)
	}
	return box
}

func TestBuild_FrameDimensions(t *testing.T) {
	box := parseXAL(t, `<frame width="1280" height="720"></frame>`)
	if box.W != 1280 {
		t.Errorf("W = %v, want 1280", box.W)
	}
	if box.H != 720 {
		t.Errorf("H = %v, want 720", box.H)
	}
	if box.X != 0 {
		t.Errorf("X = %v, want 0", box.X)
	}
	if box.Y != 0 {
		t.Errorf("Y = %v, want 0", box.Y)
	}
}

func TestBuild_DefaultFrameDimensions(t *testing.T) {
	box := parseXAL(t, `<frame></frame>`)
	if box.W != 1280 {
		t.Errorf("default W = %v, want 1280", box.W)
	}
	if box.H != 720 {
		t.Errorf("default H = %v, want 720", box.H)
	}
}

func TestBuild_SingleChild(t *testing.T) {
	box := parseXAL(t, `<frame width="800" height="600"><card title="A" /></frame>`)
	if len(box.Children) != 1 {
		t.Fatalf("expected 1 child, got %d", len(box.Children))
	}
	child := box.Children[0]
	if child.Tag != "card" {
		t.Errorf("child tag = %q, want %q", child.Tag, "card")
	}
	if child.Label != "A" {
		t.Errorf("child label = %q, want %q", child.Label, "A")
	}
}

func TestBuild_TwoChildrenStackedEvenly(t *testing.T) {
	box := parseXAL(t, `<frame width="800" height="600" gap="0"><card title="A" /><card title="B" /></frame>`)
	if len(box.Children) != 2 {
		t.Fatalf("expected 2 children, got %d", len(box.Children))
	}
	a, b := box.Children[0], box.Children[1]
	if a.H != b.H {
		t.Errorf("children heights should be equal: %v != %v", a.H, b.H)
	}
	if b.Y <= a.Y {
		t.Errorf("second child Y (%v) should be greater than first child Y (%v)", b.Y, a.Y)
	}
}

func TestBuild_PaddingReducesInnerArea(t *testing.T) {
	// pa-2 → 2*8=16px padding all sides
	box := parseXAL(t, `<frame width="800" height="600" class="pa-2"><card title="A" /></frame>`)
	if len(box.Children) != 1 {
		t.Fatalf("expected 1 child, got %d", len(box.Children))
	}
	child := box.Children[0]
	// child should be inset by 16px on each side
	if child.X != 16 {
		t.Errorf("child X = %v, want 16", child.X)
	}
	if child.Y != 16 {
		t.Errorf("child Y = %v, want 16", child.Y)
	}
	if child.W != 800-32 {
		t.Errorf("child W = %v, want %v", child.W, 800-32)
	}
}

func TestBuild_RowLayout(t *testing.T) {
	xal := `<frame width="800" height="600">
		<row gap="0">
			<col span="6"><card title="Left" /></col>
			<col span="6"><card title="Right" /></col>
		</row>
	</frame>`
	box := parseXAL(t, xal)
	if len(box.Children) != 1 {
		t.Fatalf("expected 1 child (row), got %d", len(box.Children))
	}
	row := box.Children[0]
	if len(row.Children) != 2 {
		t.Fatalf("expected 2 cols in row, got %d", len(row.Children))
	}
	left, right := row.Children[0], row.Children[1]
	if left.W != right.W {
		t.Errorf("equal span cols should have equal width: %v != %v", left.W, right.W)
	}
	if right.X <= left.X {
		t.Errorf("right col X (%v) should be greater than left col X (%v)", right.X, left.X)
	}
}

func TestBuild_HorizontalLayout(t *testing.T) {
	xal := `<frame width="800" height="600" layout="horizontal" gap="0">
		<card title="A" />
		<card title="B" />
	</frame>`
	box := parseXAL(t, xal)
	if len(box.Children) != 2 {
		t.Fatalf("expected 2 children, got %d", len(box.Children))
	}
	a, b := box.Children[0], box.Children[1]
	if b.X <= a.X {
		t.Errorf("second child X (%v) should be greater than first child X (%v)", b.X, a.X)
	}
	if a.Y != b.Y {
		t.Errorf("horizontal children should have same Y: %v != %v", a.Y, b.Y)
	}
}

func TestBuild_NilRoot_Error(t *testing.T) {
	_, err := layout.Build(model.Document{Root: nil})
	if err == nil {
		t.Error("expected error for nil root")
	}
}

func TestConstants_MinBoxSize(t *testing.T) {
	if layout.MinBoxWidth <= 0 {
		t.Errorf("MinBoxWidth should be positive, got %v", layout.MinBoxWidth)
	}
	if layout.MinBoxHeight <= 0 {
		t.Errorf("MinBoxHeight should be positive, got %v", layout.MinBoxHeight)
	}
}

func TestConstants_GroupInset(t *testing.T) {
	if layout.GroupTopInset <= 0 {
		t.Errorf("GroupTopInset should be positive, got %v", layout.GroupTopInset)
	}
	if layout.GroupSideInset <= 0 {
		t.Errorf("GroupSideInset should be positive, got %v", layout.GroupSideInset)
	}
}
