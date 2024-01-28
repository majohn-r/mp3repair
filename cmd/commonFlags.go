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

type SectionFlags struct {
	SectionName string
	Flags       map[string]*FlagDetails // keys are flag names
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

func AddFlags(o output.Bus, c *cmd_toolkit.Configuration, flags flagConsumer, defs SectionFlags, includeSearches bool) {
	config := c.SubConfiguration(defs.SectionName)
	// sort names for deterministic test output
	sortedNames := []string{}
	for name := range defs.Flags {
		sortedNames = append(sortedNames, name)
	}
	slices.Sort(sortedNames)
	for _, name := range sortedNames {
		details := defs.Flags[name]
		if details != nil {
			details.AddFlag(o, config, flags, defs.SectionName, name)
		} else {
			o.WriteCanonicalError("an internal error occurred: there are no details for flag %q", name)
			o.Log(output.Error, "internal error", map[string]any{
				"section": defs.SectionName,
				"flag":    name,
				"error":   "no details present",
			})
		}
	}
	if includeSearches {
		AddFlags(o, c, flags, SearchFlags, false)
	}
}

func reportDefaultTypeError(o output.Bus, flag, expected string, value any) {
	o.WriteCanonicalError("an internal error occurred: the type of flag %q's value, '%v', is '%T', but '%s' was expected", flag, value, value, expected)
	o.Log(output.Error, "internal error", map[string]any{
		"flag":     flag,
		"value":    value,
		"expected": expected,
		"actual":   reflect.TypeOf(value),
		"error":    "default value mistyped",
	})
}

func (f *FlagDetails) AddFlag(o output.Bus, c ConfigSource, flags flagConsumer, sectionName, flagName string) {
	switch f.ExpectedType {
	case StringType:
		if statedDefault, _ok := f.DefaultValue.(string); !_ok {
			reportDefaultTypeError(o, flagName, "string", f.DefaultValue)
		} else {
			if newDefault, err := c.StringDefault(flagName, statedDefault); err != nil {
				cmd_toolkit.ReportInvalidConfigurationData(o, sectionName, err)
			} else {
				usage := cmd_toolkit.DecorateStringFlagUsage(f.Usage, newDefault)
				if f.AbbreviatedName == "" {
					flags.String(flagName, newDefault, usage)
				} else {
					flags.StringP(flagName, f.AbbreviatedName, newDefault, usage)
				}
			}
		}
	case BoolType:
		if statedDefault, _ok := f.DefaultValue.(bool); !_ok {
			reportDefaultTypeError(o, flagName, "bool", f.DefaultValue)
		} else {
			if newDefault, err := c.BoolDefault(flagName, statedDefault); err != nil {
				cmd_toolkit.ReportInvalidConfigurationData(o, sectionName, err)
			} else {
				usage := cmd_toolkit.DecorateBoolFlagUsage(f.Usage, newDefault)
				if f.AbbreviatedName == "" {
					flags.Bool(flagName, newDefault, usage)
				} else {
					flags.BoolP(flagName, f.AbbreviatedName, newDefault, usage)
				}
			}
		}
	case IntType:
		if bounds, _ok := f.DefaultValue.(*cmd_toolkit.IntBounds); !_ok {
			reportDefaultTypeError(o, flagName, "*cmd_toolkit.IntBounds", f.DefaultValue)
		} else {
			if newDefault, err := c.IntDefault(flagName, bounds); err != nil {
				cmd_toolkit.ReportInvalidConfigurationData(o, sectionName, err)
			} else {
				usage := cmd_toolkit.DecorateIntFlagUsage(f.Usage, newDefault)
				if f.AbbreviatedName == "" {
					flags.Int(flagName, newDefault, usage)
				} else {
					flags.IntP(flagName, f.AbbreviatedName, newDefault, usage)
				}
			}
		}
	default:
		o.WriteCanonicalError(
			"An internal error occurred: unspecified flag type; section %q, flag %q", sectionName, flagName)
		o.Log(output.Error, "internal error", map[string]any{
			"section":        sectionName,
			"flag":           flagName,
			"specified-type": f.ExpectedType,
			"default":        f.DefaultValue,
			"default-type":   reflect.TypeOf(f.DefaultValue),
			"error":          "unspecified flag type",
		})
	}
}

type FlagValue struct {
	ExplicitlySet bool
	ValueType     ValueType
	Value         any
}

type FlagProducer interface {
	Changed(name string) bool
	GetBool(name string) (bool, error)
	GetInt(name string) (int, error)
	GetString(name string) (string, error)
}

func ReadFlags(producer FlagProducer, defs SectionFlags) (map[string]*FlagValue, []error) {
	m := map[string]*FlagValue{}
	e := []error{}
	// sort names for deterministic output in unit tests
	sortedNames := []string{}
	for name := range defs.Flags {
		sortedNames = append(sortedNames, name)
	}
	slices.Sort(sortedNames)
	for _, name := range sortedNames {
		var err error
		details := defs.Flags[name]
		if details == nil {
			err = fmt.Errorf("no details for flag %q", name)
		} else {
			val := &FlagValue{
				ExplicitlySet: producer.Changed(name),
				ValueType:     details.ExpectedType,
			}
			switch details.ExpectedType {
			case BoolType:
				val.Value, err = producer.GetBool(name)
			case StringType:
				val.Value, err = producer.GetString(name)
			case IntType:
				val.Value, err = producer.GetInt(name)
			default:
				err = fmt.Errorf("unknown type for flag --%s", name)
			}
			if err == nil {
				m[name] = val
			}
		}
		if err != nil {
			e = append(e, err)
		}
	}
	return m, e
}

func GetBool(o output.Bus, results map[string]*FlagValue, flagName string) (val, userSet bool, e error) {
	if fv, err := extractFlagValue(o, results, flagName); err != nil {
		e = err
	} else if fv == nil {
		e = fmt.Errorf("no data associated with flag")
		o.WriteCanonicalError("an internal error occurred: flag %q has no data", flagName)
		o.Log(output.Error, "internal error", map[string]any{
			"flag":  flagName,
			"error": e})
	} else if v, ok := fv.Value.(bool); !ok {
		e = fmt.Errorf("flag value not boolean")
		o.WriteCanonicalError("an internal error occurred: flag %q is not boolean (%v)", flagName, fv.Value)
		o.Log(output.Error, "internal error", map[string]any{
			"flag":  flagName,
			"value": fv.Value,
			"error": e})
	} else {
		val = v
		userSet = fv.ExplicitlySet
	}
	return
}

func GetInt(o output.Bus, results map[string]*FlagValue, flagName string) (val int, userSet bool, e error) {
	if fv, err := extractFlagValue(o, results, flagName); err != nil {
		e = err
	} else if fv == nil {
		e = fmt.Errorf("no data associated with flag")
		o.WriteCanonicalError("an internal error occurred: flag %q has no data", flagName)
		o.Log(output.Error, "internal error", map[string]any{
			"flag":  flagName,
			"error": e})
	} else if v, ok := fv.Value.(int); !ok {
		e = fmt.Errorf("flag value not int")
		o.WriteCanonicalError("an internal error occurred: flag %q is not an integer (%v)", flagName, fv.Value)
		o.Log(output.Error, "internal error", map[string]any{
			"flag":  flagName,
			"value": fv.Value,
			"error": e})
	} else {
		val = v
		userSet = fv.ExplicitlySet
	}
	return
}

func GetString(o output.Bus, results map[string]*FlagValue, flagName string) (val string, userSet bool, e error) {
	if fv, err := extractFlagValue(o, results, flagName); err != nil {
		e = err
	} else if fv == nil {
		e = fmt.Errorf("no data associated with flag")
		o.WriteCanonicalError("an internal error occurred: flag %q has no data", flagName)
		o.Log(output.Error, "internal error", map[string]any{
			"flag":  flagName,
			"error": e})
	} else if v, ok := fv.Value.(string); !ok {
		e = fmt.Errorf("flag value not string")
		o.WriteCanonicalError("an internal error occurred: flag %q is not a string (%v)", flagName, fv.Value)
		o.Log(output.Error, "internal error", map[string]any{
			"flag":  flagName,
			"value": fv.Value,
			"error": e})
	} else {
		val = v
		userSet = fv.ExplicitlySet
	}
	return
}

func extractFlagValue(o output.Bus, results map[string]*FlagValue, flagName string) (fv *FlagValue, e error) {
	if results == nil {
		e = fmt.Errorf("nil results")
		o.WriteCanonicalError("an internal error occurred: no flag values exist")
		o.Log(output.Error, "internal error", map[string]any{"error": "no results to extract flag values from"})
	} else if value, ok := results[flagName]; !ok {
		e = fmt.Errorf("flag not found")
		o.WriteCanonicalError("an internal error occurred: flag %q is not found", flagName)
		o.Log(output.Error, "internal error", map[string]any{
			"flag":  flagName,
			"error": e,
		})
	} else {
		fv = value
	}
	return
}

func ProcessFlagErrors(o output.Bus, eSlice []error) (ok bool) {
	if len(eSlice) != 0 {
		for _, e := range eSlice {
			o.WriteCanonicalError("an internal error occurred: %v", e)
			o.Log(output.Error, "internal error", map[string]any{"error": e})
		}
	} else {
		ok = true
	}
	return
}

func init() {
	addDefaults(SearchFlags)
}
