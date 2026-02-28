package command

import (
	"fmt"
	"os"

	"github.com/ryo-arima/xaligo/pkg/controller"
	"github.com/spf13/cobra"
)

func NewRootCmd() *cobra.Command {
	var root = &cobra.Command{
		Use:   "xaligo",
		Short: "Vue-like DSL to Excalidraw layout generator",
		Long:  "xaligo renders a Vue-like layout DSL into an Excalidraw JSON file.",
	}

	root.AddCommand(controller.InitRenderCmd())
	root.AddCommand(controller.InitInitCmd())
	root.AddCommand(controller.InitVersionCmd())
	root.AddCommand(controller.InitAddCmd())
	root.AddCommand(controller.InitGenerateCmd())
	return root
}

func Execute() {
	if err := NewRootCmd().Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
