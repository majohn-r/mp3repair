package cmd

import (
	"fmt"
	"reflect"
	"slices"

	cmd_toolkit "github.com/majohn-r/cmd-toolkit"
	"github.com/majohn-r/output"
)

type ValueType int32

const (
	UnspecifiedType ValueType = iota
	BoolType
	IntType
	StringType
)

type FlagDetails struct {
	AbbreviatedName string
	Usage           string
	ExpectedType    ValueType
	DefaultValue    any
}

func (fD *FlagDetails) Copy() *FlagDetails {
	fDNew := &FlagDetails{}
	fDNew.AbbreviatedName = fD.AbbreviatedName
	fDNew.Usage = fD.Usage
	fDNew.ExpectedType = fD.ExpectedType
	fDNew.DefaultValue = fD.DefaultValue
	return fDNew
}

type SectionFlags struct {
	SectionName string
	Details     map[string]*FlagDetails // keys are flag names
}

type flagConsumer interface {
	String(name string, value string, usage string) *string
	StringP(name, shorthand string, value string, usage string) *string
	Bool(name string, value bool, usage string) *bool
	BoolP(name, shorthand string, value bool, usage string) *bool
	Int(name string, value int, usage string) *int
	IntP(name, shorthand string, value int, usage string) *int
}

type ConfigSource interface {
	BoolDefault(string, bool) (bool, error)
	IntDefault(string, *cmd_toolkit.IntBounds) (int, error)
	StringDefault(string, string) (string, error)
}

func AddFlags(o output.Bus, c *cmd_toolkit.Configuration, flags flagConsumer,
	defs ...*SectionFlags) {
	for _, def := range defs {
		config := c.SubConfiguration(def.SectionName)
		// sort names for deterministic test output
		sortedNames := make([]string, 0, len(def.Details))
		for name := range def.Details {
			sortedNames = append(sortedNames, name)
		}
		slices.Sort(sortedNames)
		for _, name := range sortedNames {
			details := def.Details[name]
			switch details {
			case nil:
				o.WriteCanonicalError(
					"an internal error occurred: there are no details for flag %q", name)
				o.Log(output.Error, "internal error", map[string]any{
					"section": def.SectionName,
					"flag":    name,
					"error":   "no details present",
				})
			default:
				details.AddFlag(o, config, flags, Flag{
					Section: def.SectionName,
					Name:    name,
				})
			}
		}
	}
}

func reportDefaultTypeError(o output.Bus, flag, expected string, value any) {
	o.WriteCanonicalError(
		"an internal error occurred: the type of flag %q's value, '%v', is '%T',"+
			" but '%s' was expected", flag, value, value, expected)
	o.Log(output.Error, "internal error", map[string]any{
		"flag":     flag,
		"value":    value,
		"expected": expected,
		"actual":   reflect.TypeOf(value),
		"error":    "default value mistyped",
	})
}

type Flag struct {
	Section string
	Name    string
}

func (f *FlagDetails) AddFlag(o output.Bus, c ConfigSource, flags flagConsumer, flag Flag) {
	switch f.ExpectedType {
	case StringType:
		statedDefault, _ok := f.DefaultValue.(string)
		if !_ok {
			reportDefaultTypeError(o, flag.Name, "string", f.DefaultValue)
			return
		}
		newDefault, malformedDefault := c.StringDefault(flag.Name, statedDefault)
		if malformedDefault != nil {
			cmd_toolkit.ReportInvalidConfigurationData(o, flag.Section, malformedDefault)
			return
		}
		usage := cmd_toolkit.DecorateStringFlagUsage(f.Usage, newDefault)
		switch f.AbbreviatedName {
		case "":
			flags.String(flag.Name, newDefault, usage)
		default:
			flags.StringP(flag.Name, f.AbbreviatedName, newDefault, usage)
		}
	case BoolType:
		statedDefault, _ok := f.DefaultValue.(bool)
		if !_ok {
			reportDefaultTypeError(o, flag.Name, "bool", f.DefaultValue)
			return
		}
		newDefault, malformedDefault := c.BoolDefault(flag.Name, statedDefault)
		if malformedDefault != nil {
			cmd_toolkit.ReportInvalidConfigurationData(o, flag.Section, malformedDefault)
			return
		}
		usage := cmd_toolkit.DecorateBoolFlagUsage(f.Usage, newDefault)
		switch f.AbbreviatedName {
		case "":
			flags.Bool(flag.Name, newDefault, usage)
		default:
			flags.BoolP(flag.Name, f.AbbreviatedName, newDefault, usage)
		}
	case IntType:
		bounds, _ok := f.DefaultValue.(*cmd_toolkit.IntBounds)
		if !_ok {
			reportDefaultTypeError(o, flag.Name, "*cmd_toolkit.IntBounds", f.DefaultValue)
			return
		}
		newDefault, malformedDefault := c.IntDefault(flag.Name, bounds)
		if malformedDefault != nil {
			cmd_toolkit.ReportInvalidConfigurationData(o, flag.Section, malformedDefault)
			return
		}
		usage := cmd_toolkit.DecorateIntFlagUsage(f.Usage, newDefault)
		switch f.AbbreviatedName {
		case "":
			flags.Int(flag.Name, newDefault, usage)
		default:
			flags.IntP(flag.Name, f.AbbreviatedName, newDefault, usage)
		}
	default:
		o.WriteCanonicalError(
			"An internal error occurred: unspecified flag type; section %q, flag %q",
			flag.Section, flag.Name)
		o.Log(output.Error, "internal error", map[string]any{
			"section":        flag.Section,
			"flag":           flag.Name,
			"specified-type": f.ExpectedType,
			"default":        f.DefaultValue,
			"default-type":   reflect.TypeOf(f.DefaultValue),
			"error":          "unspecified flag type",
		})
	}
}

type commandFlagValue interface {
	string | int | bool | any
}

type CommandFlag[V commandFlagValue] struct {
	Value   V
	UserSet bool
}

type FlagProducer interface {
	Changed(name string) bool
	GetBool(name string) (bool, error)
	GetInt(name string) (int, error)
	GetString(name string) (string, error)
}

func ReadFlags(producer FlagProducer, defs *SectionFlags) (map[string]*CommandFlag[any], []error) {
	m := map[string]*CommandFlag[any]{}
	e := []error{}
	// sort names for deterministic output in unit tests
	sortedNames := make([]string, 0, len(defs.Details))
	for name := range defs.Details {
		sortedNames = append(sortedNames, name)
	}
	slices.Sort(sortedNames)
	for _, name := range sortedNames {
		details := defs.Details[name]
		if details == nil {
			e = append(e, fmt.Errorf("no details for flag %q", name))
			continue
		}
		val := &CommandFlag[any]{
			UserSet: producer.Changed(name),
		}
		var flagError error
		switch details.ExpectedType {
		case BoolType:
			val.Value, flagError = producer.GetBool(name)
		case StringType:
			val.Value, flagError = producer.GetString(name)
		case IntType:
			val.Value, flagError = producer.GetInt(name)
		default:
			flagError = fmt.Errorf("unknown type for flag --%s", name)
		}
		switch flagError {
		case nil:
			m[name] = val
		default:
			e = append(e, flagError)
		}
	}
	return m, e
}

func GetBool(o output.Bus, results map[string]*CommandFlag[any], flagName string) (CommandFlag[bool], error) {
	fv, flagNotFound := extractFlagValue(o, results, flagName)
	if flagNotFound != nil {
		return CommandFlag[bool]{}, flagNotFound
	}
	if fv == nil {
		e := fmt.Errorf("no data associated with flag")
		o.WriteCanonicalError("an internal error occurred: flag %q has no data", flagName)
		o.Log(output.Error, "internal error", map[string]any{
			"flag":  flagName,
			"error": e})
		return CommandFlag[bool]{}, e
	}
	v, ok := fv.Value.(bool)
	if !ok {
		e := fmt.Errorf("flag value not boolean")
		o.WriteCanonicalError("an internal error occurred: flag %q is not boolean (%v)",
			flagName, fv.Value)
		o.Log(output.Error, "internal error", map[string]any{
			"flag":  flagName,
			"value": fv.Value,
			"error": e})
		return CommandFlag[bool]{}, e
	}
	return CommandFlag[bool]{
		Value:   v,
		UserSet: fv.UserSet,
	}, nil
}

func GetInt(o output.Bus, results map[string]*CommandFlag[any], flagName string) (CommandFlag[int], error) {
	fv, flagNotFound := extractFlagValue(o, results, flagName)
	if flagNotFound != nil {
		return CommandFlag[int]{}, flagNotFound
	}
	if fv == nil {
		e := fmt.Errorf("no data associated with flag")
		o.WriteCanonicalError("an internal error occurred: flag %q has no data", flagName)
		o.Log(output.Error, "internal error", map[string]any{
			"flag":  flagName,
			"error": e})
		return CommandFlag[int]{}, e
	}
	v, ok := fv.Value.(int)
	if !ok {
		e := fmt.Errorf("flag value not int")
		o.WriteCanonicalError("an internal error occurred: flag %q is not an integer (%v)",
			flagName, fv.Value)
		o.Log(output.Error, "internal error", map[string]any{
			"flag":  flagName,
			"value": fv.Value,
			"error": e})
		return CommandFlag[int]{}, e
	}
	return CommandFlag[int]{Value: v, UserSet: fv.UserSet}, nil
}

func GetString(o output.Bus, results map[string]*CommandFlag[any], flagName string) (CommandFlag[string], error) {
	fv, flagNotFound := extractFlagValue(o, results, flagName)
	if flagNotFound != nil {
		return CommandFlag[string]{}, flagNotFound
	}
	if fv == nil {
		e := fmt.Errorf("no data associated with flag")
		o.WriteCanonicalError("an internal error occurred: flag %q has no data", flagName)
		o.Log(output.Error, "internal error", map[string]any{
			"flag":  flagName,
			"error": e})
		return CommandFlag[string]{}, e
	}
	v, ok := fv.Value.(string)
	if !ok {
		e := fmt.Errorf("flag value not string")
		o.WriteCanonicalError("an internal error occurred: flag %q is not a string (%v)",
			flagName, fv.Value)
		o.Log(output.Error, "internal error", map[string]any{
			"flag":  flagName,
			"value": fv.Value,
			"error": e})
		return CommandFlag[string]{}, e
	}
	return CommandFlag[string]{Value: v, UserSet: fv.UserSet}, nil
}

func extractFlagValue(o output.Bus, results map[string]*CommandFlag[any], flagName string) (fv *CommandFlag[any], e error) {
	if results == nil {
		e = fmt.Errorf("nil results")
		o.WriteCanonicalError("an internal error occurred: no flag values exist")
		o.Log(output.Error, "internal error", map[string]any{
			"error": "no results to extract flag values from",
		})
		return
	}
	value, found := results[flagName]
	if !found {
		e = fmt.Errorf("flag not found")
		o.WriteCanonicalError("an internal error occurred: flag %q is not found", flagName)
		o.Log(output.Error, "internal error", map[string]any{
			"flag":  flagName,
			"error": e,
		})
		return
	}
	fv = value
	return
}

func ProcessFlagErrors(o output.Bus, eSlice []error) bool {
	if len(eSlice) != 0 {
		for _, e := range eSlice {
			o.WriteCanonicalError("an internal error occurred: %v", e)
			o.Log(output.Error, "internal error", map[string]any{"error": e})
		}
		return false
	}
	return true
}

func init() {
	addDefaults(SearchFlags)
}
