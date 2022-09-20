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
	ok = true
	for i, arg := range args {
		var err error
		dereferencedArgs[i], err = InterpretEnvVarReferences(arg)
		if err != nil {
			o.WriteError(USER_BAD_ARGUMENT, arg, err)
			o.LogWriter().Error(LE_BAD_ARGUMENT, map[string]any{
				FK_VALUE: arg,
				FK_ERROR: err,
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
