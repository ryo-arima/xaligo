package excalidraw

import (
	"crypto/md5"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"time"

	"github.com/ryo-arima/xaligo/internal/layout"
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

	// Minimum dimensions below which a box cannot be rendered meaningfully.
	// A box smaller than these values will be skipped with a warning.
	minBoxWidth  = 60.0
	minBoxHeight = 48.0
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

// BuildJSON converts a Box layout tree into Excalidraw JSON.
// svgGroupDir should be the absolute path to Architecture-Group-Icons/;
// pass an empty string to skip icon embedding.
func BuildJSON(root *layout.Box, svgGroupDir string) ([]byte, error) {
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
	walk(root, &elements, files, svgGroupDir, r)

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

func walk(b *layout.Box, elements *[]map[string]any, files map[string]any, svgGroupDir string, r *rand.Rand) {
	// Skip boxes that are too small to render, including all their children.
	if b.Tag != "frame" && (b.W < minBoxWidth || b.H < minBoxHeight) {
		fmt.Fprintf(os.Stderr,
			"WARNING: skipping %q (%s) — too small to display (%.1f x %.1f, min %.0f x %.0f)\n",
			b.Label, b.Tag, b.W, b.H, minBoxWidth, minBoxHeight)
		return
	}

	if b.Tag != "frame" {
		updated := time.Now().UnixMilli()

		if gd, isGroup := awsGroups[b.Tag]; isGroup {
			// ── AWS group border ────────────────────────────────────
			rectID := fmt.Sprintf("%s-rect", b.ID)
			*elements = append(*elements, map[string]any{
				"id": rectID, "type": "rectangle",
				"x": b.X, "y": b.Y, "width": b.W, "height": b.H,
				"angle": 0,
				"strokeColor": gd.StrokeColor, "backgroundColor": "transparent",
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
				if dataURL, err := svgDataURL(iconPath); err == nil {
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
			*elements = append(*elements, map[string]any{
				"id": fmt.Sprintf("%s-label", b.ID), "type": "text",
				"x": textX, "y": textY,
				"width": b.W - (textX - b.X) - 16, "height": float64(groupFontSize + 4),
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
		} else {
			// ── Generic tag: rectangle + label ──────────────────────
			rectID := fmt.Sprintf("%s-rect", b.ID)
			textID := fmt.Sprintf("%s-text", b.ID)
			*elements = append(*elements, map[string]any{
				"id": rectID, "type": "rectangle",
				"x": b.X, "y": b.Y, "width": b.W, "height": b.H,
				"angle": 0,
				"strokeColor": "#1e1e1e", "backgroundColor": "transparent",
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
				"width": float64(len([]rune(b.Label))*8 + 20), "height": 24,
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

	for _, c := range b.Children {
		walk(c, elements, files, svgGroupDir, r)
	}
}
