package internal

import (
	"flag"

	"github.com/majohn-r/output"
)

// ProcessArgs processes a slice of command line arguments and handles common
// errors therein
func ProcessArgs(o output.Bus, f *flag.FlagSet, rawArgs []string) (ok bool) {
	args := make([]string, len(rawArgs))
	ok = true
	for i, arg := range rawArgs {
		var err error
		args[i], err = dereferenceEnvVar(arg)
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
		if err := f.Parse(args); err != nil {
			o.Log(output.Error, err.Error(), map[string]any{"arguments": args})
			ok = false
		}
	}
	return
}
