package parser

import (
	"strings"
	"testing"
)

func TestParseSimpleFrame(t *testing.T) {
	src := `<frame width="800" height="600"></frame>`
	doc, err := Parse(strings.NewReader(src))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if doc.Root == nil {
		t.Fatal("root is nil")
	}
	if doc.Root.Tag != "frame" {
		t.Errorf("root tag = %q, want %q", doc.Root.Tag, "frame")
	}
	if doc.Root.Attr("width") != "800" {
		t.Errorf("width = %q, want %q", doc.Root.Attr("width"), "800")
	}
	if doc.Root.Attr("height") != "600" {
		t.Errorf("height = %q, want %q", doc.Root.Attr("height"), "600")
	}
}

func TestParseChildren(t *testing.T) {
	src := `<frame width="1280" height="720"><card title="Test" /></frame>`
	doc, err := Parse(strings.NewReader(src))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(doc.Root.Children) != 1 {
		t.Fatalf("expected 1 child, got %d", len(doc.Root.Children))
	}
	child := doc.Root.Children[0]
	if child.Tag != "card" {
		t.Errorf("child tag = %q, want %q", child.Tag, "card")
	}
	if child.Attr("title") != "Test" {
		t.Errorf("child title = %q, want %q", child.Attr("title"), "Test")
	}
}

func TestParseTextContent(t *testing.T) {
	src := `<frame width="1280" height="720"><text>Hello World</text></frame>`
	doc, err := Parse(strings.NewReader(src))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	child := doc.Root.Children[0]
	if child.Text != "Hello World" {
		t.Errorf("text = %q, want %q", child.Text, "Hello World")
	}
}

func TestParseNonFrameRoot(t *testing.T) {
	src := `<container></container>`
	_, err := Parse(strings.NewReader(src))
	if err == nil {
		t.Fatal("expected error for non-frame root")
	}
}

func TestParseEmptyDocument(t *testing.T) {
	_, err := Parse(strings.NewReader(""))
	if err == nil {
		t.Fatal("expected error for empty document")
	}
}

func TestParseItemValid(t *testing.T) {
	src := `<frame><item id="1178" /></frame>`
	_, err := Parse(strings.NewReader(src))
	if err != nil {
		t.Errorf("unexpected error for valid item: %v", err)
	}
}

func TestParseItemSpacer(t *testing.T) {
	src := `<frame><item /></frame>`
	_, err := Parse(strings.NewReader(src))
	if err != nil {
		t.Errorf("unexpected error for spacer item: %v", err)
	}
}

func TestParseItemInvalidID(t *testing.T) {
	src := `<frame><item id="abc" /></frame>`
	_, err := Parse(strings.NewReader(src))
	if err == nil {
		t.Fatal("expected error for non-numeric item id")
	}
}

func TestParseItemMultipleIDs(t *testing.T) {
	src := `<frame><item id="1,2" /></frame>`
	_, err := Parse(strings.NewReader(src))
	if err == nil {
		t.Fatal("expected error for comma-separated item ids")
	}
}

func TestParseConnectionValid(t *testing.T) {
	src := `<frame><connection src="1178" dst="1189" /></frame>`
	_, err := Parse(strings.NewReader(src))
	if err != nil {
		t.Errorf("unexpected error for valid connection: %v", err)
	}
}

func TestParseConnectionMissingSrc(t *testing.T) {
	src := `<frame><connection dst="1189" /></frame>`
	_, err := Parse(strings.NewReader(src))
	if err == nil {
		t.Fatal("expected error for missing src")
	}
}

func TestParseConnectionMissingDst(t *testing.T) {
	src := `<frame><connection src="1178" /></frame>`
	_, err := Parse(strings.NewReader(src))
	if err == nil {
		t.Fatal("expected error for missing dst")
	}
}

func TestParseConnectionInvalidSrc(t *testing.T) {
	src := `<frame><connection src="abc" dst="1189" /></frame>`
	_, err := Parse(strings.NewReader(src))
	if err == nil {
		t.Fatal("expected error for non-numeric src")
	}
}

func TestParseConnectionInvalidDst(t *testing.T) {
	src := `<frame><connection src="1178" dst="xyz" /></frame>`
	_, err := Parse(strings.NewReader(src))
	if err == nil {
		t.Fatal("expected error for non-numeric dst")
	}
}

func TestParseNestedStructure(t *testing.T) {
	src := `<frame width="1440" height="900">
		<row gap="20">
			<col span="8"><card title="A" /></col>
			<col span="4"><card title="B" /></col>
		</row>
	</frame>`
	doc, err := Parse(strings.NewReader(src))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(doc.Root.Children) != 1 {
		t.Fatalf("expected 1 child, got %d", len(doc.Root.Children))
	}
	row := doc.Root.Children[0]
	if row.Tag != "row" {
		t.Errorf("expected row tag, got %q", row.Tag)
	}
	if len(row.Children) != 2 {
		t.Errorf("expected 2 cols, got %d", len(row.Children))
	}
}
