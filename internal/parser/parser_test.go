package parser

import (
	"strings"
	"testing"
)

func TestParse_SimpleFrame(t *testing.T) {
	src := `<frame width="800" height="600"></frame>`
	doc, err := Parse(strings.NewReader(src))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if doc.Root == nil {
		t.Fatal("expected non-nil root")
	}
	if doc.Root.Tag != "frame" {
		t.Errorf("root tag = %q, want %q", doc.Root.Tag, "frame")
	}
	if doc.Root.Attr("width") != "800" {
		t.Errorf("width = %q, want %q", doc.Root.Attr("width"), "800")
	}
}

func TestParse_NestedChildren(t *testing.T) {
	src := `<frame width="1280" height="720">
		<container>
			<card title="Hello" />
		</container>
	</frame>`
	doc, err := Parse(strings.NewReader(src))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(doc.Root.Children) != 1 {
		t.Fatalf("expected 1 child of frame, got %d", len(doc.Root.Children))
	}
	container := doc.Root.Children[0]
	if container.Tag != "container" {
		t.Errorf("child tag = %q, want container", container.Tag)
	}
	if len(container.Children) != 1 {
		t.Fatalf("expected 1 child of container, got %d", len(container.Children))
	}
	card := container.Children[0]
	if card.Attr("title") != "Hello" {
		t.Errorf("card title = %q, want Hello", card.Attr("title"))
	}
}

func TestParse_TextContent(t *testing.T) {
	src := `<frame><text>My Label</text></frame>`
	doc, err := Parse(strings.NewReader(src))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	txt := doc.Root.Children[0]
	if txt.Text != "My Label" {
		t.Errorf("text content = %q, want %q", txt.Text, "My Label")
	}
}

func TestParse_ErrorOnNonFrameRoot(t *testing.T) {
	src := `<container></container>`
	_, err := Parse(strings.NewReader(src))
	if err == nil {
		t.Fatal("expected error for non-frame root, got nil")
	}
}

func TestParse_ErrorOnEmptyDocument(t *testing.T) {
	_, err := Parse(strings.NewReader(""))
	if err == nil {
		t.Fatal("expected error for empty document, got nil")
	}
}

func TestParse_ErrorOnInvalidXML(t *testing.T) {
	src := `<frame><unclosed></frame>`
	_, err := Parse(strings.NewReader(src))
	if err == nil {
		t.Fatal("expected error for invalid XML, got nil")
	}
}

func TestParse_ItemSpacerAllowed(t *testing.T) {
	src := `<frame><item /></frame>`
	_, err := Parse(strings.NewReader(src))
	if err != nil {
		t.Fatalf("spacer item should be allowed, got error: %v", err)
	}
}

func TestParse_ItemWithValidID(t *testing.T) {
	src := `<frame><item id="1178" /></frame>`
	_, err := Parse(strings.NewReader(src))
	if err != nil {
		t.Fatalf("item with numeric id should be valid, got error: %v", err)
	}
}

func TestParse_ItemWithMultipleIDs(t *testing.T) {
	src := `<frame><item id="1,2" /></frame>`
	_, err := Parse(strings.NewReader(src))
	if err == nil {
		t.Fatal("expected error for item with multiple IDs")
	}
}

func TestParse_ItemWithNonNumericID(t *testing.T) {
	src := `<frame><item id="abc" /></frame>`
	_, err := Parse(strings.NewReader(src))
	if err == nil {
		t.Fatal("expected error for item with non-numeric id")
	}
}

func TestParse_ConnectionValid(t *testing.T) {
	src := `<frame><connection src="1178" dst="1189" /></frame>`
	_, err := Parse(strings.NewReader(src))
	if err != nil {
		t.Fatalf("valid connection should parse without error, got: %v", err)
	}
}

func TestParse_ConnectionMissingSrc(t *testing.T) {
	src := `<frame><connection dst="1189" /></frame>`
	_, err := Parse(strings.NewReader(src))
	if err == nil {
		t.Fatal("expected error for connection missing src")
	}
}

func TestParse_ConnectionMissingDst(t *testing.T) {
	src := `<frame><connection src="1178" /></frame>`
	_, err := Parse(strings.NewReader(src))
	if err == nil {
		t.Fatal("expected error for connection missing dst")
	}
}

func TestParse_ConnectionNonNumericSrc(t *testing.T) {
	src := `<frame><connection src="abc" dst="1189" /></frame>`
	_, err := Parse(strings.NewReader(src))
	if err == nil {
		t.Fatal("expected error for connection with non-numeric src")
	}
}

func TestParse_ConnectionNonNumericDst(t *testing.T) {
	src := `<frame><connection src="1178" dst="xyz" /></frame>`
	_, err := Parse(strings.NewReader(src))
	if err == nil {
		t.Fatal("expected error for connection with non-numeric dst")
	}
}
