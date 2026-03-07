//go:build js

// Package main is the WebAssembly entry point for xaligo.
// Build with:
//
//	GOOS=js GOARCH=wasm go build -o xaligo.wasm ./cmd/wasm
//
// The resulting xaligo.wasm exposes the following functions on the global JS object:
//
//	xaligoRender(xal: string): { result?: string; error?: string }
//	  Converts a .xal DSL string into Excalidraw JSON.
//	  Uses the embedded service-catalog.csv and Architecture-Group-Icons assets.
//
//	xaligoRenderWithServices(xal: string, servicesCsv: string): { result?: string; error?: string }
//	  Same as xaligoRender but also parses a services.csv string and adds the
//	  service legend sidebar (abbreviation overrides from the CSV are applied).
package main

import (
	"fmt"
	"strings"
	"syscall/js"

	awsassets "github.com/ryo-arima/xaligo/etc/resources/aws"
	"github.com/ryo-arima/xaligo/internal/excalidraw"
	"github.com/ryo-arima/xaligo/internal/layout"
	"github.com/ryo-arima/xaligo/internal/model"
	"github.com/ryo-arima/xaligo/internal/parser"
	"github.com/ryo-arima/xaligo/internal/repository"
)

func main() {
	js.Global().Set("xaligoRender", js.FuncOf(jsRender))
	js.Global().Set("xaligoRenderWithServices", js.FuncOf(jsRenderWithServices))

	// Keep the WASM module alive until the page unloads.
	<-make(chan struct{})
}

// jsResult returns { result, error } objects back to JavaScript.
func jsResult(result string, err error) any {
	if err != nil {
		return map[string]any{"error": err.Error()}
	}
	return map[string]any{"result": result}
}

// renderXAL is the core conversion logic shared by both exported functions.
func renderXAL(xalSrc string, abbrevMap map[int]string) (string, error) {
	doc, err := parser.Parse(strings.NewReader(xalSrc))
	if err != nil {
		return "", fmt.Errorf("parse DSL: %w", err)
	}

	root, err := layout.Build(doc)
	if err != nil {
		return "", fmt.Errorf("build layout: %w", err)
	}

	var connections []*model.Node
	for _, child := range doc.Root.Children {
		if child.Tag == "connection" {
			connections = append(connections, child)
		}
	}

	out, err := excalidraw.BuildJSONWithFS(
		root,
		awsassets.Assets,
		awsassets.CatalogCSV,
		awsassets.GroupIconsDir,
		48.0,
		connections,
		abbrevMap,
	)
	if err != nil {
		return "", fmt.Errorf("build excalidraw: %w", err)
	}
	return string(out), nil
}

// jsRender handles xaligoRender(xal).
func jsRender(_ js.Value, args []js.Value) any {
	if len(args) < 1 {
		return jsResult("", fmt.Errorf("xaligoRender: expected 1 argument (xal string)"))
	}
	result, err := renderXAL(args[0].String(), nil)
	return jsResult(result, err)
}

// jsRenderWithServices handles xaligoRenderWithServices(xal, servicesCsv).
// servicesCsv is the text content of a services.csv file (same format used by
// the --services flag of the CLI command `xaligo generate excalidraw`).
func jsRenderWithServices(_ js.Value, args []js.Value) any {
	if len(args) < 2 {
		return jsResult("", fmt.Errorf("xaligoRenderWithServices: expected 2 arguments (xal, servicesCsv)"))
	}
	xal := args[0].String()
	csvContent := args[1].String()

	abbrevMap, err := parseServicesCsv(csvContent)
	if err != nil {
		return jsResult("", fmt.Errorf("parse servicesCsv: %w", err))
	}

	result, err := renderXAL(xal, abbrevMap)
	return jsResult(result, err)
}

// parseServicesCsv parses the in-memory content of a services.csv into a
// catalog-ID → abbreviation map (same format as repository.ReadServiceList).
func parseServicesCsv(content string) (map[int]string, error) {
	entries, err := repository.ReadServiceListFromReader(strings.NewReader(content))
	if err != nil {
		return nil, err
	}
	m := make(map[int]string, len(entries))
	for _, e := range entries {
		if e.CatalogID > 0 && e.Abbreviation != "" {
			m[e.CatalogID] = e.Abbreviation
		}
	}
	return m, nil
}
