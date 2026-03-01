package parser_test

import (
	"strings"
	"testing"

	"github.com/ryo-arima/xaligo/internal/parser"
)

func TestParse_SimpleFrame(t *testing.T) {
	input := `<frame width="1280" height="720"></frame>`
	doc, err := parser.Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if doc.Root == nil {
		t.Fatal("expected non-nil root")
	}
	if doc.Root.Tag != "frame" {
		t.Errorf("root tag = %q, want %q", doc.Root.Tag, "frame")
	}
	if doc.Root.Attr("width") != "1280" {
		t.Errorf("width = %q, want %q", doc.Root.Attr("width"), "1280")
	}
	if doc.Root.Attr("height") != "720" {
		t.Errorf("height = %q, want %q", doc.Root.Attr("height"), "720")
	}
}

func TestParse_NestedChildren(t *testing.T) {
	input := `<frame width="1280" height="720">
		<container>
			<card title="Hello" />
		</container>
	</frame>`
	doc, err := parser.Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(doc.Root.Children) != 1 {
		t.Fatalf("expected 1 child, got %d", len(doc.Root.Children))
	}
	container := doc.Root.Children[0]
	if container.Tag != "container" {
		t.Errorf("child tag = %q, want %q", container.Tag, "container")
	}
	if len(container.Children) != 1 {
		t.Fatalf("expected 1 grandchild, got %d", len(container.Children))
	}
	card := container.Children[0]
	if card.Attr("title") != "Hello" {
		t.Errorf("card title = %q, want %q", card.Attr("title"), "Hello")
	}
}

func TestParse_TextContent(t *testing.T) {
	input := `<frame width="640" height="480"><text>hello world</text></frame>`
	doc, err := parser.Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(doc.Root.Children) != 1 {
		t.Fatalf("expected 1 child, got %d", len(doc.Root.Children))
	}
	if doc.Root.Children[0].Text != "hello world" {
		t.Errorf("text = %q, want %q", doc.Root.Children[0].Text, "hello world")
	}
}

func TestParse_NonFrameRoot_Error(t *testing.T) {
	input := `<container width="1280" height="720"></container>`
	_, err := parser.Parse(strings.NewReader(input))
	if err == nil {
		t.Error("expected error for non-frame root tag")
	}
}

func TestParse_EmptyDocument_Error(t *testing.T) {
	input := ``
	_, err := parser.Parse(strings.NewReader(input))
	if err == nil {
		t.Error("expected error for empty document")
	}
}

func TestParse_ItemWithValidID(t *testing.T) {
	input := `<frame width="640" height="480"><item id="1234" /></frame>`
	_, err := parser.Parse(strings.NewReader(input))
	if err != nil {
		t.Errorf("unexpected error for valid item id: %v", err)
	}
}

func TestParse_ItemWithNoID(t *testing.T) {
	input := `<frame width="640" height="480"><item /></frame>`
	_, err := parser.Parse(strings.NewReader(input))
	if err != nil {
		t.Errorf("unexpected error for item with no id (spacer): %v", err)
	}
}

func TestParse_ItemWithMultipleIDs_Error(t *testing.T) {
	input := `<frame width="640" height="480"><item id="1,2" /></frame>`
	_, err := parser.Parse(strings.NewReader(input))
	if err == nil {
		t.Error("expected error for item with multiple IDs")
	}
}

func TestParse_ItemWithNonNumericID_Error(t *testing.T) {
	input := `<frame width="640" height="480"><item id="abc" /></frame>`
	_, err := parser.Parse(strings.NewReader(input))
	if err == nil {
		t.Error("expected error for item with non-numeric id")
	}
}

func TestParse_ConnectionValid(t *testing.T) {
	input := `<frame width="640" height="480">
		<connection src="10" dst="20" />
	</frame>`
	doc, err := parser.Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	found := false
	for _, c := range doc.Root.Children {
		if c.Tag == "connection" {
			found = true
			if c.Attr("src") != "10" {
				t.Errorf("connection src = %q, want %q", c.Attr("src"), "10")
			}
			if c.Attr("dst") != "20" {
				t.Errorf("connection dst = %q, want %q", c.Attr("dst"), "20")
			}
		}
	}
	if !found {
		t.Error("expected to find connection child node")
	}
}

func TestParse_ConnectionMissingSrc_Error(t *testing.T) {
	input := `<frame width="640" height="480"><connection dst="20" /></frame>`
	_, err := parser.Parse(strings.NewReader(input))
	if err == nil {
		t.Error("expected error for connection missing src")
	}
}

func TestParse_ConnectionMissingDst_Error(t *testing.T) {
	input := `<frame width="640" height="480"><connection src="10" /></frame>`
	_, err := parser.Parse(strings.NewReader(input))
	if err == nil {
		t.Error("expected error for connection missing dst")
	}
}
