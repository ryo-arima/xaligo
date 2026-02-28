package controller

import (
	"fmt"

	"github.com/spf13/cobra"
)

var version = "0.1.0"

func InitVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print xaligo version",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println(version)
		},
	}
}
