package cmd

import (
	"fmt"

	"github.com/Palats/mapshot/factorio"
	"github.com/spf13/cobra"
)

var cmdInfo = &cobra.Command{
	Use:   "info",
	Short: "Show info about what mapshot knows.",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		fact, err := factorio.New(factorioSettings)
		if err != nil {
			return err
		}
		fmt.Println("datadir:", fact.DataDir())
		fmt.Println("binary:", fact.Binary())
		return nil
	},
}

func init() {
	cmdRoot.AddCommand(cmdInfo)
}
