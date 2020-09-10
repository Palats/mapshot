// Package factorio provides tooling to interact with Factorio files & binary.
package factorio

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"

	"github.com/golang/glog"
	"github.com/mitchellh/go-homedir"
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
	datadir string
	binary  string
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
	return &Factorio{
		datadir: datadir,
		binary:  binary,
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

// SaveFile gives the full path of the named savegame.
// Parameter is the name of the game save, without the .zip.
func (f *Factorio) SaveFile(name string) string {
	return path.Join(f.DataDir(), SavesDir, name+".zip")
}

// Run factorio.
func (f *Factorio) Run(ctx context.Context, args []string) error {
	glog.Infof("Running factorio with args: %v", args)
	cmd := exec.CommandContext(ctx, f.binary, args...)
	err := cmd.Run()
	glog.Infof("Factorio returned: %v", err)
	return err
}

// Settings for creating a Factorio helper instance.
type Settings struct {
	datadir string
	binary  string
}

// RegisterFlags registers a series of flags to configure Factorio (e.g., location).
func RegisterFlags(flags *flag.FlagSet, prefix string) *Settings {
	s := &Settings{}
	flags.StringVar(&s.datadir, prefix+"datadir", "", "Path to factorio data dir. Tries default locations if empty.")
	flags.StringVar(&s.binary, "binary", "", "Path to factorio binary. Tries default locations if empty.")

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
