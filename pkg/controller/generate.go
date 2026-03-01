package controller

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

var paperSizes = map[string][2]int{
	"A5":      {559, 794},
	"A4":      {794, 1122},
	"A3":      {1122, 1587},
	"A2":      {1587, 2245},
	"A1":      {2245, 3179},
	"Letter":  {816, 1056},
	"Legal":   {816, 1344},
	"Tabloid": {1056, 1632},
}

// InitGenerateCmd returns the `xaligo generate` parent command with subcommands:
//   - xaligo generate xal        … generate an AWS hierarchy .xal
//   - xaligo generate excalidraw … render a .xal into .excalidraw
func InitGenerateCmd() *cobra.Command {
	parent := &cobra.Command{
		Use:   "generate",
		Short: "Generate AWS infrastructure files",
	}
	parent.AddCommand(initGenerateXalCmd())
	parent.AddCommand(initGenerateExcalidrawCmd())
	return parent
}

// ── xaligo generate xal ──────────────────────────────────────────────────────

func initGenerateXalCmd() *cobra.Command {
	var (
		nClouds     int
		nAccounts   int
		nRegions    int
		nAZs        int
		azLayout    string
		nSubnets    int
		spacingMode string
		startMode   string
		paper       string
		orientation string
		output      string
	)

	cmd := &cobra.Command{
		Use:   "xal",
		Short: "Generate a .xal file with AWS infrastructure hierarchy",
		Long: `Generate a .xal file for a standard AWS architecture layout.

The hierarchy is: Cloud > Account > Region > VPC > AZ > Subnet.
Subnets alternate between public and private (1st=public, 2nd=private, ...).

Parameters correspond to options in generate_aws_frames.py:
  --az-layout   grid | staggered
  --spacing     vertical | horizontal | both
  --start       top (vertical stack) | left (horizontal side-by-side)
  --paper       A5 | A4 | A3 | A2 | A1 | Letter | Legal | Tabloid
  --orientation portrait | landscape`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return RunGenerate(
				nClouds, nAccounts, nRegions, nAZs,
				azLayout, nSubnets, spacingMode, startMode,
				paper, orientation, output,
			)
		},
	}

	cmd.Flags().IntVar(&nClouds, "clouds", 0, "number of AWS clouds (1–2)")
	cmd.Flags().IntVar(&nAccounts, "accounts", 0, "number of AWS accounts per cloud (1–3)")
	cmd.Flags().IntVar(&nRegions, "regions", 0, "number of regions per account (1–2)")
	cmd.Flags().IntVar(&nAZs, "azs", 0, "number of availability zones per region (1–3)")
	cmd.Flags().StringVar(&azLayout, "az-layout", "", "AZ layout: grid or staggered")
	cmd.Flags().IntVar(&nSubnets, "subnets", 0, "number of subnets per AZ (2–4)")
	cmd.Flags().StringVar(&spacingMode, "spacing", "", "spacing mode: vertical | horizontal | both")
	cmd.Flags().StringVar(&startMode, "start", "", "start mode: top (vertical) | left (horizontal)")
	cmd.Flags().StringVar(&paper, "paper", "", "paper size: A5 A4 A3 A2 A1 Letter Legal Tabloid")
	cmd.Flags().StringVar(&orientation, "orientation", "", "page orientation: portrait | landscape")
	cmd.Flags().StringVarP(&output, "output", "o", "", "output .xal file path")

	_ = cmd.MarkFlagRequired("clouds")
	_ = cmd.MarkFlagRequired("accounts")
	_ = cmd.MarkFlagRequired("regions")
	_ = cmd.MarkFlagRequired("azs")
	_ = cmd.MarkFlagRequired("az-layout")
	_ = cmd.MarkFlagRequired("subnets")
	_ = cmd.MarkFlagRequired("spacing")
	_ = cmd.MarkFlagRequired("start")
	_ = cmd.MarkFlagRequired("paper")
	_ = cmd.MarkFlagRequired("orientation")
	_ = cmd.MarkFlagRequired("output")

	return cmd
}

// ── xaligo generate excalidraw ───────────────────────────────────────────────

func initGenerateExcalidrawCmd() *cobra.Command {
	var (
		xalPath      string
		output       string
		servicesFile string
	)

	cmd := &cobra.Command{
		Use:   "excalidraw",
		Short: "Render a .xal file into a .excalidraw file",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := RunRender(xalPath, output); err != nil {
				return err
			}
			return RunAddServiceBatch(output, servicesFile)
		},
	}

	cmd.Flags().StringVar(&xalPath, "xal", "", "input .xal file path")
	cmd.Flags().StringVarP(&output, "output", "o", "", "output .excalidraw file path")
	cmd.Flags().StringVar(&servicesFile, "services", "", "services.csv listing AWS service icons to embed as legend")

	_ = cmd.MarkFlagRequired("xal")
	_ = cmd.MarkFlagRequired("output")
	_ = cmd.MarkFlagRequired("services")

	return cmd
}

// ── RunGenerate ──────────────────────────────────────────────────────────────

// RunGenerate validates parameters and writes the generated .xal to output.
func RunGenerate(
	nClouds, nAccounts, nRegions, nAZs int,
	azLayout string, nSubnets int,
	spacingMode, startMode, paper, orientation, output string,
) error {
	// ── validate ────────────────────────────────────────────────────────────
	size, ok := paperSizes[paper]
	if !ok {
		return fmt.Errorf("unknown paper size %q; valid: A5 A4 A3 A2 A1 Letter Legal Tabloid", paper)
	}
	W, H := size[0], size[1]
	if strings.EqualFold(orientation, "landscape") {
		W, H = H, W
	} else if !strings.EqualFold(orientation, "portrait") {
		return fmt.Errorf("orientation must be portrait or landscape")
	}
	if azLayout != "grid" && azLayout != "staggered" {
		return fmt.Errorf("az-layout must be grid or staggered")
	}
	if spacingMode != "vertical" && spacingMode != "horizontal" && spacingMode != "both" {
		return fmt.Errorf("spacing must be vertical, horizontal, or both")
	}
	if startMode != "top" && startMode != "left" {
		return fmt.Errorf("start must be top or left")
	}
	for _, pair := range [][2]int{{nClouds, 2}, {nAccounts, 3}, {nRegions, 2}, {nAZs, 3}, {nSubnets, 4}} {
		if pair[0] < 1 || pair[0] > pair[1] {
			return fmt.Errorf("value %d out of valid range (1–%d)", pair[0], pair[1])
		}
	}

	// ── generate ────────────────────────────────────────────────────────────
	xal := buildXAL(W, H, nClouds, nAccounts, nRegions, nAZs, azLayout, nSubnets, spacingMode, startMode)

	if err := os.WriteFile(output, []byte(xal), 0644); err != nil {
		return fmt.Errorf("write output file: %w", err)
	}
	fmt.Printf("generated: %s\n", output)
	return nil
}

// ── xal builder ─────────────────────────────────────────────────────────────

type xalBuilder struct {
	sb          strings.Builder
	startMode   string
	spacingMode string
	azLayout    string // "grid" or "staggered"
}

func (b *xalBuilder) ind(level int) string {
	return strings.Repeat("  ", level)
}

func (b *xalBuilder) group(tag, title string, level int, fn func()) {
	b.sb.WriteString(fmt.Sprintf("%s<%s title=%q>\n", b.ind(level), tag, title))
	fn()
	b.sb.WriteString(fmt.Sprintf("%s</%s>\n", b.ind(level), tag))
}

func (b *xalBuilder) groupAttrs(tag, title, extraAttrs string, level int, fn func()) {
	if extraAttrs != "" {
		b.sb.WriteString(fmt.Sprintf("%s<%s title=%q %s>\n", b.ind(level), tag, title, extraAttrs))
	} else {
		b.sb.WriteString(fmt.Sprintf("%s<%s title=%q>\n", b.ind(level), tag, title))
	}
	fn()
	b.sb.WriteString(fmt.Sprintf("%s</%s>\n", b.ind(level), tag))
}

func (b *xalBuilder) leaf(tag, title string, level int) {
	b.sb.WriteString(fmt.Sprintf("%s<%s title=%q />\n", b.ind(level), tag, title))
}

func (b *xalBuilder) spacingClass() string {
	switch b.spacingMode {
	case "vertical":
		return "pt-2 pb-2"
	case "horizontal":
		return "pl-2 pr-2"
	default:
		return "pa-2"
	}
}

func (b *xalBuilder) many(level, n int, fn func(i, level int)) {
	if b.startMode == "left" && n > 1 {
		span := 12 / n
		b.sb.WriteString(fmt.Sprintf("%s<row gap=\"16\">\n", b.ind(level)))
		for i := range n {
			b.sb.WriteString(fmt.Sprintf("%s  <col span=\"%d\" class=%q>\n", b.ind(level), span, b.spacingClass()))
			fn(i, level+2)
			b.sb.WriteString(fmt.Sprintf("%s  </col>\n", b.ind(level)))
		}
		b.sb.WriteString(fmt.Sprintf("%s</row>\n", b.ind(level)))
	} else {
		for i := range n {
			fn(i, level)
		}
	}
}

func buildXAL(W, H, nClouds, nAccounts, nRegions, nAZs int, azLayout string, nSubnets int, spacingMode, startMode string) string {
	b := &xalBuilder{startMode: startMode, spacingMode: spacingMode, azLayout: azLayout}

	b.sb.WriteString(fmt.Sprintf(
		"<!-- xaligo generate xal: clouds=%d accounts=%d regions=%d azs=%d az-layout=%s subnets=%d spacing=%s start=%s -->\n",
		nClouds, nAccounts, nRegions, nAZs, azLayout, nSubnets, spacingMode, startMode,
	))
	if azLayout == "staggered" {
		b.sb.WriteString("<!-- az-layout=staggered: AZs are rendered with depth offset in the excalidraw output -->\n")
	}
	b.sb.WriteString(fmt.Sprintf("<frame width=\"%d\" height=\"%d\" class=\"pa-4\">\n", W, H))

	b.many(1, nClouds, func(ci, level int) {
		b.group("aws-cloud", fmt.Sprintf("AWS Cloud %d", ci+1), level, func() {
			b.many(level+1, nAccounts, func(ai, level int) {
				b.group("aws-account", fmt.Sprintf("Account %d", ai+1), level, func() {
					b.many(level+1, nRegions, func(ri, level int) {
						b.group("region", fmt.Sprintf("Region %d", ri+1), level, func() {
							vpcAttr := ""
							if b.azLayout == "staggered" && nAZs >= 2 {
								vpcAttr = `layout="staggered"`
							}
							b.groupAttrs("vpc", fmt.Sprintf("VPC %d", ri+1), vpcAttr, level+1, func() {
								b.many(level+2, nAZs, func(zi, level int) {
									b.group("availability-zone", fmt.Sprintf("AZ %d", zi+1), level, func() {
										b.many(level+1, nSubnets, func(si, level int) {
											if si%2 == 0 {
												b.leaf("public-subnet", fmt.Sprintf("Public Subnet %d", si/2+1), level)
											} else {
												b.leaf("private-subnet", fmt.Sprintf("Private Subnet %d", si/2+1), level)
											}
										})
									})
								})
							})
						})
					})
				})
			})
		})
	})

	b.sb.WriteString("</frame>\n")
	return b.sb.String()
}
