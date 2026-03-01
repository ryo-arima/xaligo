package layout

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/ryo-arima/xaligo/internal/model"
)

const spacingUnit = 8

// MinBoxWidth / MinBoxHeight are the smallest dimensions at which a box
// can be meaningfully rendered. layoutStack and layoutRow clamp child
// sizes to these values so boxes are never invisible.
const (
	MinBoxWidth  = 60.0
	MinBoxHeight = 48.0
)

// defaultGroupInset is the automatic padding applied to container nodes
// (AWS group tags and unknown tags with children) when no explicit class
// padding is specified.  The top inset reserves room for the 32 px icon +
// label row; the side inset keeps children clear of the border line.
const (
	defaultGroupTopInset  = 44.0
	defaultGroupSideInset = 12.0

	// GroupTopInset / GroupSideInset are the exported equivalents used by
	// the excalidraw renderer to position item icons below the header row.
	GroupTopInset  = defaultGroupTopInset
	GroupSideInset = defaultGroupSideInset
)

type Box struct {
	ID       string
	Tag      string
	Label    string
	Attrs    map[string]string // raw DSL attributes (e.g. id for <item>)
	X        float64
	Y        float64
	W        float64
	H        float64
	Children []*Box

	// Staggered layout fields (set by layoutStagger)
	StaggerDepth int  // 0 = front, >0 = background depth
	IsStaggerBg  bool // true = background layer → skip rendering children
	InStagger    bool // true = this box participates in a staggered group
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

// layoutKids returns node's children that participate in layout,
// filtering out meta-nodes such as <connection> which are handled separately.
func layoutKids(node *model.Node) []*model.Node {
	var kids []*model.Node
	for _, c := range node.Children {
		if c.Tag == "connection" {
			continue
		}
		kids = append(kids, c)
	}
	return kids
}

func layoutNode(node *model.Node, target *Box, x, y, w, h float64) {
	target.Attrs = node.Attrs
	pad, mar := parseClassSpacing(node.Attr("class"))

	// margin shrinks the allocation passed from the parent (sibling spacing)
	boxX := x + mar.Left
	boxY := y + mar.Top
	boxW := w - mar.Left - mar.Right
	boxH := h - mar.Top - mar.Bottom
	target.X = boxX
	target.Y = boxY
	target.W = boxW
	target.H = boxH

	// padding is the inner margin of the box (offset for child placement)
	innerX := boxX + pad.Left
	innerY := boxY + pad.Top
	innerW := boxW - pad.Left - pad.Right
	innerH := boxH - pad.Top - pad.Bottom

	switch node.Tag {
	case "frame", "container":
		if node.Attr("layout") == "horizontal" {
			layoutFlexH(node, target, innerX, innerY, innerW, innerH)
		} else {
			layoutStack(node, target, innerX, innerY, innerW, innerH)
		}
	case "row":
		layoutRow(node, target, innerX, innerY, innerW, innerH)
	case "col":
		if node.Attr("layout") == "horizontal" {
			layoutFlexH(node, target, innerX, innerY, innerW, innerH)
		} else {
			layoutStack(node, target, innerX, innerY, innerW, innerH)
		}
	default:
		// AWS group tags and other unknown tags:
		// treat as container when they have children, otherwise as leaf.
		kids := layoutKids(node)
		if len(kids) > 0 {
			// Parents whose only children are <item> have no group icon/label,
			// so topInset is not applied — use layoutRow instead.
			allItems := true
			for _, ch := range kids {
				if ch.Tag != "item" {
					allItems = false
					break
				}
			}
			if allItems {
				layoutRow(node, target, innerX, innerY, innerW, innerH)
				break
			}
			// Group inset is always applied; user-specified padding is added on top.
			// This prevents class="pa-2" from causing the header row to overlap children.
			gInnerX := boxX + defaultGroupSideInset + pad.Left
			gInnerY := boxY + defaultGroupTopInset + pad.Top
			gInnerW := boxW - defaultGroupSideInset*2 - pad.Left - pad.Right
			gInnerH := boxH - defaultGroupTopInset - defaultGroupSideInset - pad.Top - pad.Bottom
			if node.Attr("layout") == "staggered" {
				layoutStagger(node, target, gInnerX, gInnerY, gInnerW, gInnerH)
			} else if node.Attr("layout") == "horizontal" {
				layoutFlexH(node, target, gInnerX, gInnerY, gInnerW, gInnerH)
			} else {
				layoutStack(node, target, gInnerX, gInnerY, gInnerW, gInnerH)
			}
		} else {
			layoutLeaf(node, target, innerX, innerY, innerW, innerH)
		}
	}
}

func layoutStack(node *model.Node, target *Box, x, y, w, h float64) {
	children := layoutKids(node)
	if len(children) == 0 {
		return
	}
	gap := attrFloat(node.Attr("gap"), 16)

	// Pre-read each child's margin to compute the total vertical margin.
	// This makes margin work as sibling spacing (CSS-like).
	// The row attribute is a flex-grow style height ratio; default 1.0 (equal).
	totalMarginH := 0.0
	totalRow := 0.0
	for _, child := range children {
		_, childMar := parseClassSpacing(child.Attr("class"))
		totalMarginH += childMar.Top + childMar.Bottom
		totalRow += attrFloat(child.Attr("row"), 1.0)
	}
	availH := h - gap*float64(len(children)-1) - totalMarginH

	curY := y
	for i, child := range children {
		_, childMar := parseClassSpacing(child.Attr("class"))
		row := attrFloat(child.Attr("row"), 1.0)
		// Child allocation = content height proportional to ratio + child's own top/bottom margin
		childH := availH * (row / totalRow)
		alloc := childH + childMar.Top + childMar.Bottom
		cb := &Box{ID: childID(target.ID, i), Tag: child.Tag, Label: labelOf(child)}
		layoutNode(child, cb, x, curY, w, alloc)
		target.Children = append(target.Children, cb)
		curY += alloc + gap
	}
}

// layoutFlexH lays out children horizontally with free ratio weights.
// Each child's width share is determined by its `col` attribute (default 1.0).
// This mirrors layoutStack but in the horizontal direction.
func layoutFlexH(node *model.Node, target *Box, x, y, w, h float64) {
	children := layoutKids(node)
	if len(children) == 0 {
		return
	}
	gap := attrFloat(node.Attr("gap"), 16)

	// Pre-aggregate each child's horizontal margin to compute available width.
	// The col attribute is a flex-grow style width ratio; default 1.0 (equal).
	totalMarginW := 0.0
	totalCol := 0.0
	for _, child := range children {
		_, childMar := parseClassSpacing(child.Attr("class"))
		totalMarginW += childMar.Left + childMar.Right
		totalCol += attrFloat(child.Attr("col"), 1.0)
	}
	availW := w - gap*float64(len(children)-1) - totalMarginW

	curX := x
	for i, child := range children {
		_, childMar := parseClassSpacing(child.Attr("class"))
		col := attrFloat(child.Attr("col"), 1.0)
		// Child allocation = content width proportional to ratio + child's own left/right margin
		childW := availW * (col / totalCol)
		alloc := childW + childMar.Left + childMar.Right
		cb := &Box{ID: childID(target.ID, i), Tag: child.Tag, Label: labelOf(child)}
		layoutNode(child, cb, curX, y, alloc, h)
		target.Children = append(target.Children, cb)
		curX += alloc + gap
	}
}

func layoutRow(node *model.Node, target *Box, x, y, w, h float64) {
	children := layoutKids(node)
	if len(children) == 0 {
		return
	}
	gap := attrFloat(node.Attr("gap"), 16)

	// Pre-read each child's horizontal margin to compute the total width.
	totalMarginW := 0.0
	for _, child := range children {
		_, childMar := parseClassSpacing(child.Attr("class"))
		totalMarginW += childMar.Left + childMar.Right
	}
	remainingW := w - gap*float64(len(children)-1) - totalMarginW
	curX := x

	for i, child := range children {
		_, childMar := parseClassSpacing(child.Attr("class"))
		span := attrFloat(child.Attr("span"), 12/float64(len(children)))
		cw := remainingW*(span/12.0) + childMar.Left + childMar.Right
		cb := &Box{ID: childID(target.ID, i), Tag: child.Tag, Label: labelOf(child)}
		layoutNode(child, cb, curX, y, cw, h)
		target.Children = append(target.Children, cb)
		curX += cw + gap
	}
}

// layoutStagger places children in staggered depth-overlap mode.
// Each child is offset staggerOffset px right-and-down from the previous.
// Children are appended to target.Children in back-to-front render order
// (highest StaggerDepth first = rendered behind, depth 0 last = on top).
// Falls back to layoutStack when fewer than 2 children.
func layoutStagger(node *model.Node, target *Box, x, y, w, h float64) {
	children := layoutKids(node)
	n := len(children)
	if n < 2 {
		layoutStack(node, target, x, y, w, h)
		return
	}
	const staggerOffset = 16.0
	childW := w - staggerOffset*float64(n-1)
	childH := h - staggerOffset*float64(n-1)
	if childW < MinBoxWidth {
		childW = MinBoxWidth
	}
	if childH < MinBoxHeight {
		childH = MinBoxHeight
	}
	// Render back-to-front: highest depth first → behind, depth 0 last → front.
	for i := n - 1; i >= 0; i-- {
		child := children[i]
		cX := x + float64(i)*staggerOffset
		cY := y + float64(i)*staggerOffset
		cb := &Box{
			ID:           childID(target.ID, i),
			Tag:          child.Tag,
			Label:        labelOf(child),
			StaggerDepth: i,
			IsStaggerBg:  i > 0,
			InStagger:    true,
		}
		layoutNode(child, cb, cX, cY, childW, childH)
		target.Children = append(target.Children, cb)
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
		// Axis shorthand: px=left+right, py=top+bottom
		case strings.HasPrefix(tok, "px-"):
			v := spacingValue(tok[3:])
			pad.Left = v
			pad.Right = v
		case strings.HasPrefix(tok, "py-"):
			v := spacingValue(tok[3:])
			pad.Top = v
			pad.Bottom = v
		case strings.HasPrefix(tok, "mx-"):
			v := spacingValue(tok[3:])
			mar.Left = v
			mar.Right = v
		case strings.HasPrefix(tok, "my-"):
			v := spacingValue(tok[3:])
			mar.Top = v
			mar.Bottom = v
		// Individual directions
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
