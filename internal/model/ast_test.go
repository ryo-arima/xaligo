package model

import "testing"

func TestNodeAttr(t *testing.T) {
	n := &Node{
		Tag:   "frame",
		Attrs: map[string]string{"width": "1280", "height": "720"},
	}
	if got := n.Attr("width"); got != "1280" {
		t.Errorf("Attr(width) = %q, want %q", got, "1280")
	}
	if got := n.Attr("missing"); got != "" {
		t.Errorf("Attr(missing) = %q, want empty", got)
	}
}

func TestNodeAttrNilNode(t *testing.T) {
	var n *Node
	if got := n.Attr("anything"); got != "" {
		t.Errorf("nil node Attr = %q, want empty", got)
	}
}

func TestNodeAttrNilAttrs(t *testing.T) {
	n := &Node{Tag: "frame"}
	if got := n.Attr("width"); got != "" {
		t.Errorf("nil attrs Attr = %q, want empty", got)
	}
}
