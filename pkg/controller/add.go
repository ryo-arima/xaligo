package controller

import (
	"crypto/rand"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strings"

	"github.com/ryo-arima/xaligo/internal/config"
	"github.com/ryo-arima/xaligo/internal/entity"
	"github.com/ryo-arima/xaligo/internal/repository"
	"github.com/spf13/cobra"
)

// InitAddCmd returns the 'add' parent command.
func InitAddCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add",
		Short: "Add elements to an existing .excalidraw file",
	}
	cmd.AddCommand(initAddServiceCmd())
	return cmd
}

// ─────────────────────────────────────────────────────────────────────────────
// add service

func initAddServiceCmd() *cobra.Command {
	var (
		targetFile string
		listFile   string
		category   string
		name       string
		size       int
		noLegend   bool
	)

	cmd := &cobra.Command{
		Use:   "service",
		Short: "Add AWS service icon(s) to a .excalidraw file",
		Long: `Searches Architecture-Service-Icons for the given service name(s) and appends
icon + label outside-bottom of the frame, with a legend entry outside the frame.

Legend placement:
  --list mode  : legend stacked on the RIGHT side of the frame
  --name mode  : legend stacked on the LEFT side of the frame

Icon placement: always outside-bottom of the frame, laid out left-to-right.

SVG data is read from service-catalog.csv (base64). The target .excalidraw file
is read, updated in-place, and written back.

Examples:
  xaligo add service --name "Amazon EC2" --file output/my.excalidraw
  xaligo add service --list services.csv --file output/my.excalidraw`,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg := config.New()
			if targetFile == "" {
				targetFile = filepath.Join(cfg.OutputFramesDir(), "A4-landscape.excalidraw")
			}
			isBatch := listFile != ""

			var entries []entity.ServiceEntry
			if isBatch {
				var err error
				entries, err = repository.ReadServiceList(listFile)
				if err != nil {
					return fmt.Errorf("read list %s: %w", listFile, err)
				}
			} else {
				if name == "" {
					return fmt.Errorf("--name or --list required")
				}
				entries = []entity.ServiceEntry{{OfficialName: name}}
			}

			return runAddBatch(targetFile, entries, category, size, noLegend, isBatch)
		},
	}

	cmd.Flags().StringVarP(&targetFile, "file", "f", "", "target .excalidraw file (required)")
	cmd.Flags().StringVar(&listFile, "list", "", "CSV file listing services to add (batch mode)")
	cmd.Flags().StringVar(&name, "name", "", "service name to add (single mode)")
	cmd.Flags().StringVar(&category, "category", "", "Architecture-Service-Icons category subdirectory")
	cmd.Flags().IntVar(&size, "size", 64, "icon size in pixels")
	cmd.Flags().BoolVar(&noLegend, "no-legend", false, "skip adding legend entries")
	_ = cmd.MarkFlagRequired("file")

	return cmd
}

// ─────────────────────────────────────────────────────────────────────────────
// Exported batch helper

// RunAddServiceBatch reads listFile (services.csv) and adds all listed service
// icons to targetFile with legend entries stacked on the right side of the frame.
func RunAddServiceBatch(targetFile, listFile string) error {
	entries, err := repository.ReadServiceList(listFile)
	if err != nil {
		return fmt.Errorf("read list %s: %w", listFile, err)
	}
	return runAddBatch(targetFile, entries, "", 64, false, true)
}

// runAddBatch is the shared implementation used by both the --list flag and
// RunAddServiceBatch. legendRight=true places legend entries on the right of
// the frame; false places them on the left.
func runAddBatch(targetFile string, entries []entity.ServiceEntry, category string, size int, noLegend, legendRight bool) error {
	cfg := config.New()
	scene, err := repository.ReadScene(targetFile)
	if err != nil {
		return err
	}

	fb := frameBounds(scene)
	iconSize := float64(size)
	gap := float64(16)
	lgSz := float64(32)
	lgLabelW := float64(120)

	for _, entry := range entries {
		var dataURL string
		var displayName string

		if entry.CatalogID > 0 {
			ce, cerr := repository.LookupCatalogByID(cfg.ServiceCatalogCSVPath(), entry.CatalogID)
			if cerr != nil {
				_, _ = fmt.Fprintf(os.Stderr, "warn: catalog ID %d: %v\n", entry.CatalogID, cerr)
				continue
			}
			dataURL = ce.DataURL
			displayName = ce.Service
		} else {
			svgPath, svgName, serr := findServiceIcon(cfg.AssetDir(), category, entry.OfficialName, size)
			if serr != nil {
				_, _ = fmt.Fprintf(os.Stderr, "warn: %v\n", serr)
				continue
			}
			displayName = svgName
			var derr error
			dataURL, derr = repository.SvgToDataURL(svgPath)
			if derr != nil {
				_, _ = fmt.Fprintf(os.Stderr, "warn: %v\n", derr)
				continue
			}
		}

		if entry.OfficialName != "" {
			displayName = entry.OfficialName
		}

		fileID := repository.FileID(dataURL)
		bgColor := repository.SVGBGColor(dataURL)
		if scene.Files == nil {
			scene.Files = map[string]map[string]interface{}{}
		}
		if _, exists := scene.Files[fileID]; !exists {
			scene.Files[fileID] = map[string]interface{}{
				"mimeType": "image/svg+xml",
				"id":       fileID,
				"dataURL":  dataURL,
				"created":  int64(1709000000000),
			}
		}

		// Main icon (outside-bottom of frame)
		ix, iy := nextIconPos(scene, fb, iconSize, gap)
		iconID := "svc-" + randomHex(8)
		seedVal := int(ix*100 + iy)
		iconEl := repository.MakeImage(iconID, ix, iy, iconSize, iconSize, fileID, bgColor, seedVal)
		scene.Elements = append(scene.Elements, iconEl)

		// Label below main icon
		label := entry.ShortLabel()
		if label == "" {
			label = shortServiceName(displayName)
		}
		labelID := "svc-lbl-" + randomHex(8)
		labelEl := repository.MakeText(labelID, ix, iy+iconSize+4, iconSize+24, 20, label, 12, "#000000", false, seedVal+1)
		scene.Elements = append(scene.Elements, labelEl)

		// Legend entry
		if !noLegend {
			var lgX, lgY float64
			if legendRight {
				lgX, lgY = nextLegendPosRight(scene, fb, lgSz, lgLabelW, gap)
			} else {
				lgX, lgY = nextLegendPosLeft(scene, fb, lgSz, lgLabelW, gap)
			}
			lgID := "svc-" + randomHex(8) + "-lg-ico"
			lgEl := repository.MakeImage(lgID, lgX, lgY, lgSz, lgSz, fileID, bgColor, seedVal+2)
			scene.Elements = append(scene.Elements, lgEl)

			lgLblID := "svc-lbl-" + randomHex(8) + "-lg"
			lgLblEl := repository.MakeText(lgLblID, lgX+lgSz+6, lgY+(lgSz-14)/2, lgLabelW, 20, label, 12, "#000000", false, seedVal+3)
			scene.Elements = append(scene.Elements, lgLblEl)
		}
	}

	return repository.WriteScene(scene, targetFile)
}

// ─────────────────────────────────────────────────────────────────────────────
// Layout helpers

// frameBBox holds the bounding rectangle of the frame or all elements.
type frameBBox struct{ x, y, w, h float64 }

// frameBounds finds the first "frame" element in the scene or falls back to
// the overall bounding box of all elements.
func frameBounds(scene *entity.Scene) frameBBox {
	// 1. Look for the first element whose type is "frame"
	for _, el := range scene.Elements {
		t, _ := el["type"].(string)
		if t == "frame" {
			x, _ := el["x"].(float64)
			y, _ := el["y"].(float64)
			w, _ := el["width"].(float64)
			h, _ := el["height"].(float64)
			if w > 0 && h > 0 {
				return frameBBox{x, y, w, h}
			}
		}
	}

	// 2. Fall back to overall bounding box of all elements
	if len(scene.Elements) == 0 {
		return frameBBox{0, 0, 800, 600}
	}
	minX, minY := math.MaxFloat64, math.MaxFloat64
	maxX, maxY := -math.MaxFloat64, -math.MaxFloat64
	found := false
	for _, el := range scene.Elements {
		x, ok1 := el["x"].(float64)
		y, ok2 := el["y"].(float64)
		w, ok3 := el["width"].(float64)
		h, ok4 := el["height"].(float64)
		if !ok1 || !ok2 || !ok3 || !ok4 {
			continue
		}
		found = true
		if x < minX {
			minX = x
		}
		if y < minY {
			minY = y
		}
		if x+w > maxX {
			maxX = x + w
		}
		if y+h > maxY {
			maxY = y + h
		}
	}
	if !found {
		return frameBBox{0, 0, 800, 600}
	}
	return frameBBox{minX, minY, maxX - minX, maxY - minY}
}

// nextIconPos returns the position for the next icon placed outside-bottom
// of the frame. Icons fill left-to-right first; when a row reaches the frame
// width the next icon wraps to the next row below.
func nextIconPos(scene *entity.Scene, fb frameBBox, iconSize, gap float64) (x, y float64) {
	rowStep := iconSize + 52 // icon + label text height
	startX := fb.x + gap
	startY := fb.y + fb.h + 60

	// How many columns fit within the frame width?
	maxCols := int(math.Floor((fb.w - gap) / (iconSize + gap)))
	if maxCols < 1 {
		maxCols = 1
	}

	count := 0
	for _, el := range scene.Elements {
		if !isMainServiceIcon(el) {
			continue
		}
		elY, _ := el["y"].(float64)
		if elY >= fb.y+fb.h+40 {
			count++
		}
	}

	// Row-first layout: fill right then down
	idx := count
	col := idx % maxCols
	row := idx / maxCols

	return startX + float64(col)*(iconSize+gap), startY + float64(row)*rowStep
}

// nextLegendPosRight returns the next legend position stacked on the RIGHT of the frame.
// When the legend reaches frame height, it adds a new column to the right.
func nextLegendPosRight(scene *entity.Scene, fb frameBBox, lgSz, lgLabelW, gap float64) (x, y float64) {
	baseX := fb.x + fb.w + 40
	rowStep := lgSz + gap
	rowsPerCol := int(math.Floor((fb.h + gap) / rowStep))
	if rowsPerCol < 1 {
		rowsPerCol = 1
	}

	count := 0
	for _, el := range scene.Elements {
		if !isLegendIcon(el) {
			continue
		}
		elX, _ := el["x"].(float64)
		elY, _ := el["y"].(float64)
		if elX >= fb.x+fb.w+10 && elY >= fb.y-10 && elY < fb.y+fb.h+40 {
			count++
		}
	}

	idx := count
	col := idx / rowsPerCol
	row := idx % rowsPerCol
	colStep := lgSz + 6 + lgLabelW + 24

	return baseX + float64(col)*colStep, fb.y + float64(row)*rowStep
}

// nextLegendPosLeft returns the next legend position stacked on the LEFT of the frame.
// When the legend reaches frame height, it adds a new column to the left.
func nextLegendPosLeft(scene *entity.Scene, fb frameBBox, lgSz, lgLabelW, gap float64) (x, y float64) {
	baseX := fb.x - lgSz - lgLabelW - 20
	rowStep := lgSz + gap
	rowsPerCol := int(math.Floor((fb.h + gap) / rowStep))
	if rowsPerCol < 1 {
		rowsPerCol = 1
	}

	count := 0
	for _, el := range scene.Elements {
		if !isLegendIcon(el) {
			continue
		}
		elX, _ := el["x"].(float64)
		elY, _ := el["y"].(float64)
		if elX < fb.x-5 && elY >= fb.y-10 && elY < fb.y+fb.h+40 {
			count++
		}
	}

	idx := count
	col := idx / rowsPerCol
	row := idx % rowsPerCol
	colStep := lgSz + 6 + lgLabelW + 24

	return baseX + float64(col)*colStep, fb.y + float64(row)*rowStep
}

func isMainServiceIcon(el map[string]interface{}) bool {
	t, _ := el["type"].(string)
	if t != "image" {
		return false
	}
	id, _ := el["id"].(string)
	return strings.HasPrefix(id, "svc-") && !strings.Contains(id, "-lg-")
}

func isLegendIcon(el map[string]interface{}) bool {
	t, _ := el["type"].(string)
	if t != "image" {
		return false
	}
	id, _ := el["id"].(string)
	return strings.HasPrefix(id, "svc-") && strings.Contains(id, "-lg-ico")
}

// ─────────────────────────────────────────────────────────────────────────────
// String helpers

// randomHex returns n random hex bytes as a string.
func randomHex(n int) string {
	b := make([]byte, n)
	_, _ = rand.Read(b)
	return fmt.Sprintf("%x", b)
}

// shortServiceName strips common cloud prefixes ("Amazon ", "AWS ") to produce
// a compact label suitable for narrow icon slots.
func shortServiceName(name string) string {
	for _, pfx := range []string{"Amazon ", "AWS "} {
		if strings.HasPrefix(name, pfx) {
			return name[len(pfx):]
		}
	}
	return name
}

// normalizeSvgName derives a display name from an SVG filename.
func normalizeSvgName(filename string) string {
	name := strings.TrimSuffix(filename, ".svg")
	for _, prefix := range []string{"Arch_", "Res_", "Arch-Category_"} {
		if strings.HasPrefix(name, prefix) {
			name = name[len(prefix):]
			break
		}
	}
	for _, suffix := range []string{"_64", "_48", "_32", "_16"} {
		if strings.HasSuffix(name, suffix) {
			name = name[:len(name)-len(suffix)]
			break
		}
	}
	return strings.ReplaceAll(strings.ReplaceAll(name, "-", " "), "_", " ")
}

// normalizeForMatch produces a lowercase, hyphen/underscore-stripped string
// for fuzzy matching against SVG filenames.
func normalizeForMatch(s string) string {
	s = strings.ToLower(s)
	s = strings.ReplaceAll(s, "-", " ")
	s = strings.ReplaceAll(s, "_", " ")
	return s
}

// findServiceIcon searches Architecture-Service-Icons for an SVG matching name.
func findServiceIcon(assetDir, category, name string, size int) (string, string, error) {
	archSvc := filepath.Join(assetDir, "Architecture-Service-Icons")
	lower := strings.ToLower(name)

	szDirs := []string{fmt.Sprintf("%d", size), "64", "48", "32", "16"}
	seen := map[string]bool{}
	var szOrder []string
	for _, s := range szDirs {
		if !seen[s] {
			szOrder = append(szOrder, s)
			seen[s] = true
		}
	}

	lowerNorm := normalizeForMatch(name)

	type candidate struct {
		path string
		name string
		base string
	}

	walkCat := func(catDir string) (string, string, bool) {
		var best *candidate
		for _, szDir := range szOrder {
			entries, err := filepath.Glob(filepath.Join(catDir, szDir, "*.svg"))
			if err != nil || len(entries) == 0 {
				continue
			}
			for _, p := range entries {
				base := filepath.Base(p)
				if strings.Contains(strings.ToLower(base), lower) ||
					strings.Contains(normalizeForMatch(base), lowerNorm) {
					c := candidate{p, normalizeSvgName(base), base}
					if best == nil || len(c.base) < len(best.base) {
						best = &c
					}
				}
			}
			if best != nil {
				break
			}
		}
		if best != nil {
			return best.path, best.name, true
		}
		return "", "", false
	}

	if category != "" {
		if p, dn, ok := walkCat(filepath.Join(archSvc, category)); ok {
			return p, dn, nil
		}
		return "", "", fmt.Errorf("service %q not found in category %q", name, category)
	}

	cats, err := filepath.Glob(filepath.Join(archSvc, "Arch_*"))
	if err != nil {
		return "", "", fmt.Errorf("scan service icons: %w", err)
	}
	for _, cat := range cats {
		if p, dn, ok := walkCat(cat); ok {
			return p, dn, nil
		}
	}
	return "", "", fmt.Errorf("service icon for %q not found in %s", name, archSvc)
}
