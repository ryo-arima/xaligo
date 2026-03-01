package controller

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/ryo-arima/xaligo/internal/config"
	"github.com/ryo-arima/xaligo/internal/excalidraw"
	"github.com/ryo-arima/xaligo/internal/layout"
	"github.com/ryo-arima/xaligo/internal/model"
	"github.com/ryo-arima/xaligo/internal/parser"
	"github.com/spf13/cobra"
)

func InitRenderCmd() *cobra.Command {
	var output string

	cmd := &cobra.Command{
		Use:   "render <input.xal>",
		Short: "Render xaligo DSL into Excalidraw JSON",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			input := args[0]
			if output == "" {
				output = "output.excalidraw"
			}
			return RunRender(input, output, nil)
		},
	}

	cmd.Flags().StringVarP(&output, "output", "o", "output.excalidraw", "output Excalidraw file path")
	return cmd
}

// abbrevMap is an optional catalog-ID → abbreviation override derived from services.csv.
// Pass nil to use only the built-in abbreviation table.
func RunRender(inputPath, outputPath string, abbrevMap map[int]string) error {
	f, err := os.Open(inputPath)
	if err != nil {
		return fmt.Errorf("open input file: %w", err)
	}
	defer f.Close()

	doc, err := parser.Parse(f)
	if err != nil {
		return fmt.Errorf("parse DSL: %w", err)
	}

	root, err := layout.Build(doc)
	if err != nil {
		return fmt.Errorf("build layout: %w", err)
	}

	// Extract <connection> nodes from the DSL root (they are meta-nodes, not layout boxes).
	var connections []*model.Node
	for _, child := range doc.Root.Children {
		if child.Tag == "connection" {
			connections = append(connections, child)
		}
	}

	cfg := config.New()
	svgGroupDir := filepath.Join(cfg.AssetDir_, "Architecture-Group-Icons")

	out, err := excalidraw.BuildJSON(root, svgGroupDir, cfg.SvcCatalogCSV, cfg.ProjectRoot, cfg.ItemIconSize, connections, abbrevMap)
	if err != nil {
		return fmt.Errorf("build excalidraw JSON: %w", err)
	}

	if err := os.WriteFile(outputPath, out, 0644); err != nil {
		return fmt.Errorf("write output file: %w", err)
	}

	fmt.Printf("generated: %s\n", outputPath)
	return nil
}
