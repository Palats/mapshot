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
	candidates := []string{
		`~/factorio`,
		`~/.factorio`,
		`~/Library/Application Support/factorio`,
		`%appdata%\Factorio`,
	}
	if s.datadir != "" {
		candidates = []string{s.datadir}

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

// Binary returns the path to the Factorio binary.
func (s *Settings) Binary() (string, error) {
	candidates := []string{
		`~/.factorio/bin/x64/factorio`,
		`~/factorio/bin/x64/factorio`,
		`~/Library/Application Support/Steam/steamapps/common/Factorio/factorio.app/Contents`,
		`/Applications/factorio.app/Contents`,
		`C:\Program Files (x86)\Steam\steamapps\common\Factorio`,
		`C:\Program Files\Factorio`,
	}
	if s.binary != "" {
		candidates = []string{s.binary}
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
