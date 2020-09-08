package main

import (
	"archive/zip"
	goflag "flag"
	"fmt"
	"os"
	"path"

	"github.com/Palats/mapshot/embed"
	"github.com/golang/glog"
	"github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	flag "github.com/spf13/pflag"
)

var rootCmd = &cobra.Command{
	Use:   "mapshot",
	Short: "mapshot generates zoomable screenshot for Factorio",
}

var factorioDataDir = rootCmd.PersistentFlags().String("datadir", "", "Path to factorio data dir. Try default locations if empty.")
var factorioBinary = rootCmd.PersistentFlags().String("binary", "", "Path to factorio binary. Try default locations if empty.")

func findDataDir() string {
	if *factorioDataDir != "" {
		return *factorioDataDir
	}

	candidates := []string{
		`~/.factorio`,
		`~/factorio`,
		`~/Library/Application Support/factorio`,
		`%appdata%\Factorio`,
	}
	for _, c := range candidates {
		s, err := homedir.Expand(c)
		if err != nil {
			glog.Infof("Unable to expand %s: %v", c, err)
			continue
		}
		info, err := os.Stat(s)
		if os.IsNotExist(err) {
			glog.Infof("Path %s does not exists, skipped", s)
			continue
		}
		if !info.IsDir() {
			glog.Infof("Path %s is a file, skipped", s)
		}
		glog.Infof("Found factorio data dir: %s", s)
		return s
	}
	glog.Infof("No Factorio data dir found")
	return ""
}

func findBinary() string {
	if *factorioBinary != "" {
		return *factorioBinary
	}
	candidates := []string{
		`~/.factorio/bin/x64/factorio`,
		`~/factorio/bin/x64/factorio`,
		`~/Library/Application Support/Steam/steamapps/common/Factorio/factorio.app/Contents`,
		`/Applications/factorio.app/Contents`,
		`C:\Program Files (x86)\Steam\steamapps\common\Factorio`,
		`C:\Program Files\Factorio`,
	}
	for _, c := range candidates {
		s, err := homedir.Expand(c)
		if err != nil {
			glog.Infof("Unable to expand %s: %v", c, err)
			continue
		}
		info, err := os.Stat(s)
		if os.IsNotExist(err) {
			glog.Infof("Path %s does not exists, skipped", s)
			continue
		}
		if info.IsDir() {
			glog.Infof("Path %s is a directory, skipped", s)
			continue
		}
		glog.Infof("Factorio binary found: %s", s)
		return s
	}
	glog.Infof("No factorio binary found")
	return ""
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
	},
}

var cmdInfo = &cobra.Command{
	Use:   "info",
	Short: "Show info about what mapshot knows.",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("datadir:", findDataDir())
		fmt.Println("binary:", findBinary())
		return nil
	},
}

func main() {
	flag.CommandLine.AddGoFlagSet(goflag.CommandLine)
	flag.Parse()

	rootCmd.AddCommand(cmdPackage)
	rootCmd.AddCommand(cmdVersion)
	rootCmd.AddCommand(cmdInfo)
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

}
