package internal

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strconv"

	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

const (
	DefaultConfigFileName = "defaults.yaml"
)

// ReadYaml reads defaults.yaml from the specified path and returns a pointer to a cooked Node instance
func ReadYaml(path string) *Node {
	var yfile []byte
	var err error
	if yfile, err = ioutil.ReadFile(filepath.Join(path, DefaultConfigFileName)); err != nil {
		logrus.WithFields(logrus.Fields{
			LOG_PATH:      path,
			LOG_FILE_NAME: DefaultConfigFileName,
			LOG_ERROR:     err,
		}).Warn(LOG_CANNOT_READ_FILE)
		return nil
	}
	data := make(map[interface{}]interface{})
	if err = yaml.Unmarshal(yfile, &data); err != nil {
		logrus.WithFields(logrus.Fields{
			LOG_PATH:      path,
			LOG_FILE_NAME: DefaultConfigFileName,
			LOG_ERROR:     err,
		}).Warn("cannot unmarshal configuration data")
		return nil
	}
	n := createNode(data)
	return n
}

// SafeSubNode returns a specified sub-Node, allowing for nil input
func SafeSubNode(n *Node, key string) *Node {
	if n == nil {
		return nil
	}
	return n.nMap[key]
}

// GetBoolDefault returns a boolean value for a specified key
func GetBoolDefault(n *Node, key string, defaultValue bool) (b bool) {
	b = defaultValue
	if n != nil {
		if value, ok := n.bMap[key]; ok {
			b = value
		} else {
			if value, ok := n.sMap[key]; ok {
				rawValue := InterpretEnvVarReferences(value)
				if cookedValue, e := strconv.ParseBool(rawValue); e == nil {
					b = cookedValue
				}
			}
		}
	}
	return
}

// GetStringDefault returns a string value for a specified key
func GetStringDefault(n *Node, key string, defaultValue string) (s string) {
	s = InterpretEnvVarReferences(defaultValue)
	if n != nil {
		if value, ok := n.sMap[key]; ok {
			s = InterpretEnvVarReferences(value)
		}
	}
	return
}

type Node struct {
	sMap map[string]string
	bMap map[string]bool
	nMap map[string]*Node
}

func createNode(data map[interface{}]interface{}) *Node {
	n := &Node{
		sMap: make(map[string]string),
		nMap: make(map[string]*Node),
		bMap: make(map[string]bool),
	}
	for k, v := range data {
		if key, ok := k.(string); ok {
			switch t := v.(type) {
			case string:
				n.sMap[key] = t
			case bool:
				n.bMap[key] = t
			case map[interface{}]interface{}:
				n.nMap[key] = createNode(t)
			default:
				logrus.WithFields(logrus.Fields{
					"key":   key,
					"value": v,
					"type":  fmt.Sprintf("%T", v),
				}).Warn("unexpected value type found")
				n.sMap[key] = fmt.Sprintf("%v", v)
			}
		}
	}
	return n
}
