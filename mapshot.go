package main

//go:generate go run embed/regen.go

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/Palats/mapshot/cmd"
	"github.com/inconshreveable/mousetrap"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

func main() {
	// By default, Cobra prevents running CLI from Windows Explorer. Setting
	// this var to empty let it run normally.
	cobra.MousetrapHelpText = ""

	// Running from Explorer means that:
	//  - The user probably just double-clicked on the binary.
	//  - The console window disappears as soon as the program exists.
	// So we want to have a useful default behavior and keep the output
	// displayed for a bit.
	if mousetrap.StartedByExplorer() {
		// os.Args contains the program name and other arguments.
		if len(os.Args) <= 1 {
			fmt.Println("No parameter given, running as a HTTP server.")
			fmt.Println("Mapshot has more functionalities, run it from cmd.exe to see more.")
			fmt.Println()
			os.Args = append(os.Args, "serve")
		}

		defer func() {
			fmt.Println("\nPress return to exit...")
			fmt.Scanln()
		}()
	}

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
