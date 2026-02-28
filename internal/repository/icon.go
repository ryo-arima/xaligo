package repository

import (
	"crypto/md5"
	"encoding/base64"
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"regexp"
	"strconv"
	"strings"
)

// SvgToDataURL reads an SVG file and returns it as a base64 data URL.
func SvgToDataURL(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("read SVG %s: %w", path, err)
	}
	encoded := base64.StdEncoding.EncodeToString(data)
	return "data:image/svg+xml;base64," + encoded, nil
}

// FileID derives a stable 16-char hex ID from a file path (MD5-based).
func FileID(name string) string {
	sum := md5.Sum([]byte(name))
	return fmt.Sprintf("%x", sum)[:16]
}

// LoadFromCSV loads the base64 data URL for svgFilename from the service catalog CSV.
// The CSV is expected to have columns: id,category,service,svg_file,rel_path,base64
func LoadFromCSV(csvPath, svgFilename string) (string, error) {
	f, err := os.Open(csvPath)
	if err != nil {
		return "", fmt.Errorf("open catalog CSV %s: %w", csvPath, err)
	}
	defer f.Close()

	r := csv.NewReader(f)
	r.Comment = '#'
	r.FieldsPerRecord = -1

	for {
		rec, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			continue
		}
		// columns: id, category, service, svg_file, rel_path, base64
		if len(rec) < 6 {
			continue
		}
		if strings.TrimSpace(rec[3]) == svgFilename {
			raw := strings.TrimSpace(rec[5])
			if raw == "" {
				return "", fmt.Errorf("empty base64 for %s", svgFilename)
			}
			if strings.HasPrefix(raw, "data:") {
				return raw, nil
			}
			return "data:image/svg+xml;base64," + raw, nil
		}
	}
	return "", fmt.Errorf("SVG %q not found in catalog CSV", svgFilename)
}

// LoadFromCSVByID loads the SVG path and data URL by catalog ID (and optionally name).
// Returns (svgRelPath, dataURL, error).
func LoadFromCSVByID(csvPath string, id int, name string) (string, string, error) {
	f, err := os.Open(csvPath)
	if err != nil {
		return "", "", fmt.Errorf("open catalog CSV %s: %w", csvPath, err)
	}
	defer f.Close()

	r := csv.NewReader(f)
	r.Comment = '#'
	r.FieldsPerRecord = -1

	for {
		rec, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			continue
		}
		if len(rec) < 6 {
			continue
		}
		rowID, err := strconv.Atoi(strings.TrimSpace(rec[0]))
		if err != nil || rowID != id {
			continue
		}
		// Optional: verify name matches service column (case-insensitive)
		if name != "" {
			svc := strings.TrimSpace(rec[2])
			if !strings.EqualFold(svc, name) &&
				!strings.Contains(strings.ToLower(svc), strings.ToLower(name)) &&
				!strings.Contains(strings.ToLower(name), strings.ToLower(svc)) {
				continue
			}
		}
		relPath := strings.TrimSpace(rec[4])
		raw := strings.TrimSpace(rec[5])
		if raw == "" {
			return relPath, "", fmt.Errorf("empty base64 for catalog ID %d", id)
		}
		var dataURL string
		if strings.HasPrefix(raw, "data:") {
			dataURL = raw
		} else {
			dataURL = "data:image/svg+xml;base64," + raw
		}
		return relPath, dataURL, nil
	}
	return "", "", fmt.Errorf("catalog ID %d not found in CSV", id)
}

// CatalogEntry holds resolved data for one row of service-catalog.csv.
type CatalogEntry struct {
	ID       int
	Category string
	Service  string
	SVGFile  string
	RelPath  string
	DataURL  string
}

// LookupCatalogByID finds the first entry with the given ID in service-catalog.csv
// and returns its data including the pre-encoded SVG data URL.
func LookupCatalogByID(csvPath string, id int) (CatalogEntry, error) {
	f, err := os.Open(csvPath)
	if err != nil {
		return CatalogEntry{}, fmt.Errorf("open catalog CSV %s: %w", csvPath, err)
	}
	defer f.Close()
	r := csv.NewReader(f)
	r.Comment = '#'
	r.FieldsPerRecord = -1
	for {
		rec, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			continue
		}
		if len(rec) < 6 {
			continue
		}
		rowID, err := strconv.Atoi(strings.TrimSpace(rec[0]))
		if err != nil || rowID != id {
			continue
		}
		raw := strings.TrimSpace(rec[5])
		var dataURL string
		if strings.HasPrefix(raw, "data:") {
			dataURL = raw
		} else if raw != "" {
			dataURL = "data:image/svg+xml;base64," + raw
		}
		return CatalogEntry{
			ID:       rowID,
			Category: strings.TrimSpace(rec[1]),
			Service:  strings.TrimSpace(rec[2]),
			SVGFile:  strings.TrimSpace(rec[3]),
			RelPath:  strings.TrimSpace(rec[4]),
			DataURL:  dataURL,
		}, nil
	}
	return CatalogEntry{}, fmt.Errorf("catalog ID %d not found", id)
}

var svgFillRe = regexp.MustCompile(`(?i)fill[=:]["']?(#[0-9a-fA-F]{3,8}|[a-zA-Z]+)`)

// SVGBGColor extracts the dominant fill colour from an SVG data URL.
// Falls back to "transparent" if none is found.
func SVGBGColor(dataURL string) string {
	var svgBytes []byte
	const prefix = "data:image/svg+xml;base64,"
	if strings.HasPrefix(dataURL, prefix) {
		b64 := dataURL[len(prefix):]
		decoded, err := base64.StdEncoding.DecodeString(b64)
		if err == nil {
			svgBytes = decoded
		}
	}
	if len(svgBytes) == 0 {
		return "transparent"
	}
	matches := svgFillRe.FindAllSubmatch(svgBytes, -1)
	for _, m := range matches {
		color := strings.ToLower(strings.Trim(string(m[1]), `"'`))
		if color != "" && color != "none" && color != "transparent" && color != "white" && color != "#ffffff" && color != "#fff" {
			return color
		}
	}
	return "transparent"
}
