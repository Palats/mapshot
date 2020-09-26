// +build ignore

// Regenerate the mod data for embedding in Go/Lua.
package main

import (
	"encoding/json"
	"fmt"
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

func genLua() {
	raw, err := ioutil.ReadFile("viewer.html")
	if err != nil {
		log.Fatal(err)
	}
	content := string(raw)
	if strings.Contains(content, "]==]") {
		log.Fatal("dumb Lua encoding cannot proceed")
	}

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
	writeLn("data.html = [==[")
	// This blindly copy the file content. As both the source and target files
	// are text file, they will preserve the end-of-lines.
	if _, err := f.WriteString(content); err != nil {
		log.Fatal(err)
	}
	writeLn("]==]")
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
	return s
}

type fileInfo struct {
	isText bool
	// The file path to read the file.
	localPath string
	// Path to the file using slashes on all platforms.
	genericPath string
	// The key in the generated array of embedded file.
	key string
	// The variable containing the data in the generated content.
	varName string
}

func genGo(version string) {
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

	// https://stackoverflow.com/a/34863211

	// modFiles is a list of globs to include. Boolean value is true if file is text.
	var modFiles = map[string]bool{
		"mod/*.lua":     true,
		"mod/info.json": true,
		"changelog.txt": true,
		"LICENSE":       true,
		"README.md":     true,
		"thumbnail.png": false,
	}

	var files []fileInfo
	for glob, isText := range modFiles {
		matches, err := filepath.Glob(glob)
		if err != nil {
			log.Fatal(err)
		}
		for _, m := range matches {
			info := fileInfo{
				isText:      isText,
				localPath:   m,
				genericPath: filepath.ToSlash(m),
				// Remove subpaths - this is used to generate the mod files, which is
				// flat structure.
				key:     filepath.Base(m),
				varName: "File" + filenameToVar(m),
			}
			files = append(files, info)
		}
	}

	sort.Slice(files, func(i, j int) bool {
		return files[i].key < files[j].key
	})

	for _, info := range files {
		data, err := ioutil.ReadFile(info.localPath)
		if err != nil {
			log.Fatal(err)
		}
		content := string(data)
		if info.isText && textEOL != "\n" {
			// For text files, we want to keep the embedded data always in
			// '\n' format.
			content = strings.ReplaceAll(content, textEOL, "\n")
		}

		writeLn(fmt.Sprintf("// %s is file %q", info.varName, info.genericPath))
		writeLn(fmt.Sprintf("var %s =", info.varName))

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
	for _, info := range files {
		writeLn(fmt.Sprintf("\t%q: %s,", info.key, info.varName))
	}
	writeLn("}")
}

func main() {
	// Expects to be called from the base repository directory. This is the case
	// when called through "go generate", as Go uses the directory of the file
	// containing the statement - which is mapshot.go, at the base of the
	// repository.

	version := getVersion()
	// Generate Lua file first as it will be embedded also in Go module files.
	genLua()
	genGo(version)
}
