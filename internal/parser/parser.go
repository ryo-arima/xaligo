package parser

import (
	"encoding/xml"
	"fmt"
	"io"
	"strings"

	"github.com/ryo-arima/xaligo/internal/model"
)

func Parse(r io.Reader) (model.Document, error) {
	dec := xml.NewDecoder(r)
	var stack []*model.Node
	var root *model.Node

	for {
		tok, err := dec.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return model.Document{}, fmt.Errorf("parse xml-like token: %w", err)
		}

		switch t := tok.(type) {
		case xml.StartElement:
			node := &model.Node{Tag: t.Name.Local, Attrs: map[string]string{}}
			for _, a := range t.Attr {
				node.Attrs[a.Name.Local] = a.Value
			}
			if node.Tag == "item" {
				if err := validateItemNode(node); err != nil {
					return model.Document{}, fmt.Errorf("parse <item>: %w", err)
				}
			}
			if node.Tag == "connection" {
				if err := validateConnectionNode(node); err != nil {
					return model.Document{}, fmt.Errorf("parse <connection>: %w", err)
				}
			}
			if len(stack) == 0 {
				root = node
			} else {
				parent := stack[len(stack)-1]
				parent.Children = append(parent.Children, node)
			}
			stack = append(stack, node)
		case xml.CharData:
			if len(stack) == 0 {
				continue
			}
			text := strings.TrimSpace(string(t))
			if text != "" {
				cur := stack[len(stack)-1]
				if cur.Text == "" {
					cur.Text = text
				} else {
					cur.Text += " " + text
				}
			}
		case xml.EndElement:
			if len(stack) == 0 {
				return model.Document{}, fmt.Errorf("unexpected closing tag: %s", t.Name.Local)
			}
			stack = stack[:len(stack)-1]
		}
	}

	if root == nil {
		return model.Document{}, fmt.Errorf("empty document")
	}
	if root.Tag != "frame" {
		return model.Document{}, fmt.Errorf("root tag must be <frame>, got <%s>", root.Tag)
	}

	return model.Document{Root: root}, nil
}

// validateItemNode ensures <item> carries at most one numeric id attribute.
// An empty (or absent) id is allowed — the item acts as a layout spacer.
func validateItemNode(node *model.Node) error {
	id, ok := node.Attrs["id"]
	if !ok || strings.TrimSpace(id) == "" {
		return nil // spacer item — no id required
	}
	if strings.Contains(id, ",") {
		return fmt.Errorf("<item id=%q> must contain a single ID; use separate <item> tags for multiple services", id)
	}
	for _, ch := range strings.TrimSpace(id) {
		if ch < '0' || ch > '9' {
			return fmt.Errorf("<item id=%q> must be a positive integer", id)
		}
	}
	return nil
}

// validateConnectionNode ensures <connection> carries numeric src and dst attributes.
func validateConnectionNode(node *model.Node) error {
	src, hasSrc := node.Attrs["src"]
	dst, hasDst := node.Attrs["dst"]
	if !hasSrc || strings.TrimSpace(src) == "" {
		return fmt.Errorf("<connection> requires a src attribute")
	}
	if !hasDst || strings.TrimSpace(dst) == "" {
		return fmt.Errorf("<connection> requires a dst attribute")
	}
	for _, ch := range strings.TrimSpace(src) {
		if ch < '0' || ch > '9' {
			return fmt.Errorf("<connection src=%q> must be a positive integer", src)
		}
	}
	for _, ch := range strings.TrimSpace(dst) {
		if ch < '0' || ch > '9' {
			return fmt.Errorf("<connection dst=%q> must be a positive integer", dst)
		}
	}
	return nil
}
