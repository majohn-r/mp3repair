package internal

import (
	"flag"
	"fmt"
)

const (
	fkArguments = "arguments"
)

func ProcessArgs(o OutputBus, f *flag.FlagSet, args []string) (ok bool) {
	dereferencedArgs := make([]string, len(args))
	for i, arg := range args {
		dereferencedArgs[i] = InterpretEnvVarReferences(arg)
	}
	f.SetOutput(o.ErrorWriter())
	// note: Parse outputs errors to o.ErrorWriter*()
	if err := f.Parse(dereferencedArgs); err != nil {
		o.LogWriter().Error(fmt.Sprintf("%v", err), map[string]interface{}{
			fkArguments: dereferencedArgs,
		})
	} else {
		ok = true
	}
	return
}
