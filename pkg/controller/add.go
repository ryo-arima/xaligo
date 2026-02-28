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

			scene, err := repository.ReadScene(targetFile)
			if err != nil {
				return fmt.Errorf("read scene %s: %w", targetFile, err)
			}
			if scene.Files == nil {
				scene.Files = map[string]map[string]interface{}{}
			}

			lgSz := float64(cfg.Legend.IconSize) // default 32
			sz := math.Min(float64(size), lgSz)  // align to smaller of the two
			iconGap := 20.0
			lgFs := cfg.Legend.FontSize
			lgLabelW := math.Max(200.0, lgSz*2)

			for i, entry := range entries {
				if entry.OfficialName == "" {
					continue
				}

				// ── find icon & load data URL ──────────────────────────────────
				var svgPath, displayName, dataURL string
				if entry.CatalogID > 0 {
					// ID-based direct lookup in service-catalog.csv
					var idErr error
					svgPath, dataURL, idErr = repository.LoadFromCSVByID(cfg.ServiceCatalogCSVPath(), entry.CatalogID, entry.OfficialName)
					if idErr != nil {
						fmt.Fprintf(os.Stderr, "warn: %v (skipping)\n", idErr)
						continue
					}
					displayName = entry.OfficialName
				} else {
					// Filesystem search fallback
					var fsErr error
					svgPath, displayName, fsErr = findServiceIcon(cfg.AssetDir(), category, entry.OfficialName, size)
					if fsErr != nil {
						fmt.Fprintf(os.Stderr, "warn: %v (skipping)\n", fsErr)
						continue
					}
					svgFilename := filepath.Base(svgPath)
					var csvErr error
					dataURL, csvErr = repository.LoadFromCSV(cfg.ServiceCatalogCSVPath(), svgFilename)
					if csvErr != nil {
						// Fallback: load directly from SVG file
						var svgErr error
						dataURL, svgErr = repository.SvgToDataURL(svgPath)
						if svgErr != nil {
							fmt.Fprintf(os.Stderr, "warn: load icon %s: %v (skipping)\n", svgPath, svgErr)
							continue
						}
					}
				}

				iconColor := repository.SVGBGColor(dataURL)

				fid := repository.FileID(svgPath)
				scene.Files[fid] = map[string]interface{}{
					"mimeType":      "image/svg+xml",
					"id":            fid,
					"dataURL":       dataURL,
					"created":       int64(1709000000000),
					"lastRetrieved": int64(1709000000000),
				}

				eid := fmt.Sprintf("svc-%s-%d", randomHex(4), i)

				// ── auto-position: icon outside-bottom of frame ────────────────
				fb := frameBounds(scene)
				iconX, iconY := nextIconPos(scene, fb, sz, iconGap)

				scene.Elements = append(scene.Elements,
					repository.MakeImage(eid, iconX, iconY, sz, sz, fid, iconColor, 6000+i))

				// ── label below icon ───────────────────────────────────────────
				labelText := entry.ShortLabel()
				labelW := sz // same width as icon; center-aligned text fits within
				labelX := iconX
				labelY := iconY + sz + 4

				lblEl := repository.MakeText(eid+"-lbl", labelX, labelY, labelW, 14, labelText, 12, "#000000", false, 6100+i)
				lblEl["textAlign"] = "center"
				scene.Elements = append(scene.Elements, lblEl)

				// ── legend entry ───────────────────────────────────────────────
				if !noLegend {
					lgFid := repository.FileID(svgPath + "-lg")
					if _, exists := scene.Files[lgFid]; !exists {
						scene.Files[lgFid] = map[string]interface{}{
							"mimeType":      "image/svg+xml",
							"id":            lgFid,
							"dataURL":       dataURL,
							"created":       int64(1709000000000),
							"lastRetrieved": int64(1709000000000),
						}
					}

					var lgX, lgY float64
					if isBatch {
						lgX, lgY = nextLegendPosRight(scene, fb, lgSz, lgLabelW, 4)
					} else {
						lgX, lgY = nextLegendPosLeft(scene, fb, lgSz, lgLabelW, 4)
					}

					scene.Elements = append(scene.Elements,
						repository.MakeImage(eid+"-lg-ico", lgX, lgY, lgSz, lgSz, lgFid, iconColor, 6201+i))

					lgLblX := lgX + lgSz + 6
					lgLblY := lgY + (lgSz-float64(lgFs))/2
					scene.Elements = append(scene.Elements,
					repository.MakeText(eid+"-lg-lbl", lgLblX, lgLblY, lgLabelW, float64(lgFs), displayName, lgFs, "#000000", false, 6200+i))
				}

				fmt.Printf("added: %s\n", displayName)
			}

			if err := repository.WriteScene(scene, targetFile); err != nil {
				return err
			}
			fmt.Printf("written: %s\n", targetFile)
			return nil
		},
	}

	cmd.Flags().StringVarP(&targetFile, "file", "f", "", "target .excalidraw file (default: output/aws-frames/A4-landscape.excalidraw)")
	cmd.Flags().StringVarP(&listFile, "list", "l", "", "path to a CSV/TXT service list (batch mode)")
	cmd.Flags().StringVar(&category, "category", "", "service icon category, e.g. Arch_Compute (optional, speeds up search)")
	cmd.Flags().StringVarP(&name, "name", "n", "", "service name to search for (single add mode)")
	cmd.Flags().IntVar(&size, "size", 64, "icon size in pixels (16 | 32 | 48 | 64)")
	cmd.Flags().BoolVar(&noLegend, "no-legend", false, "omit the legend entry")
	return cmd
}

// ─────────────────────────────────────────────────────────────────────────────
// Layout helpers

// frameBBox is the bounding box of the first frame in the scene.
type frameBBox struct{ x, y, w, h float64 }

// frameBounds returns the bounding box of the first "frame" element,
// or falls back to the overall element bounding box.
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
