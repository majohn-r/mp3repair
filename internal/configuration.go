package internal

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/majohn-r/output"
	"gopkg.in/yaml.v3"
)

// DefaultConfigFileName is the name of the configuration file that contains defaults for the commands
const DefaultConfigFileName = "defaults.yaml"

const (
	appDataVar      = "APPDATA"
	fieldKeyKeyName = "key"
	fieldKeyType    = "type"
)

var (
	appSpecificPath      string
	appSpecificPathValid bool
)

// GetAppSpecificPath returns the location for application-specific files
// (%APPPATH%\mp3) and whether that value is trustworthy
func GetAppSpecificPath() (string, bool) {
	return appSpecificPath, appSpecificPathValid
}

// SetAppSpecificPathForTesting sets the app-specific path internal variables;
// useful for test scenarios only
func SetAppSpecificPathForTesting(p string, v bool) {
	appSpecificPath = p
	appSpecificPathValid = v
}

// ReadConfigurationFile reads defaults.yaml from the specified path and returns
// a pointer to a cooked Configuration instance
func ReadConfigurationFile(o output.Bus) (c *Configuration, ok bool) {
	var appDataValue string
	var appDataSet bool
	if appDataValue, appDataSet = LookupAppData(o); !appDataSet {
		c = EmptyConfiguration()
		ok = true
		return
	}
	path := CreateAppSpecificPath(appDataValue)
	appSpecificPath = path
	appSpecificPathValid = true
	configFile := filepath.Join(path, DefaultConfigFileName)
	var err error
	var exists bool
	if exists, err = verifyFileExists(o, configFile); err != nil {
		return
	}
	if !exists {
		c = EmptyConfiguration()
		ok = true
		return
	}
	yfile, _ := os.ReadFile(configFile) // only probable error circumvented by verifyFileExists failure
	data, err := readYaml(yfile)
	if err != nil {
		o.Log(output.Error, LogErrorCannotUnmarshalYAML, map[string]any{
			FieldKeyDirectory: path,
			FieldKeyFileName:  DefaultConfigFileName,
			FieldKeyError:     err,
		})
		o.WriteCanonicalError(UserConfigurationFileGarbled, configFile, err)
		return
	}
	c = CreateConfiguration(o, data)
	ok = true
	o.Log(output.Info, LogInfoConfigurationFileRead, map[string]any{
		FieldKeyDirectory: path,
		FieldKeyFileName:  DefaultConfigFileName,
		FieldKeyValue:     c,
	})
	return
}

func readYaml(yfile []byte) (data map[string]any, err error) {
	data = make(map[string]any)
	err = yaml.Unmarshal(yfile, &data)
	return
}

// LookupAppData looks up the environment variable for finding application data
func LookupAppData(o output.Bus) (string, bool) {
	if value, ok := os.LookupEnv(appDataVar); ok {
		return value, ok
	}
	o.Log(output.Info, LogInfoNotSet, map[string]any{
		fieldKeyVarName: appDataVar,
	})
	return "", false
}

func verifyFileExists(o output.Bus, path string) (ok bool, err error) {
	f, err := os.Stat(path)
	if err == nil {
		if f.IsDir() {
			o.Log(output.Error, LogErrorFileIsDirectory, map[string]any{
				FieldKeyDirectory: filepath.Dir(path),
				FieldKeyFileName:  filepath.Base(path),
			})
			o.WriteCanonicalError(UserConfigurationFileIsDir, path)
			err = fmt.Errorf(ErrorFileIsDir)
			return
		}
		ok = true
		return
	}
	if errors.Is(err, os.ErrNotExist) {
		o.Log(output.Info, LogInfoNoSuchFile, map[string]any{
			FieldKeyDirectory: filepath.Dir(path),
			FieldKeyFileName:  filepath.Base(path),
		})
		err = nil
	}
	return
}

func (c *Configuration) String() string {
	var output []string
	if len(c.bMap) != 0 {
		output = append(output, fmt.Sprintf("%v", c.bMap))
	}
	if len(c.iMap) != 0 {
		output = append(output, fmt.Sprintf("%v", c.iMap))
	}
	if len(c.sMap) != 0 {
		output = append(output, fmt.Sprintf("%v", c.sMap))
	}
	if len(c.cMap) != 0 {
		output = append(output, fmt.Sprintf("%v", c.cMap))
	}
	return strings.Join(output, ", ")
}

// EmptyConfiguration creates an empty Configuration instance
func EmptyConfiguration() *Configuration {
	return &Configuration{
		bMap: make(map[string]bool),
		iMap: make(map[string]int),
		sMap: make(map[string]string),
		cMap: make(map[string]*Configuration),
	}
}

// SubConfiguration returns a specified sub-configuration
func (c *Configuration) SubConfiguration(key string) *Configuration {
	if configuration, ok := c.cMap[key]; ok {
		return configuration
	}
	return EmptyConfiguration()
}

// BoolDefault returns a boolean value for a specified key
func (c *Configuration) BoolDefault(key string, defaultValue bool) (b bool, err error) {
	b = defaultValue
	if value, ok := c.bMap[key]; ok {
		b = value
	} else {
		if value, ok := c.iMap[key]; ok {
			switch value {
			case 0:
				b = false
			case 1:
				b = true
			default:
				// note: deliberately imitating flags behavior when parsing an
				// invalid boolean
				err = fmt.Errorf("invalid boolean value \"%d\" for -%s: parse error", value, key)
			}
		} else {
			// True values may be specified as "t", "T", "true", "TRUE", or "True"
			// False values may be specified as "f", "F", "false", "FALSE", or "False"_."
			if value, ok := c.sMap[key]; ok {
				rawValue, dereferenceErr := InterpretEnvVarReferences(value)
				if dereferenceErr == nil {
					if cookedValue, e := strconv.ParseBool(rawValue); e == nil {
						b = cookedValue
					} else {
						// note: deliberately imitating flags behavior when parsing
						// an invalid boolean
						err = fmt.Errorf("invalid boolean value %q for -%s: parse error", value, key)
					}
				} else {
					err = fmt.Errorf("invalid boolean value %q for -%s: %v", value, key, dereferenceErr)
				}
			}
		}
	}
	return
}

// IntBounds holds the bounds for an int value which has a minimum value, a
// maximum value, and a default that lies within those bounds
type IntBounds struct {
	minValue     int
	defaultValue int
	maxValue     int
}

// NewIntBounds creates a instance of IntBounds, sorting the provided value into
// reasonable fields
func NewIntBounds(v1, v2, v3 int) *IntBounds {
	is := []int{v1, v2, v3}
	sort.Ints(is)
	return &IntBounds{
		minValue:     is[0],
		defaultValue: is[1],
		maxValue:     is[2],
	}
}

// IntDefault returns a default value for a specified key, which may or may not
// be defined in the Configuration instance
func (c *Configuration) IntDefault(key string, sortedBounds *IntBounds) (i int, err error) {
	i = sortedBounds.defaultValue
	if value, ok := c.iMap[key]; ok {
		i = constrainedValue(value, sortedBounds)
	} else {
		if value, ok := c.sMap[key]; ok {
			rawValue, dereferenceErr := InterpretEnvVarReferences(value)
			if dereferenceErr == nil {
				if cookedValue, e := strconv.Atoi(rawValue); e == nil {
					i = constrainedValue(cookedValue, sortedBounds)
				} else {
					// note: deliberately imitating flags behavior when parsing an
					// invalid int
					err = fmt.Errorf("invalid value %q for flag -%s: parse error", rawValue, key)
				}
			} else {
				err = fmt.Errorf("invalid value %q for flag -%s: %v", rawValue, key, dereferenceErr)
			}
		}
	}
	return
}

func constrainedValue(value int, sortedBounds *IntBounds) (i int) {
	if value < sortedBounds.minValue {
		i = sortedBounds.minValue
	} else if value > sortedBounds.maxValue {
		i = sortedBounds.maxValue
	} else {
		i = value
	}
	return
}

// StringDefault returns a string value for a specified key
func (c *Configuration) StringDefault(key string, defaultValue string) (s string, err error) {
	var dereferenceErr error
	s, dereferenceErr = InterpretEnvVarReferences(defaultValue)
	if dereferenceErr == nil {
		if value, ok := c.sMap[key]; ok {
			s, dereferenceErr = InterpretEnvVarReferences(value)
			if dereferenceErr != nil {
				err = fmt.Errorf("invalid value %q for flag -%s: %v", value, key, dereferenceErr)
				s = ""
			}
		}
	} else {
		err = fmt.Errorf("invalid value %q for flag -%s: %v", defaultValue, key, dereferenceErr)
		s = ""
	}
	return
}

// StringValue returns the definition of the specified key and ok if the value
// is defined
func (c *Configuration) StringValue(key string) (value string, ok bool) {
	value, ok = c.sMap[key]
	return
}

// Configuration defines the data structure for configuration information.
type Configuration struct {
	sMap map[string]string
	bMap map[string]bool
	iMap map[string]int
	cMap map[string]*Configuration
}

// CreateConfiguration returns a Configuration instance populated as specified
// by the data parameter
func CreateConfiguration(o output.Bus, data map[string]any) *Configuration {
	c := EmptyConfiguration()
	for key, v := range data {
		switch t := v.(type) {
		case string:
			c.sMap[key] = t
		case bool:
			c.bMap[key] = t
		case int:
			c.iMap[key] = t
		case map[string]any:
			c.cMap[key] = CreateConfiguration(o, t)
		default:
			o.Log(output.Error, LogErrorUnexpectedValueType, map[string]any{
				fieldKeyKeyName: key,
				FieldKeyValue:   v,
				fieldKeyType:    fmt.Sprintf("%T", v),
			})
			o.WriteCanonicalError(UserUnexpectedValueType, key, v, v)
			c.sMap[key] = fmt.Sprintf("%v", v)
		}
	}
	return c
}
