package internal

import (
	"flag"

	"github.com/majohn-r/output"
)

const (
	fieldKeyArguments = "arguments"
)

// ProcessArgs processes a slice of command line arguments and handles common
// errors therein
func ProcessArgs(o output.Bus, f *flag.FlagSet, args []string) (ok bool) {
	dereferencedArgs := make([]string, len(args))
	ok = true
	for i, arg := range args {
		var err error
		dereferencedArgs[i], err = InterpretEnvVarReferences(arg)
		if err != nil {
			o.WriteCanonicalError(UserBadArgument, arg, err)
			o.Log(output.Error, LogErrorBadArgument, map[string]any{
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
		o.Log(output.Error, err.Error(), map[string]any{
			fieldKeyArguments: dereferencedArgs,
		})
		ok = false
	} else {
		ok = true
	}
	return
}
