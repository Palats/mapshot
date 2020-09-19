package cmd

import (
	"fmt"

	"github.com/Palats/mapshot/embed"
	"github.com/spf13/cobra"
)

var cmdVersion = &cobra.Command{
	Use:   "version",
	Short: "Show the version of the mod.",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(embed.Version)
	},
}

func init() {
	cmdRoot.AddCommand(cmdVersion)
}
