// Package factorio provides tooling to interact with Factorio files & binary.
package factorio

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"strings"

	"github.com/golang/glog"
	"github.com/mitchellh/go-homedir"
	"github.com/otiai10/copy"
	"github.com/spf13/pflag"
)

const (
	// ModsDir is the datadir subdir for mods.
	ModsDir = "mods"
	// SavesDir is the datadir subdirectory for game saves.
	SavesDir = "saves"
	// ScriptOutput is the datadir subdir for where mods can write data.
	ScriptOutput = "script-output"
)

// Factorio offers methods to manipulate a Factorio install.
type Factorio struct {
	datadir     string
	binary      string
	verbose     bool
	keepRunning bool
	extraArgs   []string
}

// New creates a new Factorio instance from the settings.
func New(s *Settings) (*Factorio, error) {
	datadir, err := s.DataDir()
	if err != nil {
		return nil, err
	}
	binary, err := s.Binary()
	if err != nil {
		return nil, err
	}

	var extraArgs []string
	for _, s := range strings.Split(s.extraArgs, " ") {
		s = strings.TrimSpace(s)
		if s != "" {
			extraArgs = append(extraArgs, s)
		}
	}
	return &Factorio{
		datadir:     datadir,
		binary:      binary,
		verbose:     s.verbose,
		extraArgs:   extraArgs,
		keepRunning: s.keepRunning,
	}, nil
}

// DataDir returns the place where saves, mods and others are located.
func (f *Factorio) DataDir() string {
	return f.datadir
}

// Binary returns the path to the Factorio binary.
func (f *Factorio) Binary() string {
	return f.binary
}

// ModsDir is the directory where all the mods are located.
func (f *Factorio) ModsDir() string {
	return path.Join(f.DataDir(), ModsDir)
}

// ScriptOutput is the place where mods can write data.
func (f *Factorio) ScriptOutput() string {
	return path.Join(f.DataDir(), ScriptOutput)
}

// FindSaveFile try to find the savegame with the given name. It will look in
// current directory, in Factorio directory, with and without .zip.
func (f *Factorio) FindSaveFile(name string) (string, error) {
	candidates := []string{
		name,
		name + ".zip",
		path.Join(f.DataDir(), SavesDir, name+".zip"),
		path.Join(f.DataDir(), SavesDir, name),
	}
	for _, c := range candidates {
		_, err := os.Stat(c)
		if err == nil {
			glog.Infof("Looking for save %q; %s exists.", name, c)
			return c, nil
		}
		if !os.IsNotExist(err) {
			return "", nil
		}
		glog.Infof("Looking for save %q; %s does not exists.", name, c)
	}
	return "", os.ErrNotExist
}

// Run factorio.
func (f *Factorio) Run(ctx context.Context, args []string) error {
	args = append(append([]string{}, args...), f.extraArgs...)
	glog.Infof("Running factorio with args: %v", args)
	cmd := exec.Command(f.binary, args...)
	if f.verbose {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}
	done := make(chan struct{})
	go func() {
		select {
		case <-done:
			return
		case <-ctx.Done():
			if f.keepRunning {
				glog.Infof("interrupt requested, but keep_running specified")
			} else {
				glog.Infof("interrupt requested")
				// When killing immediately, some files will not be written, even
				// when the `done` file is already visible. Instead, ask to
				// shutdown politely.
				// Unfortunately, that will be an issue for windows.
				cmd.Process.Signal(os.Interrupt)
			}
		}
	}()
	err := cmd.Run()
	close(done)
	glog.Infof("Factorio returned: %v", err)
	return err
}

// CopyMods creates a mods directory in the given location based on the current one.
// This can serve as a base to forcefully enable a mod or similar.
func (f *Factorio) CopyMods(dstMods string, filterOut []string) error {
	srcMods := f.ModsDir()
	foundModList := false

	filtered := map[string]bool{}
	for _, f := range filterOut {
		filtered[f] = true
	}

	if err := os.MkdirAll(dstMods, 0755); err != nil {
		return fmt.Errorf("unable to create dir %q: %w", dstMods, err)
	}

	subs, err := ioutil.ReadDir(srcMods)
	if err != nil {
		return fmt.Errorf("unable to read directory %q: %w", srcMods, err)
	}
	for _, sub := range subs {
		src := path.Join(srcMods, sub.Name())
		dst := path.Join(dstMods, sub.Name())

		// Some plugins might be requested to exclude.
		modName := sub.Name()
		if idx := strings.LastIndex(modName, "_"); idx >= 0 {
			modName = modName[:idx]
		}
		glog.Infof("copying mod %s from %s to %s", modName, src, dst)

		if filtered[modName] {
			glog.Infof("ignoring mod file %q", src)
			continue
		}
		// Fiddle with the mod list to remove filtered mods.
		if sub.Name() == "mod-list.json" {
			mlist, err := LoadModList(src)
			if err != nil {
				return err
			}
			var mods []*ModListEntry
			for _, mod := range mlist.Mods {
				if !filtered[mod.Name] {
					mods = append(mods, mod)
				}
			}
			mlist.Mods = mods

			if err := mlist.Write(dst); err != nil {
				return err
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
	return nil
}

// Settings for creating a Factorio helper instance.
type Settings struct {
	datadir     string
	binary      string
	verbose     bool
	keepRunning bool
	extraArgs   string
}

// Register add flags to configure how to call Factorio on the flagset.
func (s *Settings) Register(flags *pflag.FlagSet, prefix string) *Settings {
	flags.StringVar(&s.datadir, prefix+"datadir", "", "Path to factorio data dir. Tries default locations if empty.")
	flags.StringVar(&s.binary, prefix+"binary", "", "Path to factorio binary. Tries default locations if empty.")
	flags.BoolVar(&s.verbose, prefix+"verbose", false, "If true, stream Factorio stdout/stderr to the console.")
	flags.BoolVar(&s.keepRunning, prefix+"keep_running", false, "If true, wait for Factorio to exit instead of stopping it.")
	flags.StringVar(&s.extraArgs, prefix+"extra_args", "", "Extra args to give to Factorio; e.g., '--force-graphics-preset very-low'. Split on spaces.")
	return s
}

// DataDir returns the place where saves, mods and others are located.
func (s *Settings) DataDir() (string, error) {
	// List is in reverse order of priority - last one will be preferred.
	candidates := []string{
		`~/.factorio`,
		`~/factorio`,
		`~/Library/Application Support/factorio`,
	}
	if e := os.Getenv("APPDATA"); e != "" {
		candidates = append(candidates, path.Join(e, "Factorio"))
	}

	if s.datadir != "" {
		candidates = []string{s.datadir}
	}

	match := ""
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
			continue
		}
		glog.Infof("Found factorio data dir: %s", s)
		match = s
	}
	if s == nil {
		glog.Infof("No Factorio data dir found")
		return "", errors.New("no factorio data dir found; use --alsologtostderr for more info and --datadir to specify its location")
	}
	glog.Infof("Using Factorio data dir: %s", match)
	return match, nil
}

// Binary returns the path to the Factorio binary.
func (s *Settings) Binary() (string, error) {
	// List is in reverse order of priority - last one will be preferred.
	candidates := []string{
		`~/Library/Application Support/Steam/steamapps/common/Factorio/factorio.app/Contents`,
		`~/factorio/bin/x64/factorio`,
		`~/.factorio/bin/x64/factorio`,
		`/Applications/factorio.app/Contents`,
	}
	if e := os.Getenv("ProgramFiles(x86)"); e != "" {
		candidates = append(candidates, path.Join(e, "Steam", "steamapps", "common", "Factorio", "bin", "x64", "factorio.exe"))
	}
	if e := os.Getenv("ProgramW6432"); e != "" {
		candidates = append(candidates, path.Join(e, "Factorio", "bin", "x64", "factorio.exe"))
	}

	if s.binary != "" {
		candidates = []string{s.binary}
	}
	match := ""
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
		match = s
	}
	if match == "" {
		glog.Infof("No factorio binary found")
		return "", errors.New("no factorio binary found; use --alsologtostderr for more info and --binary to specify its location")
	}
	glog.Infof("Using Factorio binary: %s", match)
	return match, nil
}

// ModList represents the content of `mod-list.json` file in Factorio.
type ModList struct {
	Mods []*ModListEntry `json:"mods"`
}

// ModListEntry is a single mod entry in the `mod-list.json` file.
type ModListEntry struct {
	Name    string `json:"name"`
	Enabled bool   `json:"enabled"`
}

// Enable marks a given mod as enabled.
func (mlist *ModList) Enable(name string) {
	for _, mod := range mlist.Mods {
		if mod.Name == name {
			mod.Enabled = true
			return
		}
	}
	mlist.Mods = append(mlist.Mods, &ModListEntry{
		Name:    name,
		Enabled: true,
	})
}

// Write writes the given filename with a serialized version of this modlist.
func (mlist *ModList) Write(filename string) error {
	raw, err := json.Marshal(mlist)
	if err != nil {
		return fmt.Errorf("unable to encode json: %w", err)
	}
	if err := ioutil.WriteFile(filename, raw, 0644); err != nil {
		return fmt.Errorf("unable to write file %q: %w", filename, err)
	}
	return nil
}

// LoadModList reads a mod-list.json file from its filename.
func LoadModList(filename string) (*ModList, error) {
	raw, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("unable to read %q: %w", filename, err)
	}

	data := &ModList{}
	if err := json.Unmarshal(raw, data); err != nil {
		return nil, fmt.Errorf("unable to decode json from %q: %w", filename, err)
	}

	return data, nil
}

// EnableMod activate the named mod in the provided mod directory.
func EnableMod(modsPath string, modName string) error {
	modListFile := path.Join(modsPath, "mod-list.json")
	mlist, err := LoadModList(modListFile)
	if err != nil {
		return err
	}
	mlist.Enable(modName)
	return mlist.Write(modListFile)
}
