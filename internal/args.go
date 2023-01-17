package internal

import (
	"flag"

	"github.com/majohn-r/output"
)

// ProcessArgs processes a slice of command line arguments and handles common
// errors therein
func ProcessArgs(o output.Bus, f *flag.FlagSet, args []string) (ok bool) {
	dereferencedArgs := make([]string, len(args))
	ok = true
	for i, arg := range args {
		var err error
		dereferencedArgs[i], err = dereferenceEnvVar(arg)
		if err != nil {
			o.WriteCanonicalError("The value for argument %q cannot be used: %v", arg, err)
			o.Log(output.Error, "argument cannot be used", map[string]any{
				"value": arg,
				"error": err,
			})
			ok = false
		}
	}
	if ok {
		f.SetOutput(o.ErrorWriter())
		// note: Parse outputs errors to o.ErrorWriter*()
		if err := f.Parse(dereferencedArgs); err != nil {
			o.Log(output.Error, err.Error(), map[string]any{"arguments": dereferencedArgs})
			ok = false
		}
	}
	return
}
