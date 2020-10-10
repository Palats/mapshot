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

func genLua(frontendFiles map[string]string, version, versionHash string) {
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

	var filenames []string
	for filename := range frontendFiles {
		filenames = append(filenames, filename)
	}
	sort.Strings(filenames)

	writeLn("-- Automatically generated, do not modify")
	writeLn("local data = {}")
	writeLn(fmt.Sprintf("data.version = %q", version))
	writeLn(fmt.Sprintf("data.version_hash = %q", versionHash))
	writeLn("data.files = {}")
	for _, filename := range filenames {
		content := frontendFiles[filename]
		if strings.Contains(content, "]==]") {
			log.Fatal("dumb Lua encoding cannot proceed")
		}

		key := filepath.Base(filename)

		writeLn(fmt.Sprintf("data.files[%q] = [==[", key))
		// This blindly copy the file content. As both the source and target files
		// are text file, they will preserve the end-of-lines.
		if _, err := f.WriteString(content); err != nil {
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
		} else {
			p = strings.ToUpper(p[0:1]) + strings.ToLower(p[1:])
		}
		s += p
	}
	return "File" + s
}

func genGo(modFiles map[string]string, version string, versionHash string) {
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

	var filenames []string
	for filename := range modFiles {
		filenames = append(filenames, filename)
	}
	sort.Strings(filenames)

	for _, filename := range filenames {
		content := modFiles[filename]
		varName := filenameToVar(filename)

		writeLn(fmt.Sprintf("// %s is file %q", varName, filepath.ToSlash(filename)))
		writeLn(fmt.Sprintf("var %s =", varName))

		for _, line := range strings.SplitAfter(content, "\n") {
			for len(line) > 120 {
				writeLn(fmt.Sprintf("\t%q + // cont.", line[:120]))
				line = line[120:]
			}
			writeLn(fmt.Sprintf("\t%q +", line))
		}
		writeLn("\t\"\"")
	}
	writeLn("")

	writeLn("// ModFiles is the list of files for the Factorio mod.")
	writeLn("var ModFiles = map[string]string{")
	for _, filename := range filenames {
		// Remove subpaths - this is used to generate the mod files, which is
		// flat structure.
		writeLn(fmt.Sprintf("\t%q: %s,", filepath.Base(filename), filenameToVar(filename)))
	}
	writeLn("}")
}

type Loader struct {
	hash    hash.Hash
	ignores map[string]bool
}

func newLoader(ignores []string) *Loader {
	l := &Loader{
		ignores: make(map[string]bool),
		hash:    sha256.New(),
	}
	for _, i := range ignores {
		l.ignores[i] = true
	}
	return l
}

func (l *Loader) record(filename string, content []byte) {
	if l.ignores[filepath.Base(filename)] {
		return
	}
	l.hash.Write(content)
}

func (l *Loader) LoadTextFile(filename string) string {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		log.Fatal(err)
	}
	content := string(data)
	if textEOL != "\n" {
		// For text files, we want to keep the embedded data always in
		// '\n' format.
		content = strings.ReplaceAll(content, textEOL, "\n")
	}
	l.record(filename, []byte(data))
	return content
}

func (l *Loader) LoadBinaryFile(filename string) []byte {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		log.Fatal(err)
	}
	l.record(filename, data)
	return data
}

func (l *Loader) LoadTextGlob(glob string) map[string]string {
	files := make(map[string]string)
	matches, err := filepath.Glob(glob)
	if err != nil {
		log.Fatal(err)
	}
	for _, m := range matches {
		files[m] = l.LoadTextFile(m)
	}
	return files
}

func (l *Loader) VersionHash() string {
	return hex.EncodeToString(l.hash.Sum(nil))
}

func main() {
	// Expects to be called from the base repository directory. This is the case
	// when called through "go generate", as Go uses the directory of the file
	// containing the statement - which is mapshot.go, at the base of the
	// repository.

	// Parameter is the list of file to ignore when creating the hash. Those
	// files are themselves generated, so reading them would create an unstable
	// hash.
	loader := newLoader([]string{"generated.lua"})

	frontendFiles := loader.LoadTextGlob("frontend/dist/*.js")
	frontendFiles["index.html"] = loader.LoadTextFile("frontend/dist/index.html")

	modFiles := loader.LoadTextGlob("mod/*.lua")
	modFiles["thumbnail.png"] = string(loader.LoadBinaryFile("thumbnail.png"))
	textfiles := []string{
		"mod/info.json",
		"changelog.txt",
		"LICENSE",
		"README.md",
		"thumbnail.png",
	}
	for _, filename := range textfiles {
		modFiles[filename] = loader.LoadTextFile(filename)
	}

	version := getVersion()
	versionHash := loader.VersionHash()
	fmt.Println("Version:", version)
	fmt.Println("Version hash:", versionHash)

	// Generate Lua file first as it will be embedded also in Go module files.
	genLua(frontendFiles, version, versionHash)
	genGo(modFiles, version, versionHash)
}
