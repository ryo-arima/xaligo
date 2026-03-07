// Package awsassets provides embedded AWS resource files used by xaligo.
//
// This package is imported by the WASM build (cmd/wasm) so that the service
// icon catalog and Architecture-Group-Icons SVGs are bundled directly into the
// binary without requiring access to the host filesystem.
package awsassets

import "embed"

// Assets is an embedded filesystem rooted at etc/resources/aws/ that contains:
//
//   - service-catalog.csv              — service icon catalog (IDs, names, SVG base64)
//   - svg/Architecture-Group-Icons/    — AWS group border icon SVGs
//   - svg/Architecture-Service-Icons/  — AWS service architecture icon SVGs
//   - svg/Category-Icons/              — AWS category icon SVGs
//   - svg/Resource-Icons/              — AWS resource icon SVGs
//
//go:embed service-catalog.csv svg
var Assets embed.FS

// CatalogCSV is the path of the service catalog CSV inside Assets.
const CatalogCSV = "service-catalog.csv"

// GroupIconsDir is the path of the Architecture group icons directory inside Assets.
const GroupIconsDir = "svg/Architecture-Group-Icons"
