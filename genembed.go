// Regenerate the mod data for embedding in Go/Lua.
package main

//go:generate go run genembed.go

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

func getVersion() string {
	raw, err := ioutil.ReadFile("info.json")
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

	f, err := os.Create("generated.lua")
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	write := func(s string) {
		if _, err := f.WriteString(s); err != nil {
			log.Fatal(err)
		}
	}

	write("-- Automatically generated, do not modify\n")
	write("local data = {}\n")
	write("data.html = [==[\n")
	write(content)
	write("]==]\n")
	write("return data\n")
}

var filenameSpecials = regexp.MustCompile(`[^a-zA-Z]`)

func filenameToVar(fname string) string {
	s := ""
	for _, p := range filenameSpecials.Split(fname, -1) {
		if len(p) == 0 {
			continue
		}
		s += strings.ToUpper(p[0:1]) + strings.ToLower(p[1:])
	}
	return s
}

func genGo(version string) {
	f, err := os.Create("embed/generated.go")
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	write := func(s string) {
		if _, err := f.WriteString(s); err != nil {
			log.Fatal(err)
		}
	}

	write("// Package embed is AUTOMATICALLY GENERATED, DO NOT EDIT\n")
	write("package embed\n\n")
	write("// Version of the mod\n")
	write(fmt.Sprintf("var Version = %q\n\n", version))

	// https://stackoverflow.com/a/34863211

	fileVarnames := make(map[string]string)

	var modFiles = []string{
		"*.lua",
		"changelog.txt",
		"info.json",
		"LICENSE",
		"README.md",
		"thumbnail.png",
	}

	var filenames []string
	for _, glob := range modFiles {
		matches, err := filepath.Glob(glob)
		if err != nil {
			log.Fatal(err)
		}
		for _, m := range matches {
			filenames = append(filenames, m)
			varName := "File" + filenameToVar(m)
			fileVarnames[m] = varName
		}
	}

	sort.Strings(filenames)
	for _, name := range filenames {
		data, err := ioutil.ReadFile(name)
		if err != nil {
			log.Fatal(err)
		}

		varName := fileVarnames[name]
		write(fmt.Sprintf("// %s is file %q\n", varName, name))
		write(fmt.Sprintf("var %s =\n", varName))
		for _, line := range strings.SplitAfter(string(data), "\n") {
			for len(line) > 120 {
				write(fmt.Sprintf("\t%q + // cont.\n", line[:120]))
				line = line[120:]
			}
			write(fmt.Sprintf("\t%q +\n", line))
		}
		write("\t\"\"\n")
	}
	write("\n")

	write("var ModFiles = map[string]string{\n")
	for _, name := range filenames {
		write(fmt.Sprintf("\t%q: %s,\n", name, fileVarnames[name]))
	}
	write("}\n")
}

func main() {
	version := getVersion()
	// Generate Lua file first as it will be embedded also in Go module files.
	genLua()
	genGo(version)
}
