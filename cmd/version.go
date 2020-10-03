package cmd

import (
	"fmt"

	"github.com/Palats/mapshot/embed"
	"github.com/golang/glog"
	"github.com/spf13/cobra"
)

var cmdVersion = &cobra.Command{
	Use:   "version",
	Short: "Show the version of the mod.",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(embed.Version)
		glog.Infof("Version hash: %s", embed.VersionHash)
	},
}

func init() {
	cmdRoot.AddCommand(cmdVersion)
}
