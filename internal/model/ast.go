package model

// Document is the root of xaligo DSL.
type Document struct {
	Root *Node
}

// Node is a Vue-like tag node.
type Node struct {
	Tag      string
	Attrs    map[string]string
	Children []*Node
	Text     string
}

func (n *Node) Attr(key string) string {
	if n == nil || n.Attrs == nil {
		return ""
	}
	return n.Attrs[key]
}
