package main

//go:generate go run embed/regen.go

import (
	"context"
	"flag"
	"os"

	"github.com/Palats/mapshot/cmd"
	"github.com/spf13/pflag"
)

func main() {
	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)

	// Fake parse the default Go flags - that appease glog, which otherwise
	// complains on each line. goflag.CommandLine do get parsed in parsed
	// through pflag and `AddGoFlagSet`.
	flag.CommandLine.Parse(nil)

	if err := cmd.Execute(context.Background()); err != nil {
		// Root cmd already prints errors of subcommands.
		os.Exit(1)
	}
}
