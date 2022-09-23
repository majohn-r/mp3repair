package internal

import (
	"flag"
	"fmt"
)

const (
	fkArguments = "arguments"
)

// ProcessArgs processes a slice of command line arguments and handles common
// errors therein
func ProcessArgs(o OutputBus, f *flag.FlagSet, args []string) (ok bool) {
	dereferencedArgs := make([]string, len(args))
	ok = true
	for i, arg := range args {
		var err error
		dereferencedArgs[i], err = InterpretEnvVarReferences(arg)
		if err != nil {
			o.WriteError(UserBadArgument, arg, err)
			o.LogWriter().Error(LogErrorBadArgument, map[string]any{
				FieldKeyValue: arg,
				FieldKeyError: err,
			})
			ok = false
		}
	}
	if !ok {
		return
	}
	f.SetOutput(o.ErrorWriter())
	// note: Parse outputs errors to o.ErrorWriter*()
	if err := f.Parse(dereferencedArgs); err != nil {
		o.LogWriter().Error(fmt.Sprintf("%v", err), map[string]any{
			fkArguments: dereferencedArgs,
		})
		ok = false
	} else {
		ok = true
	}
	return
}
