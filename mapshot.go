package main

import (
	"archive/zip"
	"context"
	"encoding/json"
	"errors"
	goflag "flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"strings"
	"time"

	"github.com/Palats/mapshot/embed"
	"github.com/golang/glog"
	"github.com/mitchellh/go-homedir"
	"github.com/otiai10/copy"
	"github.com/spf13/cobra"
	flag "github.com/spf13/pflag"
)

var rootCmd = &cobra.Command{
	Use:   "mapshot",
	Short: "mapshot generates zoomable screenshot for Factorio",
	// Do not show help if not requested - e.g., when an error is generated.
	SilenceUsage: true,
}

var factorioDataDir = rootCmd.PersistentFlags().String("datadir", "", "Path to factorio data dir. Try default locations if empty.")
var factorioBinary = rootCmd.PersistentFlags().String("binary", "", "Path to factorio binary. Try default locations if empty.")

func findDataDir() (string, error) {
	candidates := []string{
		`~/factorio`,
		`~/.factorio`,
		`~/Library/Application Support/factorio`,
		`%appdata%\Factorio`,
	}
	if *factorioDataDir != "" {
		candidates = []string{*factorioDataDir}

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
		return s, nil
	}
	glog.Infof("No Factorio data dir found")
	return "", errors.New("no factorio data dir found; use --alsologtostderr for more info and --datadir to specify its location")
}

func findBinary() (string, error) {
	candidates := []string{
		`~/.factorio/bin/x64/factorio`,
		`~/factorio/bin/x64/factorio`,
		`~/Library/Application Support/Steam/steamapps/common/Factorio/factorio.app/Contents`,
		`/Applications/factorio.app/Contents`,
		`C:\Program Files (x86)\Steam\steamapps\common\Factorio`,
		`C:\Program Files\Factorio`,
	}
	if *factorioBinary != "" {
		candidates = []string{*factorioBinary}
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
		return s, nil
	}
	glog.Infof("No factorio binary found")
	return "", errors.New("no factorio binary found; use --alsologtostderr for more info and --binary to specify its location")
}

// ModList represents the content of `mod-list.json` file in Factorio.
type ModList struct {
	Mods []ModListEntry `json:"mods"`
}

// ModListEntry is a single mod entry in the `mod-list.json` file.
type ModListEntry struct {
	Name    string `json:"name"`
	Enabled bool   `json:"enabled"`
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
		datadir, _ := findDataDir()
		fmt.Println("datadir:", datadir)
		binary, _ := findBinary()
		fmt.Println("binary:", binary)
		return nil
	},
}

var cmdRender = &cobra.Command{
	Use:   "render",
	Short: "Create a screenshot from a save.",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		fmt.Printf("Generating screenshot for save %q\n", name)

		datadir, err := findDataDir()
		if err != nil {
			return err
		}
		binary, err := findBinary()
		if err != nil {
			return err
		}

		tmpdir, err := ioutil.TempDir("", "mapshot")
		if err != nil {
			return fmt.Errorf("unable to create temp dir: %w", err)
		}
		glog.Info("temp dir: ", tmpdir)

		// Copy game save
		srcSavegame := path.Join(datadir, "saves", name+".zip")
		dstSavegame := path.Join(tmpdir, name+".zip")
		if err := copy.Copy(srcSavegame, dstSavegame); err != nil {
			return fmt.Errorf("unable to copy file %q: %w", srcSavegame, err)
		}
		glog.Infof("copied save from %q to %q", srcSavegame, dstSavegame)

		// Copy mods
		srcMods := path.Join(datadir, "mods")
		dstMods := path.Join(tmpdir, "mods")
		dstMapshot := path.Join(dstMods, "mapshot")
		foundModList := false

		// Create the mod directory first, in case the first file we encounter
		// is the mod-list.json (otherwise, the copy mechanism would create what
		// is needed).
		if err := os.MkdirAll(dstMapshot, 0755); err != nil {
			return fmt.Errorf("unable to create dir %q: %w", dstMapshot, err)
		}

		subs, err := ioutil.ReadDir(srcMods)
		if err != nil {
			return fmt.Errorf("unable to read directory %q: %w", srcMods, err)
		}
		for _, sub := range subs {
			src := path.Join(srcMods, sub.Name())
			dst := path.Join(dstMods, sub.Name())

			// Do not include existing mapshot plugin - it will be added afterward explictly.
			if sub.Name() == "mapshot" || strings.HasPrefix(sub.Name(), "mapshot_") {
				glog.Infof("ignoring mod file %q", src)
				continue
			}
			// Fiddle with the mod list to activate mapshot automatically.
			if sub.Name() == "mod-list.json" {
				raw, err := ioutil.ReadFile(src)
				if err != nil {
					return fmt.Errorf("unable to read %q: %w", src, err)
				}

				var data ModList
				if err := json.Unmarshal(raw, &data); err != nil {
					return fmt.Errorf("unable to decode json from %q: %w", src, err)
				}

				var mods []ModListEntry
				for _, mod := range data.Mods {
					if mod.Name == "mapshot" {
						continue
					}
					mods = append(mods, mod)
				}
				mods = append(mods, ModListEntry{
					Name:    "mapshot",
					Enabled: true,
				})
				data.Mods = mods

				raw, err = json.Marshal(data)
				if err != nil {
					return fmt.Errorf("unable to encode json: %w", err)
				}
				if err := ioutil.WriteFile(dst, raw, 0644); err != nil {
					return fmt.Errorf("unable to write file %q: %w", dst, err)
				}
				glog.Infof("created mod-list.json")
				foundModList = true
				continue
			}

			// Other mods and file, just copy.
			err = copy.Copy(src, dst, copy.Options{OnSymlink: func(string) copy.SymlinkAction { return copy.Deep }})
			if err != nil {
				return fmt.Errorf("unable to copy %q to %q: %w", src, dst, err)
			}
			glog.Infof("copied mod file %q to %q", src, dst)
		}

		if !foundModList {
			return fmt.Errorf("unable to find `mod-list.json` in %q", srcMods)
		}
		glog.Infof("copied mods from %q to %q", srcMods, dstMods)

		// Add the mod itself.
		for name, content := range embed.ModFiles {
			dst := path.Join(dstMapshot, name)
			if err := ioutil.WriteFile(dst, []byte(content), 0644); err != nil {
				return fmt.Errorf("unable to write file %q: %w", dst, err)
			}
		}
		glog.Infof("mod created at %q", dstMapshot)

		overrides := fmt.Sprintf(`
		return {
			onstartup = true,
			shotname = "%s",
			tilemin = 64,
		}
		`, name)
		overridesFilename := path.Join(dstMapshot, "overrides.lua")
		if err := ioutil.WriteFile(overridesFilename, []byte(overrides), 0644); err != nil {
			return fmt.Errorf("unable to write overrides file %q: %w", overridesFilename, err)
		}
		glog.Infof("overrides file created at %q", overridesFilename)

		// Remove done marker if still present
		doneFile := path.Join(datadir, "script-output", "mapshot-done")
		err = os.Remove(doneFile)
		glog.Infof("removed done-file %q: %v", doneFile, err)

		factorioArgs := []string{
			"--disable-audio",
			"--disable-prototype-history",
			"--load-game", dstSavegame,
			"--mod-directory", dstMods,
			"--force-graphics-preset", "very-low",
		}
		glog.Infof("Factorio args: %v", args)

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		errCh := make(chan error)
		go func() {
			fmt.Println("Starting Factorio...")
			factorioCmd := exec.CommandContext(ctx, binary, factorioArgs...)
			err := factorioCmd.Run()
			glog.Infof("Run factorio result: %v", err)
			errCh <- err
		}()

		for {
			_, err := os.Stat(doneFile)
			if err == nil {
				break
			}
			if !os.IsNotExist(err) {
				return fmt.Errorf("unable to stat file %q: %w", doneFile, err)
			}

			<-time.After(time.Second)
		}
		glog.Infof("done file %q now exists", doneFile)

		cancel()
		err = <-errCh
		if err != nil {
			if err.Error() != "signal: killed" {
				return fmt.Errorf("error while running Factorio: %w", err)
			}
		}

		if err := os.RemoveAll(tmpdir); err != nil {
			return fmt.Errorf("unable to remove temp dir %q: %w", tmpdir, err)
		}
		glog.Infof("temp dir %q removed", tmpdir)
		return nil
	},
}

func main() {
	flag.CommandLine.AddGoFlagSet(goflag.CommandLine)
	flag.Parse()

	rootCmd.AddCommand(cmdPackage)
	rootCmd.AddCommand(cmdVersion)
	rootCmd.AddCommand(cmdInfo)
	rootCmd.AddCommand(cmdRender)
	if err := rootCmd.Execute(); err != nil {
		// Root cmd already prints errors of subcommands.
		os.Exit(1)
	}

}
