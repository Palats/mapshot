package cmd

import (
	"context"
	"fmt"
	"net/http"

	"github.com/Palats/mapshot/factorio"
	"github.com/spf13/cobra"
)

func serve(ctx context.Context, factorioSettings *factorio.Settings, port int) error {
	fact, err := factorio.New(factorioSettings)
	if err != nil {
		return err
	}
	baseDir := fact.ScriptOutput()
	addr := fmt.Sprintf(":%d", port)
	fmt.Printf("Serving data from %s\n", baseDir)
	fmt.Printf("Listening on %s ...\n", addr)
	return http.ListenAndServe(addr, http.FileServer(http.Dir(baseDir)))

}

var cmdServe = &cobra.Command{
	Use:   "serve",
	Short: "Start a HTTP server giving access to mapshot generated data.",
	Long: `Start a HTTP server giving access to mapshot generated data.

It serves data from Factorio script-output directory.
	`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		return serve(cmd.Context(), factorioSettings, port)
	},
}

var port int

func init() {
	cmdServe.PersistentFlags().IntVar(&port, "port", 8080, "Port to listen on.")
	cmdRoot.AddCommand(cmdServe)
}
