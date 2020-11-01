package cmd

import (
	"context"
	"fmt"
	"html/template"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/Palats/mapshot/embed"
	"github.com/Palats/mapshot/factorio"
	"github.com/golang/glog"
	"github.com/spf13/cobra"
)

type shotInfo struct {
	name     string
	fsPath   string
	httpPath string
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
		shotPath := filepath.Dir(path)
		name, err := filepath.Rel(realDir, shotPath)
		if err != nil {
			glog.Infof("unable to get relative path of %q: %v", shotPath, err)
			return nil
		}
		shots = append(shots, shotInfo{
			fsPath:   shotPath,
			name:     name,
			httpPath: "/data/" + filepath.ToSlash(name) + "/",
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
	baseDir     string
	frontendMux http.Handler

	m   sync.Mutex
	mux *http.ServeMux
}

func newServer(baseDir string, frontendMux http.Handler) *Server {
	s := &Server{
		baseDir:     baseDir,
		frontendMux: frontendMux,
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
	shots, err := findShots(s.baseDir)
	if err != nil {
		shots = nil
		glog.Errorf("unable to find mapshots at %s: %v", s.baseDir, err)
	}
	mux := s.buildMux(shots)
	s.m.Lock()
	defer s.m.Unlock()
	// Only update if reading did not fail - or if it was the first call, to
	// make sure we always have a mux.
	if shots != nil || s.mux == nil {
		s.mux = mux
	}
}

func (s *Server) buildMux(shots []shotInfo) *http.ServeMux {
	mux := http.NewServeMux()

	data := []map[string]string{}
	for _, shot := range shots {
		data = append(data, map[string]string{
			"name": shot.name,
			"path": shot.httpPath,
		})
		mux.Handle(shot.httpPath, http.StripPrefix(shot.httpPath, http.FileServer(http.Dir(shot.fsPath))))
	}

	mux.Handle("/map/", http.StripPrefix("/map", s.frontendMux))
	mux.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		if err := indexHTML.Execute(w, data); err != nil {
			http.Error(w, fmt.Sprintf("unable to generate file: %v", err), http.StatusInternalServerError)
			return
		}
	})

	return mux
}

func (s *Server) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	s.m.Lock()
	mux := s.mux
	s.m.Unlock()
	mux.ServeHTTP(w, req)
}

var indexHTML = template.Must(template.New("name").Parse(`
<html>
<head>
	<title>Mapshot for Factorio</title>
	<link rel="icon" href="map/thumbnail.png" sizes="144x144">
</head>
<body>
	{{if not .}}
	No mapshots have been found. Create some and re-start mapshot server.
	{{end}}
	<ul>
	{{range .}}
	<li><a href="map?path={{.path}}">{{.name}}</a></li>
	{{end}}
	</ul>
</body>
`))

var cmdServe = &cobra.Command{
	Use:   "serve",
	Short: "Start a HTTP server giving access to mapshot generated data.",
	Long: `Start a HTTP server giving access to mapshot generated data.

It serves data from Factorio script-output directory.
	`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		fact, err := factorio.New(factorioSettings)
		if err != nil {
			return err
		}

		baseDir := fact.ScriptOutput()
		fmt.Printf("Serving data from %s\n", baseDir)
		s := newServer(baseDir, builtinFrontendMux)
		go s.watch(cmd.Context())

		addr := fmt.Sprintf(":%d", port)
		fmt.Printf("Listening on %s ...\n", addr)
		return http.ListenAndServe(addr, s)
	},
}

var port int
var builtinFrontendMux *http.ServeMux

func init() {
	cmdServe.PersistentFlags().IntVar(&port, "port", 8080, "Port to listen on.")
	cmdRoot.AddCommand(cmdServe)

	// Build a mux to serve frontend files from embed/ module.
	builtinFrontendMux = http.NewServeMux()
	for fname, content := range embed.FrontendFiles {
		fname := fname
		content := content
		builtinFrontendMux.HandleFunc("/"+fname, func(w http.ResponseWriter, req *http.Request) {
			w.Write([]byte(content))
		})
		if fname == "index.html" {
			builtinFrontendMux.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
				w.Write([]byte(content))
			})
		}
	}
}
