package cmd

import (
	"archive/zip"
	"fmt"
	"os"
	"path/filepath"

	"github.com/Palats/mapshot/embed"
	"github.com/spf13/cobra"
)

func genPackage() error {
	name := fmt.Sprintf("mapshot_%s", embed.Version)
	zipfilename := name + ".zip"
	zipfile, err := os.Create(zipfilename)
	if err != nil {
		return fmt.Errorf("unable to open file %s for creation: %w", zipfilename, err)
	}
	defer zipfile.Close()

	w := zip.NewWriter(zipfile)
	for filename, content := range embed.ModFiles {
		f, err := w.Create(filepath.Join(name, filename))
		if err != nil {
			return fmt.Errorf("unable to add %q to zip file: %w", filename, err)
		}
		if _, err = f.Write([]byte(content)); err != nil {
			return fmt.Errorf("unable to write %q to zip file: %w", filename, err)
		}
	}
	if err := w.Close(); err != nil {
		return fmt.Errorf("unable to close zipfile %s: %w", zipfilename, err)
	}
	return nil
}

var cmdPackage = &cobra.Command{
	Use:   "package",
	Short: "Generates the zip file of the Factorio mod.",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		return genPackage()
	},
}

func init() {
	cmdRoot.AddCommand(cmdPackage)
}
