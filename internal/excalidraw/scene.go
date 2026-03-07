package excalidraw

import (
	"crypto/md5"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/fs"
	"math"
	"math/rand"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/ryo-arima/xaligo/internal/entity"
	"github.com/ryo-arima/xaligo/internal/layout"
	"github.com/ryo-arima/xaligo/internal/model"
	"github.com/ryo-arima/xaligo/internal/repository"
)

type file struct {
	Type     string           `json:"type"`
	Version  int              `json:"version"`
	Source   string           `json:"source"`
	Elements []map[string]any `json:"elements"`
	AppState map[string]any   `json:"appState"`
	Files    map[string]any   `json:"files"`
}

// groupDef holds visual style for an AWS architecture group tag.
type groupDef struct {
	StrokeColor string
	StrokeStyle string
	StrokeWidth int
	IconFile    string // filename inside Architecture-Group-Icons dir, empty = no icon
}

// awsGroups maps xal tag names to their AWS group visual definitions.
var awsGroups = map[string]groupDef{
	"aws-cloud":                     {"#000000", "solid", 2, "AWS-Cloud-logo_32.svg"},
	"aws-cloud-alt":                 {"#000000", "solid", 2, "AWS-Cloud_32.svg"},
	"region":                        {"#00A1C9", "dashed", 2, "Region_32.svg"},
	"availability-zone":             {"#00A1C9", "dashed", 2, ""},
	"security-group":                {"#CC0000", "dashed", 2, ""},
	"auto-scaling-group":            {"#E7601B", "dashed", 2, "Auto-Scaling-group_32.svg"},
	"vpc":                           {"#8C4FFF", "solid", 2, "Virtual-private-cloud-VPC_32.svg"},
	"private-subnet":                {"#00A1C9", "solid", 2, "Private-subnet_32.svg"},
	"public-subnet":                 {"#3F8624", "solid", 2, "Public-subnet_32.svg"},
	"server-contents":               {"#7A7C7F", "solid", 2, "Server-contents_32.svg"},
	"corporate-data-center":         {"#7A7C7F", "solid", 2, "Corporate-data-center_32.svg"},
	"ec2-instance-contents":         {"#E7601B", "solid", 2, "EC2-instance-contents_32.svg"},
	"spot-fleet":                    {"#E7601B", "solid", 2, "Spot-Fleet_32.svg"},
	"aws-account":                   {"#E7008A", "solid", 2, "AWS-Account_32.svg"},
	"aws-iot-greengrass-deployment": {"#3F8624", "solid", 2, "AWS-IoT-Greengrass-Deployment_32.svg"},
	"aws-iot-greengrass":            {"#3F8624", "solid", 2, ""},
	"elastic-beanstalk-container":   {"#E7601B", "solid", 2, ""},
	"aws-step-functions-workflow":   {"#E7008A", "solid", 2, ""},
	"generic-group":                 {"#AAB7B8", "dashed", 1, ""},
}

const (
	groupIconSize   = 32
	groupFontSize   = 14
	groupFontFamily = 2 // Helvetica (normal)
)

// staggerFills are background fill colors for staggered AZ layers.
// Index = StaggerDepth (0 = front/white, 1/2 = progressively darker teal).
var staggerFills = []string{"#ffffff", "#c8e8e8", "#92cecd"}

// staggerBGColor returns the appropriate backgroundColor for a box.
// Boxes that participate in a staggered group get a solid fill so that
// overlapping back-layers are visually distinct.
func staggerBGColor(b *layout.Box) string {
	if !b.InStagger {
		return "transparent"
	}
	idx := b.StaggerDepth
	if idx >= len(staggerFills) {
		idx = len(staggerFills) - 1
	}
	return staggerFills[idx]
}

const (
	itemMaxSize  = 32.0
	itemMinSize  = 16.0
	itemLabelH   = 14.0
	itemLabelW   = 56.0 // text box width for item labels (wider than icon, centred on icon)
	itemGap      = 8.0
)

// paperSizeNames maps (short-side, long-side) → paper name for reverse lookup.
var paperSizeNames = map[[2]int]string{
	{559, 794}:   "A5",
	{794, 1122}:  "A4",
	{1122, 1587}: "A3",
	{1587, 2245}: "A2",
	{2245, 3179}: "A1",
	{816, 1056}:  "Letter",
	{816, 1344}:  "Legal",
	{1056, 1632}: "Tabloid",
}

// detectPaperName returns e.g. "A4 landscape" / "A4 portrait" from box dimensions.
func detectPaperName(w, h float64) string {
	wi, hi := int(w), int(h)
	short, long := wi, hi
	orientation := "portrait"
	if wi > hi {
		short, long = hi, wi
		orientation = "landscape"
	}
	if name, ok := paperSizeNames[[2]int{short, long}]; ok {
		return name + " " + orientation
	}
	return fmt.Sprintf("%d×%d", wi, hi)
}

func svgFileID(name string) string {
	h := md5.Sum([]byte(name))
	return fmt.Sprintf("%x", h)[:16]
}

func svgDataURL(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return "data:image/svg+xml;base64," + base64.StdEncoding.EncodeToString(data), nil
}

// svgDataURLFS reads an SVG from an fs.FS and returns it as a base64 data URL.
func svgDataURLFS(fsys fs.FS, path string) (string, error) {
	data, err := fs.ReadFile(fsys, path)
	if err != nil {
		return "", err
	}
	return "data:image/svg+xml;base64," + base64.StdEncoding.EncodeToString(data), nil
}

// BuildJSONWithFS is a convenience wrapper for WASM / embedded builds.
// It uses fsys (typically an embed.FS) for all asset reads instead of the OS
// filesystem.  catalogCSV and svgGroupDir are resolved relative to the root
// of fsys (e.g. "service-catalog.csv" and "svg/Architecture-Group-Icons").
func BuildJSONWithFS(root *layout.Box, fsys fs.FS, catalogCSV, svgGroupDir string, itemIconSize float64, connections []*model.Node, abbrevMap map[int]string) ([]byte, error) {
	return BuildJSON(root, svgGroupDir, catalogCSV, "", itemIconSize, connections, abbrevMap, fsys)
}

// BuildJSON converts a Box layout tree into Excalidraw JSON.
// svgGroupDir:  absolute path to Architecture-Group-Icons/ (or FS-relative path when fsys≠nil)
// catalogCSV:   absolute path to service-catalog.csv (or FS-relative path when fsys≠nil)
// projectRoot:  project root directory (used to resolve rel_path from catalog; ignored when fsys≠nil)
// itemIconSize: default maximum icon size (px) for <item> elements.
// connections:  <connection> nodes extracted from the DSL (may be nil).
// abbrevMap:    optional catalog-ID → abbreviation map derived from services.csv.
// fsys:         when non-nil, all asset reads go through this fs.FS (WASM / embedded mode).
func BuildJSON(root *layout.Box, svgGroupDir string, catalogCSV string, projectRoot string, itemIconSize float64, connections []*model.Node, abbrevMap map[int]string, fsys fs.FS) ([]byte, error) {
	if root == nil {
		return nil, fmt.Errorf("root layout is nil")
	}
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	updated := time.Now().UnixMilli()

	// Outermost Excalidraw frame element representing the paper size.
	frameElem := map[string]any{
		"id": "paper-frame", "type": "frame",
		"x": root.X, "y": root.Y, "width": root.W, "height": root.H,
		"angle": 0,
		"name":        detectPaperName(root.W, root.H),
		"strokeColor": "#bbb", "backgroundColor": "transparent",
		"fillStyle": "solid", "strokeWidth": 1, "strokeStyle": "solid",
		"roughness": 0, "opacity": 100,
		"groupIds": []string{}, "roundness": nil,
		"seed": r.Intn(99999999), "version": 1,
		"versionNonce": r.Intn(99999999),
		"isDeleted": false, "boundElements": nil,
		"updated": updated, "link": nil, "locked": false,
	}

	var elements []map[string]any
	elements = append(elements, frameElem)
	files := map[string]any{}

	// 2パス: 1) item を visibleAncestorID ごとに収集, 2) グリッド一括描画
	itemGroups := map[string][]*layout.Box{}
	ancestorBoxes := map[string]*layout.Box{}
	// <frame item-size="N"> overrides the global itemIconSize.
	if v := root.Attrs["item-size"]; v != "" {
		if f, err := strconv.ParseFloat(v, 64); err == nil && f > 0 {
			itemIconSize = f
		}
	}

	// itemImgRects / itemLblRects / itemImgIDs / itemLblIDs:
	// catalog ID → bounding rect (x,y,w,h) and element ID of the image / label elements.
	// Populated during renderItemGrid → renderIconAt, used for edge-based connections.
	itemImgRects := map[int][4]float64{}
	itemLblRects := map[int][4]float64{}
	itemImgIDs   := map[int]string{}
	itemLblIDs   := map[int]string{}

	walk(root, &elements, files, svgGroupDir, catalogCSV, projectRoot, fsys, r, root, itemGroups, ancestorBoxes)
	for ancID, items := range itemGroups {
		renderItemGrid(items, ancestorBoxes[ancID], &elements, files, catalogCSV, projectRoot, fsys, itemIconSize, r, itemImgRects, itemLblRects, itemImgIDs, itemLblIDs, abbrevMap)
	}
	renderConnections(connections, itemImgRects, itemLblRects, itemImgIDs, itemLblIDs, &elements, r)

	out := file{
		Type:    "excalidraw",
		Version: 2,
		Source:  "https://github.com/ryo-arima/xaligo",
		Elements: elements,
		AppState: map[string]any{
			"gridSize":            20,
			"viewBackgroundColor": "#ffffff",
		},
		Files: files,
	}
	return json.MarshalIndent(out, "", "  ")
}

func walk(b *layout.Box, elements *[]map[string]any, files map[string]any, svgGroupDir string, catalogCSV string, projectRoot string, fsys fs.FS, r *rand.Rand, visibleAncestor *layout.Box, itemGroups map[string][]*layout.Box, ancestorBoxes map[string]*layout.Box) {
	if layout.IsItemLike(b.Tag) {
		// 描画はしない: visibleAncestor に結び付けて収集のみ (<item> / <spacer> 共通)
		key := visibleAncestor.ID
		itemGroups[key] = append(itemGroups[key], b)
		ancestorBoxes[key] = visibleAncestor
		return
	}

	// selfVisible=false のとき: 自身の描画 (枠・アイコン・ラベル) はスキップするが
	// 子要素の描画は継続する (親子関係なく個別に制御可能)。
	selfVisible := b.Attrs["visible"] != "false"

	if b.Tag != "frame" && (b.W < layout.MinBoxWidth || b.H < layout.MinBoxHeight) {
		fmt.Fprintf(os.Stderr,
			"WARNING: skipping %q (%s) — too small to display (%.1f x %.1f, min %.0f x %.0f)\n",
			b.Label, b.Tag, b.W, b.H, layout.MinBoxWidth, layout.MinBoxHeight)
		// 子の item も同じ visibleAncestor に結び付けて収集
		for _, c := range b.Children {
			if layout.IsItemLike(c.Tag) {
				key := visibleAncestor.ID
				itemGroups[key] = append(itemGroups[key], c)
				ancestorBoxes[key] = visibleAncestor
			} else {
				walk(c, elements, files, svgGroupDir, catalogCSV, projectRoot, fsys, r, visibleAncestor, itemGroups, ancestorBoxes)
			}
		}
		return
	}

	if b.Tag != "frame" && selfVisible {
		updated := time.Now().UnixMilli()

		noBorder := b.Attrs["border"] == "none"

		if gd, isGroup := awsGroups[b.Tag]; isGroup {
			// ── AWS group border ────────────────────────────────────
			rectID := fmt.Sprintf("%s-rect", b.ID)
			groupStroke := gd.StrokeColor
			if noBorder {
				groupStroke = "transparent"
			}
			*elements = append(*elements, map[string]any{
				"id": rectID, "type": "rectangle",
				"x": b.X, "y": b.Y, "width": b.W, "height": b.H,
				"angle": 0,
				"strokeColor": groupStroke, "backgroundColor": staggerBGColor(b),
				"fillStyle": "solid",
				"strokeWidth": gd.StrokeWidth, "strokeStyle": gd.StrokeStyle,
				"roughness": 0, "opacity": 100,
				"groupIds": []string{}, "roundness": nil,
				"seed": r.Intn(99999999), "version": 1,
				"versionNonce": r.Intn(99999999),
				"isDeleted": false, "boundElements": nil,
				"updated": updated, "link": nil, "locked": false,
			})

			// ── AWS group icon ──────────────────────────────────────
			textX := b.X + 4
			if gd.IconFile != "" && svgGroupDir != "" {
				iconPath := filepath.Join(svgGroupDir, gd.IconFile)
				var dataURL string
				var err error
				if fsys != nil {
					// In embedded mode, use forward slashes even on Windows.
					iconPath = svgGroupDir + "/" + gd.IconFile
					dataURL, err = svgDataURLFS(fsys, iconPath)
				} else {
					dataURL, err = svgDataURL(iconPath)
				}
				if err == nil {
					fid := svgFileID(gd.IconFile)
					*elements = append(*elements, map[string]any{
						"id": fmt.Sprintf("%s-icon", b.ID), "type": "image",
						"x": b.X, "y": b.Y,
						"width": float64(groupIconSize), "height": float64(groupIconSize),
						"fileId": fid, "status": "saved",
						"scale": []int{1, 1},
						"strokeColor": "transparent", "backgroundColor": "transparent",
						"fillStyle": "solid", "strokeWidth": 1, "strokeStyle": "solid",
						"roughness": 0, "opacity": 100, "angle": 0,
						"version": 1, "versionNonce": r.Intn(99999999),
						"isDeleted": false, "groupIds": []string{},
						"frameId": nil, "boundElements": nil,
						"updated": updated, "link": nil, "locked": false,
					})
					if _, exists := files[fid]; !exists {
						files[fid] = map[string]any{
							"mimeType": "image/svg+xml", "id": fid,
							"dataURL": dataURL,
							"created": updated, "lastRetrieved": updated,
						}
					}
					textX = b.X + float64(groupIconSize) + 4
				}
			}

			// ── AWS group label ─────────────────────────────────────
			textY := b.Y + float64(groupIconSize-groupFontSize)/2
			// groupFontFamily=2 (Helvetica 14px): ~7.5px/rune
			lblW := textWidth(b.Label, 7.5)
			*elements = append(*elements, map[string]any{
				"id": fmt.Sprintf("%s-label", b.ID), "type": "text",
				"x": textX, "y": textY,
				"width": lblW, "height": float64(groupFontSize + 4),
				"angle": 0,
				"strokeColor": gd.StrokeColor, "backgroundColor": "transparent",
				"fillStyle": "solid", "strokeWidth": 1, "strokeStyle": "solid",
				"roughness": 0, "opacity": 100,
				"groupIds": []string{}, "roundness": nil,
				"seed": r.Intn(99999999), "version": 1,
				"versionNonce": r.Intn(99999999),
				"isDeleted": false, "boundElements": nil,
				"updated": updated, "link": nil, "locked": false,
				"text": b.Label, "fontSize": groupFontSize, "fontFamily": groupFontFamily,
				"textAlign": "left", "verticalAlign": "middle",
				"containerId": nil, "originalText": b.Label, "lineHeight": 1.25,
			})
		} else if !isLayoutTag(b.Tag) {
			// ── Generic tag: rectangle + label ──────────────────────
			rectID := fmt.Sprintf("%s-rect", b.ID)
			textID := fmt.Sprintf("%s-text", b.ID)
			genStroke := "#1e1e1e"
			if noBorder {
				genStroke = "transparent"
			}
			*elements = append(*elements, map[string]any{
				"id": rectID, "type": "rectangle",
				"x": b.X, "y": b.Y, "width": b.W, "height": b.H,
				"angle": 0,
				"strokeColor": genStroke, "backgroundColor": "transparent",
				"fillStyle": "hachure", "strokeWidth": 1, "strokeStyle": "solid",
				"roughness": 0, "opacity": 100,
				"groupIds": []string{}, "roundness": map[string]any{"type": 3},
				"seed": r.Intn(99999999), "version": 1,
				"versionNonce": r.Intn(99999999),
				"isDeleted": false,
				"boundElements": []map[string]any{{"type": "text", "id": textID}},
				"updated": updated, "link": nil, "locked": false,
			})
			*elements = append(*elements, map[string]any{
				"id": textID, "type": "text",
				"x": b.X + 12, "y": b.Y + 12,
				// fontFamily=1 (Virgil 20px): ~10px/rune
				"width": textWidth(b.Label, 10.0), "height": 24,
				"angle": 0,
				"strokeColor": "#1e1e1e", "backgroundColor": "transparent",
				"fillStyle": "solid", "strokeWidth": 1, "strokeStyle": "solid",
				"roughness": 0, "opacity": 100,
				"groupIds": []string{}, "roundness": nil,
				"seed": r.Intn(99999999), "version": 1,
				"versionNonce": r.Intn(99999999),
				"isDeleted": false, "boundElements": nil,
				"updated": updated, "link": nil, "locked": false,
				"text": b.Label, "fontSize": 20, "fontFamily": 1,
				"textAlign": "left", "verticalAlign": "top",
				"containerId": rectID, "originalText": b.Label, "lineHeight": 1.2,
			})
		}
	}

	// Stagger background layers: render border + label only, skip children.
	if b.IsStaggerBg {
		return
	}
	// 非表示要素は visibleAncestor を引き継ぐ (子の item が正しい親に紐付くよう)
	nextVisible := b
	if !selfVisible {
		nextVisible = visibleAncestor
	}
	for _, c := range b.Children {
		walk(c, elements, files, svgGroupDir, catalogCSV, projectRoot, fsys, r, nextVisible, itemGroups, ancestorBoxes)
	}
}

// isLayoutTag reports whether a tag is a pure layout container
// (<row>, <col>, <container>) that should not render any visible border or label.
func isLayoutTag(tag string) bool {
	return tag == "row" || tag == "col" || tag == "container"
}

// textWidth estimates the rendered width of a string in pixels.
// charW: approximate pixel width per rune (font-specific).
func textWidth(s string, charW float64) float64 {
	return math.Ceil(float64(len([]rune(s)))*charW) + 8
}

// parseItemAlign parses an align attribute value (e.g. "top-left", "middle-center")
// into vertical ("top"/"middle"/"bottom") and horizontal ("left"/"center"/"right") parts.
// Defaults to "middle" / "center" when absent or unrecognised.
func parseItemAlign(align string) (vert, horiz string) {
	vert, horiz = "middle", "center"
	parts := strings.SplitN(strings.ToLower(strings.TrimSpace(align)), "-", 2)
	if len(parts) == 2 {
		if parts[0] == "top" || parts[0] == "middle" || parts[0] == "bottom" {
			vert = parts[0]
		}
		if parts[1] == "left" || parts[1] == "center" || parts[1] == "right" || parts[1] == "spread" {
			horiz = parts[1]
		}
	}
	return
}

// renderItemGrid lays out all items collected under the same visibleAncestor
// as a single horizontal row within the ancestor's content area.
//
// Alignment of the row within the content area is controlled by the ancestor's
// `align` attribute (e.g. "top-left", "middle-center", "bottom-right").
// Default is "middle-center".
//
// For AWS group ancestors the content area starts below the header row.
// Exception: when the natural (space-constrained, uncapped) icon size is ≤ 32 px,
// items are placed inline with the header (right of the label, centred in the
// 32 px header band) to avoid wasting the large empty space below.
func renderItemGrid(items []*layout.Box, ancestor *layout.Box, elements *[]map[string]any, files map[string]any, catalogCSV string, projectRoot string, fsys fs.FS, maxSize float64, r *rand.Rand, itemImgRects map[int][4]float64, itemLblRects map[int][4]float64, itemImgIDs map[int]string, itemLblIDs map[int]string, abbrevMap map[int]string) {
	if catalogCSV == "" || len(items) == 0 || ancestor == nil {
		return
	}
	n := float64(len(items))
	vert, horiz := parseItemAlign(ancestor.Attrs["align"])

	var areaX, areaY, areaW, areaH float64

	if _, isGroup := awsGroups[ancestor.Tag]; isGroup {
		// When a group's children are ALL items, layout.go used layoutRow (no GroupTopInset).
		// In that case we must also skip the topInset here so icons aren't pushed off-screen.
		allItemChildren := true
		for _, ch := range ancestor.Children {
			if !layout.IsItemLike(ch.Tag) {
				allItemChildren = false
				break
			}
		}

		if allItemChildren {
			// No topInset: treat like a generic container but respect side insets.
			areaX = ancestor.X + layout.GroupSideInset
			areaY = ancestor.Y + itemGap
			areaW = ancestor.W - layout.GroupSideInset*2
			areaH = ancestor.H - itemGap*2 - itemLabelH - 4
		} else {
			// Content area: below the header row.
			areaX = ancestor.X + layout.GroupSideInset
			areaY = ancestor.Y + layout.GroupTopInset + itemGap
			areaW = ancestor.W - layout.GroupSideInset*2
			areaH = ancestor.H - layout.GroupTopInset - itemGap*2 - itemLabelH - 4

			// Natural (uncapped) icon size — used only to decide inline vs below-header.
			naturalW := (areaW - itemGap*(n-1)) / n
			naturalH := areaH
			naturalSize := naturalW
			if naturalH < naturalSize {
				naturalSize = naturalH
			}

			// Inline placement when natural size ≤ 32: icons fit beside the header text.
			if naturalSize <= 32 {
				gd := awsGroups[ancestor.Tag]
				var headerTextStartX float64
				if gd.IconFile != "" {
					headerTextStartX = ancestor.X + float64(groupIconSize) + 4
				} else {
					headerTextStartX = ancestor.X + 4
				}
				headerEndX := headerTextStartX + textWidth(ancestor.Label, 7.5) + itemGap
				inlineAvailW := (ancestor.X + ancestor.W - layout.GroupSideInset) - headerEndX
				if inlineAvailW >= itemMinSize {
					inlineSizeW := (inlineAvailW - itemGap*(n-1)) / n
					inlineSize := inlineSizeW
					if float64(groupIconSize) < inlineSize {
						inlineSize = float64(groupIconSize)
					}
					if inlineSize > maxSize {
						inlineSize = maxSize
					}
					if inlineSize >= itemMinSize {
						// Vertically centre within the 32 px header band.
						iconY := ancestor.Y + math.Max(0, (float64(groupIconSize)-inlineSize)/2)
						totalW := inlineSize*n + itemGap*(n-1)
					var startX, stepX float64
					switch horiz {
					case "right":
						startX = ancestor.X + ancestor.W - layout.GroupSideInset - totalW
						stepX = inlineSize + itemGap
					case "spread":
						// space-evenly across full ancestor width.
						if n <= 1 {
							startX = ancestor.X + math.Max(0, (ancestor.W-inlineSize)/2)
							stepX = 0
						} else {
							evenGap := (ancestor.W - inlineSize*n) / (n + 1)
							if evenGap < itemGap {
								evenGap = itemGap
							}
							startX = ancestor.X + evenGap
							stepX = inlineSize + evenGap
						}
					case "center":
						startX = ancestor.X + math.Max(0, (ancestor.W-totalW)/2)
						stepX = inlineSize + itemGap
					default: // "left"
						startX = headerEndX
						stepX = inlineSize + itemGap
						}
						for i, item := range items {
							iconX := startX + float64(i)*stepX
							renderIconAt(item.ID, item.Attrs["id"], iconX, iconY, inlineSize, elements, files, catalogCSV, projectRoot, fsys, r, itemImgRects, itemLblRects, itemImgIDs, itemLblIDs, abbrevMap)
						}
						return
					}
				}
			}
		}
	} else {
		// 汎用コンテナ (frame, container, col など).
		areaX = ancestor.X + itemGap
		areaY = ancestor.Y + itemGap
		areaW = ancestor.W - itemGap*2
		areaH = ancestor.H - itemGap*2 - itemLabelH - 4
	}

	// ── Compute icon size within content area ────────────────────────────────
	iconSizeW := (areaW - itemGap*(n-1)) / n
	iconSize := iconSizeW
	if areaH < iconSize {
		iconSize = areaH
	}
	if iconSize > maxSize {
		iconSize = maxSize
	}
	if iconSize < itemMinSize {
		iconSize = itemMinSize
	}

	totalW := iconSize*n + itemGap*(n-1)

	// ── Horizontal placement ───────────────────────────────────────────────
	// "center" (default): pack icons and centre the group within the area.
	// "spread": space-evenly — equal gaps between icons and edges.
	// "left" / "right": pack icons at the respective edge.
	var startX float64
	var stepX float64
	switch horiz {
	case "left":
		startX = areaX
		stepX = iconSize + itemGap
	case "right":
		startX = areaX + areaW - totalW
		stepX = iconSize + itemGap
	case "spread":
		if n <= 1 {
			startX = areaX + math.Max(0, (areaW-iconSize)/2)
			stepX = 0
		} else {
			evenGap := (areaW - iconSize*n) / (n + 1)
			if evenGap < itemGap {
				evenGap = itemGap
			}
			startX = areaX + evenGap
			stepX = iconSize + evenGap
		}
	default: // "center" — pack and centre
		startX = areaX + math.Max(0, (areaW-totalW)/2)
		stepX = iconSize + itemGap
	}

	// ── Vertical alignment ───────────────────────────────────────────────────
	var iconY float64
	switch vert {
	case "top":
		iconY = areaY
	case "bottom":
		iconY = areaY + math.Max(0, areaH-iconSize)
	default: // "middle"
		iconY = areaY + math.Max(0, (areaH-iconSize)/2)
	}

	for i, item := range items {
		iconX := startX + float64(i)*stepX
		renderIconAt(item.ID, item.Attrs["id"], iconX, iconY, iconSize, elements, files, catalogCSV, projectRoot, fsys, r, itemImgRects, itemLblRects, itemImgIDs, itemLblIDs, abbrevMap)
	}
}

// renderIconAt draws a single service icon (image + label) at an explicit position.
// itemImgRects/itemLblRects/itemImgIDs/itemLblIDs are populated with the bounding rect
// and element ID of the image and label elements, keyed by the catalog integer ID.
func renderIconAt(boxID, idAttr string, iconX, iconY, iconSize float64, elements *[]map[string]any, files map[string]any, catalogCSV string, projectRoot string, fsys fs.FS, r *rand.Rand, itemImgRects map[int][4]float64, itemLblRects map[int][4]float64, itemImgIDs map[int]string, itemLblIDs map[int]string, abbrevMap map[int]string) {
	if catalogCSV == "" {
		return
	}
	idAttr = strings.TrimSpace(idAttr)
	if idAttr == "" {
		return
	}

	// 1:1 — id は単一の整数
	id, err := strconv.Atoi(idAttr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "WARNING: <item id=%q>: id must be a single integer: %v\n", idAttr, err)
		return
	}
	var ce repository.CatalogEntry
	if fsys != nil {
		ce, err = repository.LookupCatalogByIDFS(fsys, catalogCSV, id)
	} else {
		ce, err = repository.LookupCatalogByID(catalogCSV, id)
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "WARNING: <item id=%d>: %v\n", id, err)
		return
	}
	if ce.DataURL == "" && ce.RelPath != "" && projectRoot != "" {
		svgPath := filepath.Join(projectRoot, ce.RelPath)
		if du, err2 := svgDataURL(svgPath); err2 == nil {
			ce.DataURL = du
		} else {
			fmt.Fprintf(os.Stderr, "WARNING: <item id=%d>: cannot load SVG %s: %v\n", id, svgPath, err2)
		}
	}
	if ce.DataURL == "" {
		return
	}

	updated := time.Now().UnixMilli()
	fid := fmt.Sprintf("item-cat-%d", id)
	if _, exists := files[fid]; !exists {
		files[fid] = map[string]any{
			"mimeType": "image/svg+xml", "id": fid,
			"dataURL": ce.DataURL,
			"created": updated, "lastRetrieved": updated,
		}
	}
	seed := r.Intn(99999999)
	iconID := fmt.Sprintf("%s-item", boxID)
	// Record bounding rects and element IDs for edge-based connection arrows.
	if itemImgRects != nil {
		itemImgRects[id] = [4]float64{iconX, iconY, iconSize, iconSize}
		itemImgIDs[id] = iconID
	}
	*elements = append(*elements, map[string]any{
		"id": iconID, "type": "image",
		"x": iconX, "y": iconY,
		"width": iconSize, "height": iconSize,
		"fileId": fid, "status": "saved",
		"scale": []int{1, 1},
		"strokeColor": "transparent", "backgroundColor": repository.SVGBGColor(ce.DataURL),
		"fillStyle": "solid", "strokeWidth": 1, "strokeStyle": "solid",
		"roughness": 0, "opacity": 100, "angle": 0,
		"groupIds": []string{}, "roundness": nil,
		"seed": seed, "version": 1, "versionNonce": seed,
		"isDeleted": false, "boundElements": nil,
		"updated": updated, "link": nil, "locked": false, "frameId": nil,
	})
	var label string
	if abbrevMap != nil {
		label = abbrevMap[id]
	}
	if label == "" {
		label = entity.ItemShortName(ce.Service)
	}
	labelY := iconY + iconSize + 4
	labelX := iconX + (iconSize-itemLabelW)/2 // centre label on icon
	// Record label bounding rect for bottom-side connection binding.
	if itemLblRects != nil {
		itemLblRects[id] = [4]float64{labelX, labelY, itemLabelW, itemMaxSize + 2}
		itemLblIDs[id] = iconID + "-lbl"
	}
	textSeed := r.Intn(99999999)
	*elements = append(*elements, map[string]any{
		"id": iconID + "-lbl", "type": "text",
		"x": labelX, "y": labelY,
		"width": itemLabelW, "height": itemMaxSize + 2,
		"angle": 0,
		"strokeColor": "#1e1e1e", "backgroundColor": "transparent",
		"fillStyle": "solid", "strokeWidth": 1, "strokeStyle": "solid",
		"roughness": 0, "opacity": 100,
		"groupIds": []string{}, "roundness": nil,
		"seed": textSeed, "version": 1, "versionNonce": textSeed,
		"isDeleted": false, "boundElements": nil,
		"updated": updated, "link": nil, "locked": false, "frameId": nil,
		"text": label, "rawText": label, "originalText": label,
		"fontSize": 11, "fontFamily": 4,
		"textAlign": "center", "verticalAlign": "top",
		"containerId": nil, "lineHeight": 1.25,
	})
}

// connectionSide determines which edge exits src (srcSide) and enters dst (dstSide)
// based on the direction between their center points.
// Returns "top", "bottom", "left", or "right".
func connectionSide(srcCx, srcCy, dstCx, dstCy float64) (srcSide, dstSide string) {
	dx := dstCx - srcCx
	dy := dstCy - srcCy
	if math.Abs(dx) >= math.Abs(dy) {
		if dx >= 0 {
			return "right", "left"
		}
		return "left", "right"
	}
	if dy >= 0 {
		return "bottom", "top"
	}
	return "top", "bottom"
}

// rectEdgePoint returns the midpoint of the named edge of a rectangle.
// rect = [x, y, w, h]; side is "top", "bottom", "left", or "right".
func rectEdgePoint(rect [4]float64, side string) [2]float64 {
	x, y, w, h := rect[0], rect[1], rect[2], rect[3]
	cx := x + w/2
	cy := y + h/2
	switch side {
	case "top":
		return [2]float64{cx, y}
	case "bottom":
		return [2]float64{cx, y + h}
	case "left":
		return [2]float64{x, cy}
	default: // "right"
		return [2]float64{x + w, cy}
	}
}

// fixedPointForSide returns the normalized [x, y] fixedPoint on an element's bounding box
// that corresponds to the given side. This matches Excalidraw's binding coordinate system:
// [0,0]=top-left, [1,1]=bottom-right; each side midpoint:
//
//	top=[0.5,0], bottom=[0.5,1], left=[0,0.5], right=[1,0.5]
func fixedPointForSide(side string) [2]float64 {
	switch side {
	case "top":
		return [2]float64{0.5, 0}
	case "bottom":
		return [2]float64{0.5, 1}
	case "left":
		return [2]float64{0, 0.5}
	default: // "right"
		return [2]float64{1, 0.5}
	}
}

// renderConnections generates elbowed arrow elements for each <connection> node and
// updates the boundElements of the bound source/destination elements — required by
// Excalidraw so that the application recognises the binding relationship.
//
// src/dst are catalog integer IDs; the corresponding item rects and element IDs
// must already be populated in itemImgRects/itemLblRects/itemImgIDs/itemLblIDs by renderIconAt.
// Arrows start/end at the actual element edge; when the connection exits/enters from the
// bottom the label text element is used instead of the image element.
func renderConnections(connections []*model.Node, itemImgRects map[int][4]float64, itemLblRects map[int][4]float64, itemImgIDs map[int]string, itemLblIDs map[int]string, elements *[]map[string]any, r *rand.Rand) {
	if len(connections) == 0 {
		return
	}
	updated := time.Now().UnixMilli()

	// boundMap accumulates the arrow binding entries that must be written back
	// into each referenced element's boundElements array.
	// key = element ID, value = slice of {"type":"arrow","id":<arrowID>}
	boundMap := map[string][]map[string]any{}

	for i, conn := range connections {
		srcIDStr := strings.TrimSpace(conn.Attrs["src"])
		dstIDStr := strings.TrimSpace(conn.Attrs["dst"])
		srcID, err1 := strconv.Atoi(srcIDStr)
		dstID, err2 := strconv.Atoi(dstIDStr)
		if err1 != nil || err2 != nil {
			fmt.Fprintf(os.Stderr, "WARNING: <connection> invalid src/dst: %v %v\n", err1, err2)
			continue
		}
		srcImgRect, srcOk := itemImgRects[srcID]
		dstImgRect, dstOk := itemImgRects[dstID]
		if !srcOk {
			fmt.Fprintf(os.Stderr, "WARNING: <connection src=%d>: item not found or not rendered\n", srcID)
			continue
		}
		if !dstOk {
			fmt.Fprintf(os.Stderr, "WARNING: <connection dst=%d>: item not found or not rendered\n", dstID)
			continue
		}

		// Determine exit/entry side from image-center to image-center.
		srcCx := srcImgRect[0] + srcImgRect[2]/2
		srcCy := srcImgRect[1] + srcImgRect[3]/2
		dstCx := dstImgRect[0] + dstImgRect[2]/2
		dstCy := dstImgRect[1] + dstImgRect[3]/2
		srcSide, dstSide := connectionSide(srcCx, srcCy, dstCx, dstCy)

		// Choose element: bottom edge → label text box; other edges → image element.
		var srcElemID string
		var srcRect [4]float64
		if srcSide == "bottom" {
			if lblRect, ok := itemLblRects[srcID]; ok {
				srcRect = lblRect
				srcElemID = itemLblIDs[srcID]
			} else {
				srcRect = srcImgRect
				srcElemID = itemImgIDs[srcID]
			}
		} else {
			srcRect = srcImgRect
			srcElemID = itemImgIDs[srcID]
		}

		var dstElemID string
		var dstRect [4]float64
		if dstSide == "bottom" {
			if lblRect, ok := itemLblRects[dstID]; ok {
				dstRect = lblRect
				dstElemID = itemLblIDs[dstID]
			} else {
				dstRect = dstImgRect
				dstElemID = itemImgIDs[dstID]
			}
		} else {
			dstRect = dstImgRect
			dstElemID = itemImgIDs[dstID]
		}

		srcEdge := rectEdgePoint(srcRect, srcSide)
		dstEdge := rectEdgePoint(dstRect, dstSide)
		dx := dstEdge[0] - srcEdge[0]
		dy := dstEdge[1] - srcEdge[1]

		srcFP := fixedPointForSide(srcSide)
		dstFP := fixedPointForSide(dstSide)

		// seed は src/dst/index から決定論的に計算し、再生成しても描画ばらつきが出ないようにする。
		seed := srcID*1_000_000 + dstID*1_000 + i + 1
		connID := fmt.Sprintf("conn-%d-%d-%d", srcID, dstID, i)

		// arrowhead-size 属性: "s" / "m" / "l"。未指定時は最小 "s" を使用する。
		ahSize := strings.TrimSpace(conn.Attrs["arrowhead-size"])
		if ahSize != "s" && ahSize != "m" && ahSize != "l" {
			ahSize = "s"
		}

		*elements = append(*elements, map[string]any{
			"id": connID, "type": "arrow",
			"x": srcEdge[0], "y": srcEdge[1],
			"width": math.Abs(dx), "height": math.Abs(dy),
			"angle": 0,
			"strokeColor": "#1e1e1e", "backgroundColor": "transparent",
			"fillStyle": "solid", "strokeWidth": 1, "strokeStyle": "solid",
			"roughness": 0, "opacity": 100,
			"groupIds": []string{}, "roundness": map[string]any{"type": 2},
			"seed": seed, "version": 1, "versionNonce": seed,
			"isDeleted": false, "boundElements": nil,
			"updated": updated, "link": nil, "locked": false, "frameId": nil,
			"points":             [][]float64{{0, 0}, {dx, dy}},
			"lastCommittedPoint": nil,
			"startBinding": map[string]any{
				"elementId":  srcElemID,
				"focus":      0.0,
				"gap":        5.0,
				"fixedPoint": []float64{srcFP[0], srcFP[1]},
			},
			"endBinding": map[string]any{
				"elementId":  dstElemID,
				"focus":      0.0,
				"gap":        5.0,
				"fixedPoint": []float64{dstFP[0], dstFP[1]},
			},
			"startArrowhead":     nil,
			"endArrowhead":       "arrow",
			"endArrowheadSize":   ahSize,
			"startArrowheadSize": ahSize,
			"elbowed":            true,
		})

		// Register this arrow in boundMap for both endpoints.
		entry := map[string]any{"type": "arrow", "id": connID}
		boundMap[srcElemID] = append(boundMap[srcElemID], entry)
		boundMap[dstElemID] = append(boundMap[dstElemID], entry)
	}

	// Second pass: write back boundElements into each referenced element so that
	// Excalidraw recognises the binding relationship.
	if len(boundMap) == 0 {
		return
	}
	for idx := range *elements {
		elem := (*elements)[idx]
		id, _ := elem["id"].(string)
		if entries, ok := boundMap[id]; ok {
			// Merge with any existing bound elements (e.g. text containerId refs).
			existing, _ := elem["boundElements"].([]map[string]any)
			elem["boundElements"] = append(existing, entries...)
			(*elements)[idx] = elem
		}
	}
}
