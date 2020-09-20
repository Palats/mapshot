package cmd

import (
	"context"
	"io/ioutil"
	"os"

	"github.com/Palats/mapshot/factorio"
	"github.com/golang/glog"
	"github.com/spf13/cobra"
)

var cmdRoot = &cobra.Command{
	Use:   "mapshot",
	Short: "mapshot generates zoomable screenshot for Factorio",
	// Do not show help if not requested - e.g., when an error is generated.
	SilenceUsage: true,
}

func getWorkDir() (string, func()) {
	if workDir != "" {
		glog.Infof("using work dir %s", workDir)
		return workDir, func() {}
	}

	tmpdir, err := ioutil.TempDir("", "mapshot")
	if err != nil {
		glog.Fatalf("unable to create temp dir: %v", err)
	}
	glog.Info("temp dir: ", tmpdir)

	cleanup := func() {
		// Remove temporary directory.
		if err := os.RemoveAll(tmpdir); err != nil {
			glog.Errorf("unable to remove temp dir %q: %v", tmpdir, err)
		} else {
			glog.Infof("temp dir %q removed", tmpdir)
		}
	}
	return tmpdir, cleanup
}

var (
	factorioSettings = &factorio.Settings{}
	workDir          string
)

func init() {
	factorioSettings.Register(cmdRoot.PersistentFlags(), "factorio_")
	cmdRoot.PersistentFlags().StringVar(&workDir, "work_dir", "", "If specified, uses this as working directory. Otherwise, creates a temporary one and delete it on exit.")
}

// Execute run the full command tree.
func Execute(ctx context.Context) error {
	return cmdRoot.ExecuteContext(ctx)
}
