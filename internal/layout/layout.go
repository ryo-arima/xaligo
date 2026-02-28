package layout

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/ryo-arima/xaligo/internal/model"
)

const spacingUnit = 8

// defaultGroupInset is the automatic padding applied to container nodes
// (AWS group tags and unknown tags with children) when no explicit class
// padding is specified.  The top inset reserves room for the 32 px icon +
// label row; the side inset keeps children clear of the border line.
const (
	defaultGroupTopInset  = 44.0
	defaultGroupSideInset = 12.0
)

type Box struct {
	ID       string
	Tag      string
	Label    string
	X        float64
	Y        float64
	W        float64
	H        float64
	Children []*Box
}

type Spacing struct {
	Top    float64
	Right  float64
	Bottom float64
	Left   float64
}

func Build(doc model.Document) (*Box, error) {
	if doc.Root == nil {
		return nil, fmt.Errorf("document root is nil")
	}
	w := attrFloat(doc.Root.Attr("width"), 1280)
	h := attrFloat(doc.Root.Attr("height"), 720)
	root := &Box{ID: "frame", Tag: "frame", Label: "frame", X: 0, Y: 0, W: w, H: h}
	layoutNode(doc.Root, root, 0, 0, w, h)
	return root, nil
}

func layoutNode(node *model.Node, target *Box, x, y, w, h float64) {
	pad, mar := parseClassSpacing(node.Attr("class"))
	innerX := x + mar.Left + pad.Left
	innerY := y + mar.Top + pad.Top
	innerW := w - mar.Left - mar.Right - pad.Left - pad.Right
	innerH := h - mar.Top - mar.Bottom - pad.Top - pad.Bottom

	target.X = x + mar.Left
	target.Y = y + mar.Top
	target.W = w - mar.Left - mar.Right
	target.H = h - mar.Top - mar.Bottom

	switch node.Tag {
	case "frame", "container":
		layoutStack(node, target, innerX, innerY, innerW, innerH)
	case "row":
		layoutRow(node, target, innerX, innerY, innerW, innerH)
	case "col":
		layoutStack(node, target, innerX, innerY, innerW, innerH)
	default:
		// AWS group tags (vpc, region, aws-cloud, etc.) and other unknown tags:
		// if they have children act as a container, otherwise as a leaf.
		if len(node.Children) > 0 {
			// If no explicit padding was given via class="...", apply the default
			// group inset so children don't overlap the parent's border/icon area.
			if pad == (Spacing{}) {
				innerX = target.X + defaultGroupSideInset
				innerY = target.Y + defaultGroupTopInset
				innerW = target.W - defaultGroupSideInset*2
				innerH = target.H - defaultGroupTopInset - defaultGroupSideInset
			}
			layoutStack(node, target, innerX, innerY, innerW, innerH)
		} else {
			layoutLeaf(node, target, innerX, innerY, innerW, innerH)
		}
	}
}

func layoutStack(node *model.Node, target *Box, x, y, w, h float64) {
	if len(node.Children) == 0 {
		return
	}
	gap := attrFloat(node.Attr("gap"), 16)
	childH := (h - gap*float64(len(node.Children)-1)) / float64(len(node.Children))
	curY := y
	for i, child := range node.Children {
		cb := &Box{ID: childID(target.ID, i), Tag: child.Tag, Label: labelOf(child)}
		layoutNode(child, cb, x, curY, w, childH)
		target.Children = append(target.Children, cb)
		curY += childH + gap
	}
}

func layoutRow(node *model.Node, target *Box, x, y, w, h float64) {
	if len(node.Children) == 0 {
		return
	}
	gap := attrFloat(node.Attr("gap"), 16)
	remainingW := w - gap*float64(len(node.Children)-1)
	curX := x

	for i, child := range node.Children {
		span := attrFloat(child.Attr("span"), 12/float64(len(node.Children)))
		cw := remainingW * (span / 12.0)
		cb := &Box{ID: childID(target.ID, i), Tag: child.Tag, Label: labelOf(child)}
		layoutNode(child, cb, curX, y, cw, h)
		target.Children = append(target.Children, cb)
		curX += cw + gap
	}
}

func layoutLeaf(node *model.Node, target *Box, x, y, w, h float64) {
	target.X = x
	target.Y = y
	target.W = w
	target.H = h
}

func childID(parent string, index int) string {
	return fmt.Sprintf("%s-%d", parent, index)
}

func labelOf(n *model.Node) string {
	if title := n.Attr("title"); title != "" {
		return title
	}
	if n.Text != "" {
		return n.Text
	}
	return n.Tag
}

func attrFloat(v string, fallback float64) float64 {
	if strings.TrimSpace(v) == "" {
		return fallback
	}
	f, err := strconv.ParseFloat(v, 64)
	if err != nil {
		return fallback
	}
	return f
}

func parseClassSpacing(class string) (Spacing, Spacing) {
	pad := Spacing{}
	mar := Spacing{}
	for _, tok := range strings.Fields(class) {
		switch {
		case strings.HasPrefix(tok, "pa-"):
			v := spacingValue(tok[3:])
			pad = Spacing{Top: v, Right: v, Bottom: v, Left: v}
		case strings.HasPrefix(tok, "ma-"):
			v := spacingValue(tok[3:])
			mar = Spacing{Top: v, Right: v, Bottom: v, Left: v}
		case strings.HasPrefix(tok, "pt-"):
			pad.Top = spacingValue(tok[3:])
		case strings.HasPrefix(tok, "pr-"):
			pad.Right = spacingValue(tok[3:])
		case strings.HasPrefix(tok, "pb-"):
			pad.Bottom = spacingValue(tok[3:])
		case strings.HasPrefix(tok, "pl-"):
			pad.Left = spacingValue(tok[3:])
		case strings.HasPrefix(tok, "mt-"):
			mar.Top = spacingValue(tok[3:])
		case strings.HasPrefix(tok, "mr-"):
			mar.Right = spacingValue(tok[3:])
		case strings.HasPrefix(tok, "mb-"):
			mar.Bottom = spacingValue(tok[3:])
		case strings.HasPrefix(tok, "ml-"):
			mar.Left = spacingValue(tok[3:])
		}
	}
	return pad, mar
}

func spacingValue(s string) float64 {
	n, err := strconv.Atoi(s)
	if err != nil {
		return 0
	}
	return float64(n * spacingUnit)
}
