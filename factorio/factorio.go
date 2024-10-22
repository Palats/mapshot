// Package factorio provides tooling to interact with Factorio files & binary.
package factorio

import (
	"bytes"
	"compress/zlib"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
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
)

// Factorio offers methods to manipulate a Factorio install.
type Factorio struct {
	datadir      string
	scriptOutput string
	binary       string
	verbose      bool
	keepRunning  bool
	extraArgs    []string
}

// New creates a new Factorio instance from the settings.
func New(s *Settings) (*Factorio, error) {
	datadir := s.DataDir()
	if datadir == "" {
		return nil, fmt.Errorf("no factorio data dir found; use --alsologtostderr for more info and --%sdatadir to specify its location", s.flagPrefix)
	}
	scriptOutput, err := s.ScriptOutput()
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
		datadir:      datadir,
		scriptOutput: scriptOutput,
		binary:       binary,
		verbose:      s.verbose,
		extraArgs:    extraArgs,
		keepRunning:  s.keepRunning,
	}, nil
}

// ForceVerbose set verbose to true.
func (f *Factorio) ForceVerbose() {
	f.verbose = true
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
	return filepath.Join(f.DataDir(), ModsDir)
}

// ScriptOutput is the place where mods can write data.
func (f *Factorio) ScriptOutput() string {
	return f.scriptOutput
}

// FindSaveFile try to find the savegame with the given name. It will look in
// current directory, in Factorio directory, with and without .zip.
func (f *Factorio) FindSaveFile(name string) (string, error) {
	candidates := []string{
		name,
		name + ".zip",
		filepath.Join(f.DataDir(), SavesDir, name+".zip"),
		filepath.Join(f.DataDir(), SavesDir, name),
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
				if runtime.GOOS == "windows" {
					// On Windows, os.Interrupt is a no-op, so be a bit more direct.
					cmd.Process.Signal(os.Kill)
				} else {
					cmd.Process.Signal(os.Interrupt)
				}
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
		src := filepath.Join(srcMods, sub.Name())
		dst := filepath.Join(dstMods, sub.Name())

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
	flagPrefix   string
	datadir      string
	scriptOutput string
	binary       string
	verbose      bool
	keepRunning  bool
	extraArgs    string
}

// Register add flags to configure how to call Factorio on the flagset.
func (s *Settings) Register(flags *pflag.FlagSet, prefix string) *Settings {
	s.flagPrefix = prefix
	flags.StringVar(&s.datadir, prefix+"datadir", "", "Path to factorio data dir. Tries default locations if empty.")
	flags.StringVar(&s.scriptOutput, prefix+"scriptoutput", "", "Path to factorio script-output dir. If unspecified, uses <datadir>/script-output.")
	flags.StringVar(&s.binary, prefix+"binary", "", "Path to factorio binary. Tries default locations if empty.")
	flags.BoolVar(&s.verbose, prefix+"verbose", false, "If true, stream Factorio stdout/stderr to the console.")
	flags.BoolVar(&s.keepRunning, prefix+"keep_running", false, "If true, wait for Factorio to exit instead of stopping it.")
	flags.StringVar(&s.extraArgs, prefix+"extra_args", "", "Extra args to give to Factorio; e.g., '--force-graphics-preset very-low'. Split on spaces.")
	return s
}

// DataDir returns the place where saves, mods and others are located.
// Returns "" if no directory is found.
func (s *Settings) DataDir() string {
	// List is in reverse order of priority - last one will be preferred.
	candidates := []string{
		`/opt/factorio`,
		`~/.factorio`,
		`~/factorio`,
		`~/Library/Application Support/factorio`,
	}
	if e := os.Getenv("APPDATA"); e != "" {
		candidates = append(candidates, filepath.Join(e, "Factorio"))
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
	if match == "" {
		glog.Infof("No Factorio data dir found")
		return ""
	}
	glog.Infof("Using Factorio data dir: %s", match)
	return match
}

// ScriptOutput returns the Factorio script-output directory.
func (s *Settings) ScriptOutput() (string, error) {
	if s.scriptOutput == "" {
		dataDir := s.DataDir()
		if dataDir == "" {
			return "", fmt.Errorf("no factorio data dir found; use --alsologtostderr for more info; use --%sscriptoutput to specify directly the script-output location", s.flagPrefix)
		}
		// Don't check extra subpath when using the default script-output
		// location - Factorio might not have created it by default, and not
		// everything requires to have it.
		return filepath.Join(dataDir, "script-output"), nil
	}

	d := s.scriptOutput
	info, err := os.Stat(d)
	if os.IsNotExist(err) {
		return "", fmt.Errorf("script-output dir %s does not exists", d)
	}
	if !info.IsDir() {
		return "", fmt.Errorf("script-output path %s is a file, not a directory", d)
	}
	return d, nil
}

// Binary returns the path to the Factorio binary.
func (s *Settings) Binary() (string, error) {
	// List is in reverse order of priority - last one will be preferred.
	candidates := []string{
		// Steam is a bit tricky to start/stop automatically, so ignore it for now.
		// `~/Library/Application Support/Steam/steamapps/common/Factorio/factorio.app/Contents`,
		`/opt/factorio/bin/x64/factorio`,
		`/usr/bin/factorio`,
		`~/factorio/bin/x64/factorio`,
		`~/.factorio/bin/x64/factorio`,
		`/Applications/factorio.app/Contents`,
	}
	// Steam is a bit tricky to start/stop automatically, so ignore it for now.
	// if e := os.Getenv("ProgramFiles(x86)"); e != "" {
	//	candidates = append(candidates, filepath.Join(e, "Steam", "steamapps", "common", "Factorio", "bin", "x64", "factorio.exe"))
	// }
	if e := os.Getenv("ProgramW6432"); e != "" {
		candidates = append(candidates, filepath.Join(e, "Factorio", "bin", "x64", "factorio.exe"))
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
		return "", fmt.Errorf("no factorio binary found; use --alsologtostderr for more info and --%sbinary to specify its location", s.flagPrefix)
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
	modListFile := filepath.Join(modsPath, "mod-list.json")
	mlist, err := LoadModList(modListFile)
	if err != nil {
		return err
	}
	mlist.Enable(modName)
	return mlist.Write(modListFile)
}

// Encode data in a format suitable for Factorio `game.decode_string`.
func Encode(data []byte) string {
	// Factorio
	var b bytes.Buffer
	encoder := base64.NewEncoder(base64.StdEncoding, &b)
	compress := zlib.NewWriter(encoder)
	compress.Write(data)
	compress.Close()
	encoder.Close()
	return b.String()
}
