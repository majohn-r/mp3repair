package internal

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

const (
	defaultConfigFileName = "defaults.yaml"
	appDataVar            = "APPDATA"
	fkKey                 = "key"
	fkType                = "type"
	fkValue               = "value"
)

// ReadConfigurationFile reads defaults.yaml from the specified path and returns
// a pointer to a cooked Configuration instance
func ReadConfigurationFile(o OutputBus) (c *Configuration, ok bool) {
	var appDataValue string
	var appDataSet bool
	if appDataValue, appDataSet = appData(o); !appDataSet {
		c = EmptyConfiguration()
		ok = true
		return
	}
	path := CreateAppSpecificPath(appDataValue)
	configFile := filepath.Join(path, defaultConfigFileName)
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
	yfile, _ := ioutil.ReadFile(configFile) // only probable error circumvented by verifyFileExists failure
	data := make(map[string]interface{})
	if err = yaml.Unmarshal(yfile, &data); err != nil {
		o.LogWriter().Log(WARN, LW_CANNOT_UNMARSHAL_YAML, map[string]interface{}{
			FK_DIRECTORY: path,
			FK_FILE_NAME: defaultConfigFileName,
			FK_ERROR:     err,
		})
		fmt.Fprintf(o.ErrorWriter(), USER_CONFIGURATION_FILE_GARBLED, configFile, err)
		return
	}
	c = createConfiguration(o, data)
	ok = true
	o.LogWriter().Log(INFO, LI_CONFIGURATION_FILE_READ, map[string]interface{}{
		FK_DIRECTORY: path,
		FK_FILE_NAME: defaultConfigFileName,
		fkValue:      c,
	})
	return
}

func appData(o OutputBus) (string, bool) {
	if value, ok := os.LookupEnv(appDataVar); ok {
		return value, ok
	}
	o.LogWriter().Log(INFO, LI_NOT_SET, map[string]interface{}{
		fkVarName: appDataVar,
	})
	return "", false
}

func verifyFileExists(o OutputBus, path string) (ok bool, err error) {
	f, err := os.Stat(path)
	if err == nil {
		if f.IsDir() {
			o.LogWriter().Log(ERROR, LE_FILE_IS_DIR, map[string]interface{}{
				FK_DIRECTORY: filepath.Dir(path),
				FK_FILE_NAME: filepath.Base(path),
			})
			fmt.Fprintf(o.ErrorWriter(), USER_CONFIGURATION_FILE_IS_DIR, path)
			err = fmt.Errorf(ERROR_FILE_IS_DIR)
			return
		}
		ok = true
		return
	}
	if errors.Is(err, os.ErrNotExist) {
		o.LogWriter().Log(INFO, LI_NO_SUCH_FILE, map[string]interface{}{
			FK_DIRECTORY: filepath.Dir(path),
			FK_FILE_NAME: filepath.Base(path),
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
	if len(c.sMap) != 0 {
		output = append(output, fmt.Sprintf("%v", c.sMap))
	}
	if len(c.cMap) != 0 {
		output = append(output, fmt.Sprintf("%v", c.cMap))
	}
	return strings.Join(output, ", ")
}

func EmptyConfiguration() *Configuration {
	return &Configuration{
		bMap: make(map[string]bool),
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
func (c *Configuration) BoolDefault(key string, defaultValue bool) (b bool) {
	b = defaultValue
	if value, ok := c.bMap[key]; ok {
		b = value
	} else {
		if value, ok := c.sMap[key]; ok {
			rawValue := InterpretEnvVarReferences(value)
			if cookedValue, e := strconv.ParseBool(rawValue); e == nil {
				b = cookedValue
			}
		}
	}
	return
}

// StringDefault returns a string value for a specified key
func (c *Configuration) StringDefault(key string, defaultValue string) (s string) {
	s = InterpretEnvVarReferences(defaultValue)
	if value, ok := c.sMap[key]; ok {
		s = InterpretEnvVarReferences(value)
	}
	return
}

// StringValue returns the definition of the specified key and ok if the value
// is defined
func (c *Configuration) StringValue(key string) (value string, ok bool) {
	value, ok = c.sMap[key]
	return
}

type Configuration struct {
	sMap map[string]string
	bMap map[string]bool
	cMap map[string]*Configuration
}

func createConfiguration(o OutputBus, data map[string]interface{}) *Configuration {
	c := EmptyConfiguration()
	for key, v := range data {
		switch t := v.(type) {
		case string:
			c.sMap[key] = t
		case bool:
			c.bMap[key] = t
		case map[string]interface{}:
			c.cMap[key] = createConfiguration(o, t)
		default:
			o.LogWriter().Log(WARN, LW_UNEXPECTED_VALUE_TYPE, map[string]interface{}{
				fkKey:   key,
				fkValue: v,
				fkType:  fmt.Sprintf("%T", v),
			})
			c.sMap[key] = fmt.Sprintf("%v", v)
		}
	}
	return c
}
