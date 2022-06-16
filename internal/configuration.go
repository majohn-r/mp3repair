package internal

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
)

const defaultConfigFileName = "defaults.yaml"

// ReadConfigurationFile reads defaults.yaml from the specified path and returns
// a pointer to a cooked Node instance
func ReadConfigurationFile(path string) *Configuration {
	var yfile []byte
	var err error
	if yfile, err = ioutil.ReadFile(filepath.Join(path, defaultConfigFileName)); err != nil {
		logrus.WithFields(logrus.Fields{
			FK_DIRECTORY: path,
			FK_FILE_NAME: defaultConfigFileName,
			FK_ERROR:     err,
		}).Warn(LW_CANNOT_READ_FILE)
		return EmptyConfiguration()
	}
	data := make(map[string]interface{})
	if err = yaml.Unmarshal(yfile, &data); err != nil {
		logrus.WithFields(logrus.Fields{
			FK_DIRECTORY: path,
			FK_FILE_NAME: defaultConfigFileName,
			FK_ERROR:     err,
		}).Warn(LW_CANNOT_UNMARSHAL_YAML)
		return EmptyConfiguration()
	}
	configuration := createConfiguration(data)
	logrus.WithFields(logrus.Fields{
		FK_DIRECTORY: path,
		FK_FILE_NAME: defaultConfigFileName,
		FK_VALUE:     configuration,
	}).Info(LI_CONFIGURATION_FILE_READ)
	return configuration
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

type Configuration struct {
	sMap map[string]string
	bMap map[string]bool
	cMap map[string]*Configuration
}

func createConfiguration(data map[string]interface{}) *Configuration {
	c := EmptyConfiguration()
	for key, v := range data {
		switch t := v.(type) {
		case string:
			c.sMap[key] = t
		case bool:
			c.bMap[key] = t
		case map[string]interface{}:
			c.cMap[key] = createConfiguration(t)
		default:
			logrus.WithFields(logrus.Fields{
				FK_KEY:   key,
				FK_VALUE: v,
				FK_TYPE:  fmt.Sprintf("%T", v),
			}).Warn(LW_UNEXPECTED_VALUE_TYPE)
			c.sMap[key] = fmt.Sprintf("%v", v)
		}
	}
	return c
}
