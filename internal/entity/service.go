package entity

import "strings"

// ServiceEntry represents a single AWS service in a list or catalog.
type ServiceEntry struct {
	CatalogID    int
	OfficialName string
	Abbreviation string
	Summary      string
	Usage        string
	Note         string
}

// ShortLabel returns the abbreviation if set; otherwise strips "Amazon " / "AWS "
// from the official name to produce a compact display label.
func (s ServiceEntry) ShortLabel() string {
	if s.Abbreviation != "" {
		return s.Abbreviation
	}
	for _, pfx := range []string{"Amazon ", "AWS "} {
		if strings.HasPrefix(s.OfficialName, pfx) {
			return s.OfficialName[len(pfx):]
		}
	}
	return s.OfficialName
}
