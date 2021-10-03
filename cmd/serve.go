package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"

	"github.com/Palats/mapshot/embed"
	"github.com/golang/glog"
	"github.com/spf13/cobra"
)

// shotInfo gives internal information about a single mapshot.
type shotInfo struct {
	name string
	// HTTP path were the tiles & data is served.
	path string
	// Name of the save. Always uses slashes.
	savename string
	json     *MapshotJSON
	// Filesystem path of this mapshot.
	fsPath string
}

// ShotsJSON is the data sent to the UI to build the listing.
type ShotsJSON struct {
	All []*ShotsJSONSave `json:"all"`
}

// ShotsJSONSave is part of ShotsJSON.
type ShotsJSONSave struct {
	Savename string           `json:"savename"`
	Versions []*ShotsJSONInfo `json:"versions"`
}

// ShotsJSONInfo is part of ShotsJSONSave.
type ShotsJSONInfo struct {
	Name        string `json:"name,omitempty"`
	Path        string `json:"path,omitempty"`
	TicksPlayed int64  `json:"ticks_played,omitempty"`
	Surface     string `json:"surface,omitempty"`
}

// MapshotJSON is a partial representation of the content of mapshot.json.
type MapshotJSON struct {
	// Many field omitted that are not used from go.
	Surface     string `json:"surface,omitempty"`
	TicksPlayed int64  `json:"ticks_played,omitempty"`
}

// MapshotConfigJSON is a representation of the viewer configuration.
type MapshotConfigJSON struct {
	Path string `json:"path"`
}

func findShots(baseDir string) ([]shotInfo, error) {
	realDir, err := filepath.EvalSymlinks(baseDir)
	if err != nil {
		return nil, fmt.Errorf("unable to eval symlinks for %s: %w", baseDir, err)
	}
	glog.Infof("Looking for shots in %s", realDir)
	var shots []shotInfo
	err = filepath.Walk(realDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if filepath.Base(path) != "mapshot.json" {
			return nil
		}
		glog.Infof("found mapshot.json: %s", path)
		raw, err := ioutil.ReadFile(path)
		if err != nil {
			glog.Errorf("file %s is not readable", path)
			return nil
		}

		mapshotData := &MapshotJSON{}
		if err := json.Unmarshal(raw, mapshotData); err != nil {
			glog.Errorf("file %s does not have valid JSON", path)
			return nil
		}

		shotPath := filepath.Dir(path)
		relpath, err := filepath.Rel(realDir, shotPath)
		if err != nil {
			glog.Infof("unable to get relative path of %q: %v", shotPath, err)
			return nil
		}
		savename := filepath.ToSlash(filepath.Dir(relpath))

		shots = append(shots, shotInfo{
			fsPath:   shotPath,
			name:     filepath.ToSlash(relpath),
			savename: savename,
			json:     mapshotData,
			path:     "/data/" + filepath.ToSlash(relpath) + "/",
		})
		return nil
	})
	if err != nil {
		return nil, err
	}
	return shots, nil
}

// Server implements a server presenting available mapshots and serving their
// content.
type Server struct {
	baseDir               string
	listingMux, viewerMux http.Handler

	m   sync.Mutex
	mux *http.ServeMux
}

func newServer(baseDir string, listingMux, viewerMux http.Handler) *Server {
	s := &Server{
		baseDir:    baseDir,
		listingMux: listingMux,
		viewerMux:  viewerMux,
	}
	s.updateMux()
	return s
}

// watch regularly updates the list of available maps. Current implementation is
// the dumbest possible one - it just rescan files every few seconds and
// recreate a completely new mux in that case.
func (s *Server) watch(ctx context.Context) {
	for {
		// Update list of maps regular, with some fuzzing.
		delay := time.Duration(8000+rand.Int63n(2000)) * time.Millisecond
		select {
		case <-time.After(delay):
		case <-ctx.Done():
			return
		}
		s.updateMux()
	}
}

func (s *Server) updateMux() {
	// Find all existing mapshots.
	shots, err := findShots(s.baseDir)
	if err != nil {
		shots = nil
		glog.Errorf("unable to find mapshots at %s: %v", s.baseDir, err)
	}

	// Build shots.json
	sort.Slice(shots, func(i, j int) bool {
		return shots[i].json.TicksPlayed > shots[j].json.TicksPlayed
	})
	kwShots := map[string]*ShotsJSONSave{}
	var savenames []string
	for _, shot := range shots {
		if kwShots[shot.savename] == nil {
			savenames = append(savenames, shot.savename)
			kwShots[shot.savename] = &ShotsJSONSave{
				Savename: shot.savename,
			}
		}
		kwShots[shot.savename].Versions = append(kwShots[shot.savename].Versions, &ShotsJSONInfo{
			Name:        shot.name,
			Path:        shot.path,
			Surface:     shot.json.Surface,
			TicksPlayed: shot.json.TicksPlayed,
		})
	}
	sort.Strings(savenames)

	var data ShotsJSON
	for _, savename := range savenames {
		data.All = append(data.All, kwShots[savename])
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		jsonData = nil
		glog.Errorf("unable to build shots.json: %v", err)
	}

	// Serve each shot data
	mux := http.NewServeMux()
	for _, shot := range shots {
		mux.Handle(shot.path, http.StripPrefix(shot.path, http.FileServer(http.Dir(shot.fsPath))))
	}

	// Serve pointer to latest
	for _, versions := range data.All {
		savename := versions.Savename
		if len(versions.Versions) < 1 {
			continue
		}
		latest := versions.Versions[0]

		cfg := &MapshotConfigJSON{
			Path: latest.Path,
		}
		jsonCfg, err := json.Marshal(cfg)
		if err != nil {
			glog.Errorf("unable to build mapshot config: %v", err)
		}
		mux.HandleFunc("/latest/"+savename, func(w http.ResponseWriter, req *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.Write(jsonCfg)
		})
	}

	// Serve basic site.
	mux.Handle("/", s.listingMux)
	mux.HandleFunc("/shots.json", func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write(jsonData)
	})

	// Serve map viewer.
	mux.Handle("/map/", http.StripPrefix("/map", s.viewerMux))

	s.m.Lock()
	defer s.m.Unlock()
	// Only update if reading did not fail - or if it was the first call, to
	// make sure we always have a mux.
	if shots != nil || s.mux == nil {
		s.mux = mux
	}
}

func (s *Server) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	s.m.Lock()
	mux := s.mux
	s.m.Unlock()
	mux.ServeHTTP(w, req)
}

var cmdServe = &cobra.Command{
	Use:   "serve",
	Short: "Start a HTTP server giving access to mapshot generated data.",
	Long: `Start a HTTP server giving access to mapshot generated data.

It serves data from Factorio script-output directory.
	`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		baseDir, err := factorioSettings.ScriptOutput()
		if err != nil {
			return err
		}
		fmt.Printf("Serving data from %s\n", baseDir)
		s := newServer(baseDir, builtinListingMux, builtinViewerMux)
		go s.watch(cmd.Context())

		addr := fmt.Sprintf(":%d", port)
		fmt.Printf("Listening on %s ...\n", addr)
		return http.ListenAndServe(addr, s)
	},
}

func buildMux(files map[string]string) *http.ServeMux {
	mux := http.NewServeMux()
	for fname, content := range files {
		fname := fname
		content := content
		mux.HandleFunc("/"+fname, func(w http.ResponseWriter, req *http.Request) {
			b := bytes.NewReader([]byte(content))
			http.ServeContent(w, req, fname, builtinModTime, b)
		})
		if fname == "index.html" {
			mux.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
				b := bytes.NewReader([]byte(content))
				http.ServeContent(w, req, fname, builtinModTime, b)
			})
		}
	}
	return mux
}

var port int
var builtinModTime = time.Now()
var builtinListingMux = buildMux(embed.ListingFiles)
var builtinViewerMux = buildMux(embed.ViewerFiles)

func init() {
	cmdServe.PersistentFlags().IntVar(&port, "port", 8080, "Port to listen on.")
	cmdRoot.AddCommand(cmdServe)
}
