package repository

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/ryo-arima/xaligo/internal/entity"
)

// ReadServiceList reads a CSV/TXT service list from the given file path.
func ReadServiceList(path string) ([]entity.ServiceEntry, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open service list %s: %w", path, err)
	}
	defer f.Close()
	entries, err := ReadServiceListFromReader(f)
	if err != nil {
		return nil, fmt.Errorf("read service list %s: %w", path, err)
	}
	return entries, nil
}

// ReadServiceListFromReader parses service list CSV content from an io.Reader.
// This is the Reader-based variant used by the WASM build to avoid file I/O.
//
// Format support:
//   - Lines beginning with '#' are comments and skipped.
//   - Single-column: service name only
//   - Two-column:    id,service_name  OR  service_name,category
//   - Three-column+: id,service_name,category,...
func ReadServiceListFromReader(r io.Reader) ([]entity.ServiceEntry, error) {
	var entries []entity.ServiceEntry
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		// Split up to 7 columns: id,正式名称,略語,サービス概要,用途,備考
		parts := strings.SplitN(line, ",", 7)
		for i := range parts {
			parts[i] = strings.TrimSpace(parts[i])
		}

		var entry entity.ServiceEntry
		switch len(parts) {
		case 1:
			entry.OfficialName = parts[0]
		case 2:
			// Could be "id,name" or "name,category"
			if id, err := strconv.Atoi(parts[0]); err == nil {
				entry.CatalogID = id
				entry.OfficialName = parts[1]
			} else {
				entry.OfficialName = parts[0]
			}
		default:
			// 3+ columns: id,正式名称,略語,...
			if id, err := strconv.Atoi(parts[0]); err == nil {
				entry.CatalogID = id
				entry.OfficialName = parts[1]
				if len(parts) >= 3 {
					entry.Abbreviation = parts[2]
				}
			} else {
				entry.OfficialName = parts[0]
				if len(parts) >= 2 {
					entry.Abbreviation = parts[1]
				}
			}
		}

		if entry.OfficialName != "" {
			entries = append(entries, entry)
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("scan service list: %w", err)
	}
	return entries, nil
}
