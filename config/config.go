package config

import (
	"fmt"
	"reflect"
	"strconv"
	"unicode"
	"unicode/utf8"

	"github.com/mc0239/logm"
	"github.com/mitchellh/mapstructure"
)

// Util is used for retrieving config values from available sources.
// Util should be initialized with config.NewUtil() function
type Util struct {
	configSources []configSource
	Logger        logm.Logm
}

// Bundle is used for filling a user-defined struct with config values.
// Bundle should be initialized with config.NewBundle() function
type Bundle struct {
	prefixKey string
	fields    interface{}
	conf      Util
	Logger    logm.Logm
}

// Options struct is used when instantiating a new Util or Bundle.
type Options struct {
	// Additional configuration source to connect to. Possible values are: "consul"
	Extension string
	// FilePath is a path to configuration file, including the configuration file name.
	// Passing an empty string will default to config/config.yaml
	FilePath string
	// LogLevel can be used to limit the amount of logging output. Default log level is 0. Level 4
	// will only output Warnings and Errors, and level 5 will only output errors.
	// See package github.com/mc0239/logm for more details on logging and log levels.
	LogLevel int
}

type configSource interface {
	Name() string
	ordinal() int
	Get(key string) interface{}
	Subscribe(key string, callback func(key string, value string))
}

// NewUtil instantiates a new Util with given options
func NewUtil(options Options) Util {
	lgr := logm.New("Kumuluz-config")
	lgr.LogLevel = options.LogLevel

	configs := make([]configSource, 0)

	envConfigSource := initEnvConfigSource(&lgr)
	if envConfigSource != nil {
		configs = append(configs, envConfigSource)
	}

	fileConfigSource := initFileConfigSource(options.FilePath, &lgr)
	if fileConfigSource != nil {
		configs = append(configs, fileConfigSource)
	}

	if options.Extension == "consul" {
		extConfigSource := initConsulConfigSource(fileConfigSource, &lgr)
		if extConfigSource != nil {
			configs = append(configs, extConfigSource)
		}
	} else if options.Extension == "etcd" {
		// TODO:
	} else {
		// TODO: invalid ext
	}

	// TODO: sort sources by source.ordinal()

	k := Util{
		configs,
		lgr,
	}

	return k
}

// NewBundle instantiates a new Bundle with given options.
// Fields must be a pointer to a struct that will be filled with configuration values.
// Note that you don't have to preserve the returned Bundle struct, as the configuration is written
// back to the passed fields struct.
func NewBundle(prefixKey string, fields interface{}, options Options) Bundle {
	lgr := logm.New("Kumuluz-config")
	lgr.LogLevel = options.LogLevel

	options.LogLevel = 5
	util := NewUtil(options)

	// convert fields struct to map
	var fieldsMap map[string]interface{}
	err := mapstructure.Decode(fields, &fieldsMap)
	if err != nil {
		panic(err)
	}

	// recursively traverse all fields and set their values using Util.Get()
	if prefixKey != "" {
		prefixKey += "."
	}
	setMapValues(fieldsMap, prefixKey, util)

	// convert map back to struct
	err = mapstructure.Decode(fieldsMap, &fields)
	if err != nil {
		panic(err)
	}

	// TODO: register watches on fields tagged with config:"watch"

	return Bundle{
		prefixKey: "",
		fields:    fields,
		conf:      util,
		Logger:    lgr,
	}
}

// recursively traverse the generated map and assign configuration values using Util
func setMapValues(m map[string]interface{}, keyPrefix string, util Util) {
	for key := range m {
		// if mapstructure tag is not defined, the key is the same as the name of the field.
		// Since exposed struct fields are capitalized, we make the first letter lower-case
		// (capitalized key can be explicitely assigned by using mapstruct tag on field)
		r, n := utf8.DecodeRuneInString(key)
		lkey := string(unicode.ToLower(r)) + key[n:]

		valType := reflect.TypeOf(m[key])
		if valType.Kind() == reflect.Map {
			setMapValues(m[key].(map[string]interface{}), keyPrefix+lkey+".", util)
		} else {
			valFromConf := util.Get(keyPrefix + lkey)
			m[key] = valFromConf
		}
	}
}

// Subscribe creates a watch on a given configuration key.
// Note that watch will be enabled on an extension configuration source, if one has been defined
// when Util was created.
// When value in configuration updated, callback is fired with the key and the new value.
func (c Util) Subscribe(key string, callback func(key string, value string)) {

	// find extension configSource and deploy a watch
	for _, cs := range c.configSources {
		if cs.Name() == "consul" || cs.Name() == "etcd" {
			cs.Subscribe(key, callback)
			break
		}
	}

}

// Get returns the value for a given key, stored in configuration.
// Configuration sources are checked by their ordinal numbers, and value is returned from first
// configuration source it was found in.
func (c Util) Get(key string) interface{} {
	// iterate through configSources and try to get some value ...
	var val interface{}

	for _, cs := range c.configSources {
		val = cs.Get(key)
		if val != nil {
			c.Logger.Verbose(fmt.Sprintf("Found value for key %s, source: %s", key, cs.Name()))
			break
		}
	}
	return val
}

// GetBool is a helper method that calls Util.Get() internally and type asserts the value to
// bool before returning it.
// If value is not found in any configuration source or the value could not be type asserted to
// bool, a false is returned with ok equal to false.
func (c Util) GetBool(key string) (value bool, ok bool) {
	// if value is type asserted as byte array, cast to string and convert to int
	svalue, ok := c.Get(key).([]byte)
	if ok {
		ivalue, err := strconv.ParseBool(string(svalue))
		if err == nil {
			return ivalue, true
		}
	}

	// if value is type asserted as int, return it
	ivalue, ok := c.Get(key).(bool)
	if ok {
		return ivalue, true
	}

	// can't assert to int, return 0
	return false, false
}

// GetInt is a helper method that calls Util.Get() internally and type asserts the value to
// int before returning it.
// If value is not found in any configuration source or the value could not be type asserted to
// int, a zero is returned with ok equal to false.
func (c Util) GetInt(key string) (value int, ok bool) {
	// if value is type asserted as byte array, cast to string and convert to int
	svalue, ok := c.Get(key).([]byte)
	if ok {
		ivalue, err := strconv.Atoi(string(svalue))
		if err == nil {
			return ivalue, true
		}
	}

	// if value is type asserted as int, return it
	ivalue, ok := c.Get(key).(int)
	if ok {
		return ivalue, true
	}

	// can't assert to int, return 0
	return 0, false
}

// GetFloat is a helper method that calls Util.Get() internally and type asserts the value to
// float64 before returning it.
// If value is not found in any configuration source or the value could not be type asserted to
// int, a zero is returned with ok equal to false.
func (c Util) GetFloat(key string) (value float64, ok bool) {
	// if value is type asserted as byte array, cast to string and convert to int
	svalue, ok := c.Get(key).([]byte)
	if ok {
		ivalue, err := strconv.ParseFloat(string(svalue), 64)
		if err == nil {
			return ivalue, true
		}
	}

	// if value is type asserted as int, return it
	ivalue, ok := c.Get(key).(float64)
	if ok {
		return ivalue, true
	}

	// can't assert to int, return 0
	return 0, false
}

// GetString is a helper method that calls Util.Get() internally and type asserts the value to
// string before returning it.
// If value is not found in any configuration source or the value could not be type asserted to
// string, an empty string is returned with ok equal to false.
func (c Util) GetString(key string) (value string, ok bool) {
	// try to type assert as string
	svalue, ok := c.Get(key).(string)
	if ok {
		return svalue, true
	}
	// try to type assert as byte array and cast to string
	bvalue, ok := c.Get(key).([]byte)
	if ok {
		return string(bvalue), true
	}
	// can't assert to string, return nil
	return "", false
}
