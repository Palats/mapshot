package cmd

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"path"
	"path/filepath"

	"github.com/Palats/mapshot/factorio"
	"github.com/golang/glog"
	"github.com/spf13/cobra"
	"golang.org/x/sync/errgroup"
)

func devFactorio(ctx context.Context, fact *factorio.Factorio, checkoutDir string) error {
	tmpdir, cleanup := getWorkDir()
	defer cleanup()

	// Copy mods
	dstMods := filepath.Join(tmpdir, "mods")
	if err := fact.CopyMods(dstMods, []string{"mapshot"}); err != nil {
		return err
	}

	// Add the mod itself.
	dstMapshot := filepath.Join(dstMods, "mapshot")
	modDir := filepath.Join(checkoutDir, "mod")
	if err := os.Symlink(modDir, dstMapshot); err != nil {
		return fmt.Errorf("unable to symlink %q: %w", modDir, err)
	}
	glog.Infof("mod linked at %q", dstMapshot)

	factorioArgs := []string{
		"--disable-audio",
		"--mod-directory", dstMods,
	}

	fmt.Println("Starting Factorio...")
	if err := fact.Run(ctx, factorioArgs); err != nil {
		return fmt.Errorf("error while running Factorio: %w", err)
	}
	return nil
}

func devServe(ctx context.Context, fact *factorio.Factorio, checkoutDir string) error {
	baseDir := fact.ScriptOutput()
	fmt.Printf("Serving data from %s\n", baseDir)
	s := newServer(
		baseDir,
		http.FileServer(http.Dir(path.Join(checkoutDir, "frontend", "dist", "listing"))),
		http.FileServer(http.Dir(path.Join(checkoutDir, "frontend", "dist", "viewer"))),
	)
	go s.watch(ctx)

	addr := fmt.Sprintf(":%d", port)
	fmt.Printf("Listening on %s ...\n", addr)
	return http.ListenAndServe(addr, s)
}

var cmdDev = &cobra.Command{
	Use:   "dev",
	Short: "Run Factorio & HTTP server to work on the mod & UI code.",
	Long: `Run Factorio using the mod files in this directory.
The mod files are linked. This means that changes to the mod will be taken
into account when Factorio reads them - i.e., when loading a save.

No override file is created, beside the default one - all rendering
parameters come from the game.

Flag --factorio_verbose is forced to true.

It runs a HTTP server to serve the content in Factorio script-output. It
uses the UI code in frontend/dist/.
	`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		fact, err := factorio.New(factorioSettings)
		if err != nil {
			return err
		}
		fact.ForceVerbose()

		checkoutDir, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("unable to find current directory: %w", err)
		}

		grp, ctx := errgroup.WithContext(cmd.Context())

		if flagDevServe {
			grp.Go(func() error {
				return devServe(ctx, fact, checkoutDir)
			})
		}
		if flagDevFactorio {
			grp.Go(func() error {
				return devFactorio(ctx, fact, checkoutDir)
			})
		}
		return grp.Wait()
	},
}

var (
	flagDevServe    = true
	flagDevFactorio = true
)

func init() {
	renderFlags.Register(cmdDev.PersistentFlags(), "")
	cmdRoot.AddCommand(cmdDev)
	cmdDev.PersistentFlags().BoolVar(&flagDevFactorio, "factorio", true, "Run Factorio.")
	cmdDev.PersistentFlags().BoolVar(&flagDevServe, "serve", true, "Run HTTP server.")
}
