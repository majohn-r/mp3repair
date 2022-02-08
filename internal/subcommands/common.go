package subcommands

import (
	"flag"
	"fmt"
	"os"
)

type CommandProcessor interface {
	Name() string
	Exec([]string)
}

// processArgs parses command line arguments for a CommandProcessor; it does not
// return if there are errors in the input arguments
func processArgs(fs *flag.FlagSet, args []string){
	err := fs.Parse(args)
	if err != nil {
		fmt.Print(err)
		os.Exit(1)
	}
}