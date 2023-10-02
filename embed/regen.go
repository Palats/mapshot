// +build ignore

// Regenerate the mod data for embedding in Go/Lua.
// Coding style is very direct - this is a short regen script, errors are
// directly fatal.
package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"hash"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strings"

	"github.com/Palats/mapshot/factorio"
)

// textEOL contains the end-of-line chars to use. That allows to generate \r\n
// on Windows (assuming autocrlf with git).
var textEOL = "\n"

func init() {
	if runtime.GOOS == "windows" {
		textEOL = "\r\n"
	}
}

func getVersion() string {
	raw, err := ioutil.ReadFile("mod/info.json")
	if err != nil {
		log.Fatal(err)
	}
	var data map[string]interface{}
	if err := json.Unmarshal(raw, &data); err != nil {
		log.Fatal(err)
	}
	version := data["version"].(string)
	if version == "" {
		log.Fatal("Missing version info")
	}
	return version
}

func genLua(viewerFiles []*FileInfo, version, versionHash string) {
	f, err := os.Create("mod/generated.lua")
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	writeLn := func(s string) {
		if _, err := f.WriteString(s + textEOL); err != nil {
			log.Fatal(err)
		}
	}

	writeLn("-- Automatically generated, do not modify")
	writeLn("local data = {}")
	writeLn(fmt.Sprintf("data.version = %q", version))
	writeLn(fmt.Sprintf("data.version_hash = %q", versionHash))
	writeLn("data.files = {}")

	for _, fi := range viewerFiles {
		key := filepath.Base(fi.Filename)

		if fi.Binary {
			content := factorio.Encode(fi.Content)
			writeLn(fmt.Sprintf(`data.files[%q] = function() return game.decode_string(table.concat({`, key))
			begin := 0
			end := 78
			for begin < len(content) {
				if end > len(content) {
					end = len(content)
				}
				writeLn(fmt.Sprintf(`  "%s",`, content[begin:end]))
				begin = end
				end = end + 78
			}
			writeLn("})) end")
		} else {
			if strings.Contains(string(fi.Content), "]==]") {
				log.Fatal("dumb Lua encoding cannot proceed")
			}

			writeLn(fmt.Sprintf("data.files[%q] = function() return [==[", key))
			// This blindly copy the file content. As both the source and target files
			// are text file, they will preserve the end-of-lines.
			if _, err := f.Write(fi.Content); err != nil {
				log.Fatal(err)
			}
			writeLn("]==] end")
		}
		writeLn("")
	}
	writeLn("return data")
}

var filenameSpecials = regexp.MustCompile(`[^a-zA-Z]`)

func filenameToVar(fname string) string {
	s := ""
	for _, p := range filenameSpecials.Split(fname, -1) {
		if len(p) == 0 {
			continue
		}
		if p == "json" {
			p = "JSON"
		} else if p == "html" {
			p = "HTML"
		} else {
			p = strings.ToUpper(p[0:1]) + strings.ToLower(p[1:])
		}
		s += p
	}
	return "File" + s
}

func genGo(modFiles, viewerFiles, listingFiles []*FileInfo, version string, versionHash string) {
	f, err := os.Create("embed/generated.go")
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	writeLn := func(s string) {
		if _, err := f.WriteString(s + textEOL); err != nil {
			log.Fatal(err)
		}
	}

	writeLn("// Package embed is AUTOMATICALLY GENERATED, DO NOT EDIT")
	writeLn("package embed")
	writeLn("")
	writeLn("// Version of the mod")
	writeLn(fmt.Sprintf("var Version = %q", version))
	writeLn("")
	writeLn("// VersionHash is a hash of the mod content")
	writeLn(fmt.Sprintf("var VersionHash = %q", versionHash))
	writeLn("")

	var files []*FileInfo

	writeLn("// ModFiles is the list of files for the Factorio mod.")
	writeLn("var ModFiles = map[string]string{")
	for _, fi := range modFiles {
		files = append(files, fi)
		writeLn(fmt.Sprintf("\t%q: %s,", filepath.Base(fi.Filename), filenameToVar(fi.Filename)))
	}
	writeLn("}")
	writeLn("")

	writeLn("// ViewerFiles is the files for the UI to navigate a single mapshot (map view).")
	writeLn("var ViewerFiles = map[string]string{")
	for _, fi := range viewerFiles {
		files = append(files, fi)
		writeLn(fmt.Sprintf("\t%q: %s,", filepath.Base(fi.Filename), filenameToVar(fi.Filename)))
	}
	writeLn("}")
	writeLn("")

	writeLn("// ListingFiles is the files for the UI to navigate the list of mapshots.")
	writeLn("var ListingFiles = map[string]string{")
	for _, fi := range listingFiles {
		files = append(files, fi)
		writeLn(fmt.Sprintf("\t%q: %s,", filepath.Base(fi.Filename), filenameToVar(fi.Filename)))
	}
	writeLn("}")
	writeLn("")

	seen := map[*FileInfo]bool{}
	for _, fi := range sortFiles(files) {
		if seen[fi] {
			continue
		}
		seen[fi] = true
		varName := filenameToVar(fi.Filename)

		writeLn(fmt.Sprintf("// %s is file %q", varName, filepath.ToSlash(fi.Filename)))
		writeLn(fmt.Sprintf("var %s =", varName))

		for _, line := range strings.SplitAfter(string(fi.Content), "\n") {
			for len(line) > 120 {
				writeLn(fmt.Sprintf("\t%q + // cont.", line[:120]))
				line = line[120:]
			}
			writeLn(fmt.Sprintf("\t%q +", line))
		}
		writeLn("\t\"\"")
		writeLn("")
	}
}

type Loader struct {
	hash        hash.Hash
	hashingDone bool
	// Map filename to content
	files map[string]*FileInfo
}

func newLoader() *Loader {
	return &Loader{
		hash:  sha256.New(),
		files: make(map[string]*FileInfo),
	}
}

func (l *Loader) record(f *FileInfo) *FileInfo {
	if !l.hashingDone {
		l.hash.Write(f.Content)
	}
	l.files[f.Filename] = f
	return f
}

func (l *Loader) LoadTextFile(filename string) *FileInfo {
	if fi := l.files[filename]; fi != nil {
		return fi
	}
	data, err := ioutil.ReadFile(filepath.FromSlash(filename))
	if err != nil {
		log.Fatal(err)
	}
	content := string(data)
	if textEOL != "\n" {
		// For text files, we want to keep the embedded data always in
		// '\n' format.
		content = strings.ReplaceAll(content, textEOL, "\n")
	}
	return l.record(&FileInfo{
		Filename: filename,
		Content:  []byte(data),
	})
}

func (l *Loader) LoadBinaryFile(filename string) *FileInfo {
	if fi := l.files[filename]; fi != nil {
		return fi
	}
	data, err := ioutil.ReadFile(filepath.FromSlash(filename))
	if err != nil {
		log.Fatal(err)
	}
	return l.record(&FileInfo{
		Filename: filename,
		Content:  data,
		Binary:   true,
	})
}

func (l *Loader) LoadTextGlob(glob string, excludes []string) []*FileInfo {
	var files []*FileInfo
	matches, err := filepath.Glob(filepath.FromSlash(glob))
	if err != nil {
		log.Fatal(err)
	}

	ex := map[string]bool{}
	for _, e := range excludes {
		ex[e] = true
	}

	for _, m := range matches {
		if ex[m] {
			continue
		}
		files = append(files, l.LoadTextFile(filepath.ToSlash(m)))
	}
	return files
}

func (l *Loader) MarkHashingDone() {
	l.hashingDone = true
}

func (l *Loader) VersionHash() string {
	if !l.hashingDone {
		panic("hashing still on going")
	}
	return hex.EncodeToString(l.hash.Sum(nil))
}

type FileInfo struct {
	// Filename, using slashes no matter the platform.
	Filename string
	// Content, with \n for EOL.
	Content []byte
	Binary  bool
}

func sortFiles(files []*FileInfo) []*FileInfo {
	sort.Slice(files, func(i, j int) bool {
		return files[i].Filename < files[j].Filename
	})
	return files
}

func main() {
	// Expects to be called from the base repository directory. This is the case
	// when called through "go generate", as Go uses the directory of the file
	// containing the statement - which is mapshot.go, at the base of the
	// repository.

	loader := newLoader()

	viewerFiles := sortFiles(append(append(
		loader.LoadTextGlob("frontend/dist/viewer/*.js", nil),
		loader.LoadTextFile("frontend/dist/viewer/index.html"),
		loader.LoadTextFile("frontend/dist/viewer/manifest.json"),
		loader.LoadBinaryFile("thumbnail.png"),
	), loader.LoadTextGlob("frontend/dist/viewer/*.svg", nil)...))

	listingFiles := sortFiles(append(append(
		loader.LoadTextGlob("frontend/dist/listing/*.js", nil),
		loader.LoadTextFile("frontend/dist/listing/index.html"),
		loader.LoadBinaryFile("thumbnail.png"),
	)))

	modFiles := sortFiles(append(
		loader.LoadTextGlob("mod/*.lua", []string{"mod/generated.lua"}),
		loader.LoadBinaryFile("thumbnail.png"),
		loader.LoadTextFile("mod/info.json"),
		loader.LoadTextFile("changelog.txt"),
		loader.LoadTextFile("LICENSE"),
		loader.LoadTextFile("README.md"),
	))

	loader.MarkHashingDone()

	version := getVersion()
	versionHash := loader.VersionHash()
	fmt.Println("Version:", version)
	fmt.Println("Version hash:", versionHash)

	// Generate Lua file first as it will be embedded also in Go module files.
	genLua(viewerFiles, version, versionHash)
	modFiles = append(modFiles, loader.LoadTextFile("mod/generated.lua"))
	genGo(modFiles, viewerFiles, listingFiles, version, versionHash)
}
