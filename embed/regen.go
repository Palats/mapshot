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

func genLua(frontendFiles []*FileInfo, version, versionHash string) {
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

	for _, fi := range frontendFiles {
		if strings.Contains(string(fi.Content), "]==]") {
			log.Fatal("dumb Lua encoding cannot proceed")
		}

		key := filepath.Base(fi.Filename)

		writeLn(fmt.Sprintf("data.files[%q] = [==[", key))
		// This blindly copy the file content. As both the source and target files
		// are text file, they will preserve the end-of-lines.
		if _, err := f.Write(fi.Content); err != nil {
			log.Fatal(err)
		}
		writeLn("]==]")
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

func genGo(modFiles []*FileInfo, frontendFiles []*FileInfo, version string, versionHash string) {
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

	writeLn("// FrontendFiles is the files for the UI to navigate the mapshots.")
	writeLn("var FrontendFiles = map[string]string{")
	for _, fi := range frontendFiles {
		files = append(files, fi)
		writeLn(fmt.Sprintf("\t%q: %s,", filepath.Base(fi.Filename), filenameToVar(fi.Filename)))
	}
	writeLn("}")
	writeLn("")

	for _, fi := range sortFiles(files) {
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
	hash    hash.Hash
	ignores map[string]bool
	// Map filename to content
	files map[string]*FileInfo
}

func newLoader(ignores []string) *Loader {
	l := &Loader{
		ignores: make(map[string]bool),
		hash:    sha256.New(),
		files:   make(map[string]*FileInfo),
	}
	for _, i := range ignores {
		l.ignores[i] = true
	}
	return l
}

func (l *Loader) record(f *FileInfo) *FileInfo {
	l.files[f.Filename] = f
	if !l.ignores[f.Filename] {
		l.hash.Write(f.Content)
	}
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
	})
}

func (l *Loader) LoadTextGlob(glob string) []*FileInfo {
	var files []*FileInfo
	matches, err := filepath.Glob(filepath.FromSlash(glob))
	if err != nil {
		log.Fatal(err)
	}
	for _, m := range matches {
		files = append(files, l.LoadTextFile(filepath.ToSlash(m)))
	}
	return files
}

func (l *Loader) VersionHash() string {
	return hex.EncodeToString(l.hash.Sum(nil))
}

type FileInfo struct {
	// Filename, using slashes no matter the platform.
	Filename string
	// Content, with \n for EOL.
	Content []byte
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

	// Parameter is the list of file to ignore when creating the hash. Those
	// files are themselves generated, so reading them would create an unstable
	// hash.
	loader := newLoader([]string{"mod/generated.lua"})

	frontendFiles := sortFiles(append(
		loader.LoadTextGlob("frontend/dist/*.js"),
		loader.LoadTextFile("frontend/dist/index.html"),
	))

	modFiles := sortFiles(append(
		loader.LoadTextGlob("mod/*.lua"),
		loader.LoadBinaryFile("thumbnail.png"),
		loader.LoadTextFile("mod/info.json"),
		loader.LoadTextFile("changelog.txt"),
		loader.LoadTextFile("LICENSE"),
		loader.LoadTextFile("README.md"),
	))

	version := getVersion()
	versionHash := loader.VersionHash()
	fmt.Println("Version:", version)
	fmt.Println("Version hash:", versionHash)

	// Generate Lua file first as it will be embedded also in Go module files.
	genLua(frontendFiles, version, versionHash)
	genGo(modFiles, frontendFiles, version, versionHash)
}
