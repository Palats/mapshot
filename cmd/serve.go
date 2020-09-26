package cmd

import (
	"context"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"path/filepath"

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
			httpPath: "/" + filepath.ToSlash(name) + "/",
		})
		return nil
	})
	if err != nil {
		return nil, err
	}
	return shots, nil
}

func serve(ctx context.Context, factorioSettings *factorio.Settings, port int) error {
	fact, err := factorio.New(factorioSettings)
	if err != nil {
		return err
	}
	baseDir := fact.ScriptOutput()
	shots, err := findShots(baseDir)
	if err != nil {
		return err
	}

	mux := http.NewServeMux()

	data := []map[string]string{}
	for _, s := range shots {
		data = append(data, map[string]string{
			"name": s.name,
			"path": s.httpPath,
		})
		mux.Handle(s.httpPath, http.StripPrefix(s.httpPath, http.FileServer(http.Dir(s.fsPath))))
	}

	mux.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		if err := indexHTML.Execute(w, data); err != nil {
			http.Error(w, fmt.Sprintf("unable to generate file: %v", err), http.StatusInternalServerError)
			return
		}
	})

	addr := fmt.Sprintf(":%d", port)
	fmt.Printf("Serving data from %s\n", baseDir)
	fmt.Printf("Listening on %s ...\n", addr)
	return http.ListenAndServe(addr, mux)
}

var indexHTML = template.Must(template.New("name").Parse(`
<html>
<head>
	<title>Mapshot for Factorio</title>
</head>
<body>
	<ul>
	{{range .}}
	<li><a href="{{.path}}">{{.name}}</a></li>
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
		return serve(cmd.Context(), factorioSettings, port)
	},
}

var port int

func init() {
	cmdServe.PersistentFlags().IntVar(&port, "port", 8080, "Port to listen on.")
	cmdRoot.AddCommand(cmdServe)
}
