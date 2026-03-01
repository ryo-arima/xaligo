package model_test

import (
	"testing"

	"github.com/ryo-arima/xaligo/internal/model"
)

func TestNode_Attr_ReturnsValue(t *testing.T) {
	n := &model.Node{
		Tag:   "frame",
		Attrs: map[string]string{"width": "1280", "height": "720"},
	}
	if got := n.Attr("width"); got != "1280" {
		t.Errorf("Attr(width) = %q, want %q", got, "1280")
	}
	if got := n.Attr("height"); got != "720" {
		t.Errorf("Attr(height) = %q, want %q", got, "720")
	}
}

func TestNode_Attr_MissingKey(t *testing.T) {
	n := &model.Node{Tag: "frame", Attrs: map[string]string{}}
	if got := n.Attr("nonexistent"); got != "" {
		t.Errorf("Attr(nonexistent) = %q, want empty string", got)
	}
}

func TestNode_Attr_NilNode(t *testing.T) {
	var n *model.Node
	if got := n.Attr("any"); got != "" {
		t.Errorf("nil.Attr(any) = %q, want empty string", got)
	}
}

func TestNode_Attr_NilAttrs(t *testing.T) {
	n := &model.Node{Tag: "frame", Attrs: nil}
	if got := n.Attr("any"); got != "" {
		t.Errorf("Attr(any) with nil Attrs = %q, want empty string", got)
	}
}

func TestDocument_RootNode(t *testing.T) {
	root := &model.Node{Tag: "frame", Attrs: map[string]string{}}
	doc := model.Document{Root: root}
	if doc.Root != root {
		t.Error("Document.Root should point to the provided node")
	}
}

func TestNode_Children(t *testing.T) {
	parent := &model.Node{Tag: "frame", Attrs: map[string]string{}}
	child := &model.Node{Tag: "container", Attrs: map[string]string{}}
	parent.Children = append(parent.Children, child)
	if len(parent.Children) != 1 {
		t.Fatalf("expected 1 child, got %d", len(parent.Children))
	}
	if parent.Children[0].Tag != "container" {
		t.Errorf("child tag = %q, want %q", parent.Children[0].Tag, "container")
	}
}
