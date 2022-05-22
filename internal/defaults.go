package internal

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

const (
	defaultConfigFileBaseName = "defaults"
	DefaultConfigFileName     = defaultConfigFileBaseName + ".yaml"
)

func ReadDefaultsYaml(path string) (v *viper.Viper) {
	v = viper.New()
	v.SetConfigName(defaultConfigFileBaseName)
	v.AddConfigPath(path)
	if err := v.ReadInConfig(); err != nil {
		logrus.WithFields(logrus.Fields{
			LOG_DIRECTORY: path,
			LOG_FILE_NAME: DefaultConfigFileName,
			LOG_ERROR:     err,
		}).Warn("error reading defaults configuration file")
		fmt.Fprintf(
			os.Stderr,
			"There was an error reading the configuration file %q: %v\n",
			filepath.Join(path, DefaultConfigFileName),
			err)
		v = nil
	}
	return
}

func SafeSubViper(v *viper.Viper, key string) *viper.Viper {
	if v == nil {
		return nil
	}
	return v.Sub(key)
}

func GetBoolDefault(v *viper.Viper, key string, defaultValue bool) (b bool) {
	b = defaultValue
	if v != nil && v.IsSet(key) {
		rawValue := v.GetString(key)
		if cookedValue, e := strconv.ParseBool(rawValue); e == nil {
			b = cookedValue
		}
	}
	return
}

func GetStringDefault(v *viper.Viper, key string, defaultValue string) (s string) {
	s = defaultValue
	if v != nil && v.IsSet(key) {
		s = v.GetString(key)
	}
	return
}
