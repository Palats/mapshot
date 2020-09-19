package cmd

import (
	"context"

	"github.com/Palats/mapshot/factorio"
	"github.com/spf13/cobra"
)

var cmdRoot = &cobra.Command{
	Use:   "mapshot",
	Short: "mapshot generates zoomable screenshot for Factorio",
	// Do not show help if not requested - e.g., when an error is generated.
	SilenceUsage: true,
}

var factorioSettings = &factorio.Settings{}

func init() {
	factorioSettings.Register(cmdRoot.PersistentFlags(), "factorio_")
}

// Execute run the full command tree.
func Execute(ctx context.Context) error {
	return cmdRoot.ExecuteContext(ctx)
}
