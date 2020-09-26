package cmd

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"github.com/Palats/mapshot/factorio"
	"github.com/golang/glog"
	"github.com/spf13/cobra"
)

func findShots(baseDir string) ([]string, error) {
	realDir, err := filepath.EvalSymlinks(baseDir)
	if err != nil {
		return nil, fmt.Errorf("unable to eval symlinks for %s: %w", baseDir, err)
	}
	glog.Infof("Looking for shots in %s", realDir)
	var paths []string
	err = filepath.Walk(realDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if filepath.Base(path) != "mapshot.json" {
			return nil
		}
		glog.Infof("found mapshot.json: %s", path)
		shotpath := filepath.Dir(path)
		paths = append(paths, shotpath)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return paths, nil
}

func serve(ctx context.Context, factorioSettings *factorio.Settings, port int) error {
	fact, err := factorio.New(factorioSettings)
	if err != nil {
		return err
	}
	baseDir := fact.ScriptOutput()
	paths, err := findShots(baseDir)
	if err != nil {
		return err
	}
	for _, p := range paths {
		fmt.Printf("Found shot: %s\n", p)
	}
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
