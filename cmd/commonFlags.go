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
	abbreviatedName string
	usage           string
	expectedType    ValueType
	defaultValue    any
}

func NewFlagDetails() *FlagDetails {
	return &FlagDetails{}
}

func (fD *FlagDetails) Copy() *FlagDetails {
	fDNew := NewFlagDetails()
	fDNew.abbreviatedName = fD.abbreviatedName
	fDNew.usage = fD.usage
	fDNew.expectedType = fD.expectedType
	fDNew.defaultValue = fD.defaultValue
	return fDNew
}

func (fD *FlagDetails) DefaultValue() any {
	return fD.defaultValue
}

func (fD *FlagDetails) WithAbbreviatedName(s string) *FlagDetails {
	fD.abbreviatedName = s
	return fD
}

func (fD *FlagDetails) WithUsage(s string) *FlagDetails {
	fD.usage = s
	return fD
}

func (fD *FlagDetails) WithExpectedType(t ValueType) *FlagDetails {
	fD.expectedType = t
	return fD
}

func (fD *FlagDetails) WithDefaultValue(a any) *FlagDetails {
	fD.defaultValue = a
	return fD
}

type SectionFlags struct {
	sectionName string
	flags       map[string]*FlagDetails // keys are flag names
}

func NewSectionFlags() *SectionFlags {
	return &SectionFlags{}
}

func (sF *SectionFlags) SectionName() string {
	return sF.sectionName
}

func (sF *SectionFlags) Flags() map[string]*FlagDetails {
	return sF.flags
}

func (sF *SectionFlags) WithSectionName(s string) *SectionFlags {
	sF.sectionName = s
	return sF
}

func (sF *SectionFlags) WithFlags(m map[string]*FlagDetails) *SectionFlags {
	sF.flags = m
	return sF
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

func AddFlags(o output.Bus, c *cmd_toolkit.Configuration, flags flagConsumer, defs ...*SectionFlags) {
	for _, def := range defs {
		config := c.SubConfiguration(def.sectionName)
		// sort names for deterministic test output
		sortedNames := []string{}
		for name := range def.flags {
			sortedNames = append(sortedNames, name)
		}
		slices.Sort(sortedNames)
		for _, name := range sortedNames {
			details := def.flags[name]
			if details != nil {
				details.AddFlag(o, config, flags, def.sectionName, name)
			} else {
				o.WriteCanonicalError("an internal error occurred: there are no details for flag %q", name)
				o.Log(output.Error, "internal error", map[string]any{
					"section": def.sectionName,
					"flag":    name,
					"error":   "no details present",
				})
			}
		}
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
	switch f.expectedType {
	case StringType:
		if statedDefault, _ok := f.defaultValue.(string); !_ok {
			reportDefaultTypeError(o, flagName, "string", f.defaultValue)
		} else {
			if newDefault, err := c.StringDefault(flagName, statedDefault); err != nil {
				cmd_toolkit.ReportInvalidConfigurationData(o, sectionName, err)
			} else {
				usage := cmd_toolkit.DecorateStringFlagUsage(f.usage, newDefault)
				if f.abbreviatedName == "" {
					flags.String(flagName, newDefault, usage)
				} else {
					flags.StringP(flagName, f.abbreviatedName, newDefault, usage)
				}
			}
		}
	case BoolType:
		if statedDefault, _ok := f.defaultValue.(bool); !_ok {
			reportDefaultTypeError(o, flagName, "bool", f.defaultValue)
		} else {
			if newDefault, err := c.BoolDefault(flagName, statedDefault); err != nil {
				cmd_toolkit.ReportInvalidConfigurationData(o, sectionName, err)
			} else {
				usage := cmd_toolkit.DecorateBoolFlagUsage(f.usage, newDefault)
				if f.abbreviatedName == "" {
					flags.Bool(flagName, newDefault, usage)
				} else {
					flags.BoolP(flagName, f.abbreviatedName, newDefault, usage)
				}
			}
		}
	case IntType:
		if bounds, _ok := f.defaultValue.(*cmd_toolkit.IntBounds); !_ok {
			reportDefaultTypeError(o, flagName, "*cmd_toolkit.IntBounds", f.defaultValue)
		} else {
			if newDefault, err := c.IntDefault(flagName, bounds); err != nil {
				cmd_toolkit.ReportInvalidConfigurationData(o, sectionName, err)
			} else {
				usage := cmd_toolkit.DecorateIntFlagUsage(f.usage, newDefault)
				if f.abbreviatedName == "" {
					flags.Int(flagName, newDefault, usage)
				} else {
					flags.IntP(flagName, f.abbreviatedName, newDefault, usage)
				}
			}
		}
	default:
		o.WriteCanonicalError(
			"An internal error occurred: unspecified flag type; section %q, flag %q", sectionName, flagName)
		o.Log(output.Error, "internal error", map[string]any{
			"section":        sectionName,
			"flag":           flagName,
			"specified-type": f.expectedType,
			"default":        f.defaultValue,
			"default-type":   reflect.TypeOf(f.defaultValue),
			"error":          "unspecified flag type",
		})
	}
}

type FlagValue struct {
	explicitlySet bool
	valueType     ValueType
	value         any
}

func NewFlagValue() *FlagValue {
	return &FlagValue{}
}

func (fV *FlagValue) WithExplicitlySet(b bool) *FlagValue {
	fV.explicitlySet = b
	return fV
}

func (fV *FlagValue) WithValueType(t ValueType) *FlagValue {
	fV.valueType = t
	return fV
}

func (fV *FlagValue) WithValue(a any) *FlagValue {
	fV.value = a
	return fV
}

type FlagProducer interface {
	Changed(name string) bool
	GetBool(name string) (bool, error)
	GetInt(name string) (int, error)
	GetString(name string) (string, error)
}

func ReadFlags(producer FlagProducer, defs *SectionFlags) (map[string]*FlagValue, []error) {
	m := map[string]*FlagValue{}
	e := []error{}
	// sort names for deterministic output in unit tests
	sortedNames := []string{}
	for name := range defs.flags {
		sortedNames = append(sortedNames, name)
	}
	slices.Sort(sortedNames)
	for _, name := range sortedNames {
		var err error
		details := defs.flags[name]
		if details == nil {
			err = fmt.Errorf("no details for flag %q", name)
		} else {
			val := &FlagValue{
				explicitlySet: producer.Changed(name),
				valueType:     details.expectedType,
			}
			switch details.expectedType {
			case BoolType:
				val.value, err = producer.GetBool(name)
			case StringType:
				val.value, err = producer.GetString(name)
			case IntType:
				val.value, err = producer.GetInt(name)
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
	} else if v, ok := fv.value.(bool); !ok {
		e = fmt.Errorf("flag value not boolean")
		o.WriteCanonicalError("an internal error occurred: flag %q is not boolean (%v)", flagName, fv.value)
		o.Log(output.Error, "internal error", map[string]any{
			"flag":  flagName,
			"value": fv.value,
			"error": e})
	} else {
		val = v
		userSet = fv.explicitlySet
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
	} else if v, ok := fv.value.(int); !ok {
		e = fmt.Errorf("flag value not int")
		o.WriteCanonicalError("an internal error occurred: flag %q is not an integer (%v)", flagName, fv.value)
		o.Log(output.Error, "internal error", map[string]any{
			"flag":  flagName,
			"value": fv.value,
			"error": e})
	} else {
		val = v
		userSet = fv.explicitlySet
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
	} else if v, ok := fv.value.(string); !ok {
		e = fmt.Errorf("flag value not string")
		o.WriteCanonicalError("an internal error occurred: flag %q is not a string (%v)", flagName, fv.value)
		o.Log(output.Error, "internal error", map[string]any{
			"flag":  flagName,
			"value": fv.value,
			"error": e})
	} else {
		val = v
		userSet = fv.explicitlySet
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
