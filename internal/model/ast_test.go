package model

import "testing"

func TestNodeAttr(t *testing.T) {
	t.Run("returns attribute value when present", func(t *testing.T) {
		n := &Node{Attrs: map[string]string{"title": "hello"}}
		if got := n.Attr("title"); got != "hello" {
			t.Errorf("Attr(title) = %q, want %q", got, "hello")
		}
	})

	t.Run("returns empty string when key is absent", func(t *testing.T) {
		n := &Node{Attrs: map[string]string{}}
		if got := n.Attr("missing"); got != "" {
			t.Errorf("Attr(missing) = %q, want %q", got, "")
		}
	})

	t.Run("returns empty string for nil Attrs", func(t *testing.T) {
		n := &Node{}
		if got := n.Attr("x"); got != "" {
			t.Errorf("Attr on nil Attrs = %q, want empty", got)
		}
	})

	t.Run("returns empty string for nil Node", func(t *testing.T) {
		var n *Node
		if got := n.Attr("x"); got != "" {
			t.Errorf("Attr on nil Node = %q, want empty", got)
		}
	})
}
