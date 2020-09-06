// Regenerate the mod data for embedding in Go/Lua.
package main

//go:generate go run embed.go

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var modFiles = []string{
	"*.lua",
	"changelog.txt",
	"info.json",
	"LICENSE",
	"README.md",
	"thumbnail.png",
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

func genGo() {
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

	// https://stackoverflow.com/a/34863211

	allFiles := make(map[string]string)

	for _, glob := range modFiles {
		matches, err := filepath.Glob(glob)
		if err != nil {
			log.Fatal(err)
		}
		for _, m := range matches {
			data, err := ioutil.ReadFile(m)
			if err != nil {
				log.Fatal(err)
			}

			varName := "File" + filenameToVar(m)
			allFiles[m] = varName

			write(fmt.Sprintf("// %s is file %q\n", varName, m))
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
	}

	write("var ModFiles = map[string]string{\n")
	for name, varname := range allFiles {
		write(fmt.Sprintf("\t%q: %s,\n", "/"+name, varname))
	}
	write("}\n")
}

func main() {
	// Generate Lua file first as it will be embedded also in Go module files.
	genLua()
	genGo()
}
