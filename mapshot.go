package main

import (
	"archive/zip"
	"fmt"
	"os"
	"path"

	"github.com/Palats/mapshot/embed"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "mapshot",
	Short: "mapshot generates zoomable screenshot for Factorio",
}

var cmdVersion = &cobra.Command{
	Use:   "version",
	Short: "Show the version of the mod.",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(embed.Version)
	},
}

var cmdPackage = &cobra.Command{
	Use:   "package",
	Short: "Generates the zip file of the Factorio mod.",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		name := fmt.Sprintf("mapshot_%s", embed.Version)
		zipfilename := name + ".zip"
		zipfile, err := os.Create(zipfilename)
		if err != nil {
			return fmt.Errorf("unable to open file %s for creation: %w", zipfilename, err)
		}
		defer zipfile.Close()

		w := zip.NewWriter(zipfile)
		for filename, content := range embed.ModFiles {
			f, err := w.Create(path.Join(name, filename))
			if err != nil {
				return fmt.Errorf("unable to add %q to zip file: %w", err)
			}
			if _, err = f.Write([]byte(content)); err != nil {
				return fmt.Errorf("unable to write %q to zip file: %w", err)
			}
		}
		if err := w.Close(); err != nil {
			return fmt.Errorf("unable to close zipfile %s: %w", zipfilename, err)
		}
		return nil
	},
}

func main() {
	rootCmd.AddCommand(cmdPackage)
	rootCmd.AddCommand(cmdVersion)
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

}
