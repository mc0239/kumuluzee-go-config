package config

import (
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
	// ConfigPath is a path to configuration file, including the configuration file name.
	// Passing an empty string will default to config/config.yaml
	ConfigPath string
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

	fileConfigSource := initFileConfigSource(options.ConfigPath, &lgr)
	if fileConfigSource != nil {
		configs = append(configs, fileConfigSource)
	} else {
		lgr.Error("File configuration source failed to load!")
	}

	switch options.Extension {
	case "consul":
		extConfigSource := initConsulConfigSource(fileConfigSource, &lgr)
		if extConfigSource != nil {
			configs = append(configs, extConfigSource)
		}
		break
	case "etcd":
		// TODO: implement etcd extension
		break
	case "":
		// no extension
		break
	default:
		lgr.Error("Invalid extension specified, extension configuration source will not be available")
		break
	}

	// insertion sort
	for i := 1; i < len(configs); i++ {
		for k := i; k > 0 && configs[k].ordinal() > configs[k-1].ordinal(); k-- {
			// swap
			temp := configs[k]
			configs[k] = configs[k-1]
			configs[k-1] = temp
		}
	}

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

	util := NewUtil(options)

	bun := Bundle{
		prefixKey: prefixKey,
		fields:    &fields,
		conf:      util,
		Logger:    lgr,
	}

	// convert fields struct to map
	var fieldsMap map[string]interface{}

	decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		Result:     &fieldsMap,
		TagName:    "config",
		ZeroFields: false,
	})
	if err != nil {
		lgr.Error("Failed to make a new decoder: %s", err.Error())
		return bun
	}

	err = decoder.Decode(fields)
	if err != nil {
		lgr.Error("Failed to decode fields: %s", err.Error())
		return bun
	}

	// recursively traverse all fields and set their values using Util.Get()
	if prefixKey != "" {
		prefixKey += "."
	}
	traverseFillBundle(fieldsMap, prefixKey, util)

	// convert map back to struct
	recoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		Result:     &fields,
		TagName:    "config",
		ZeroFields: false,
	})
	if err != nil {
		lgr.Error("Failed to make a new decoder: %s", err.Error())
		return bun
	}

	err = recoder.Decode(fieldsMap)
	if err != nil {
		lgr.Error("Failed to recode fieldsMap: %s", err.Error())
		return bun
	}

	// TODO: register watches on fields tagged with config:"watch"

	return bun
}

// recursively traverse the generated map and assign configuration values using Util
func traverseFillBundle(m map[string]interface{}, keyPrefix string, util Util) {
	for key := range m {
		// if config tag is not defined, the key is the same as the name of the field.
		// Since exposed struct fields are capitalized, we make the first letter lower-case
		// (capitalized key can be explicitely assigned by using config tag on field)
		r, n := utf8.DecodeRuneInString(key)
		lkey := string(unicode.ToLower(r)) + key[n:]

		valType := reflect.TypeOf(m[key])
		if valType.Kind() == reflect.Map {
			traverseFillBundle(m[key].(map[string]interface{}), keyPrefix+lkey+".", util)
		} else {
			switch t := m[key].(type) {
			case bool:
				if val, ok := util.GetBool(keyPrefix + lkey); ok {
					m[key] = val
				}
				break
			case float32, float64:
				if val, ok := util.GetFloat(keyPrefix + lkey); ok {
					m[key] = val
				}
				break
			case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
				if val, ok := util.GetInt(keyPrefix + lkey); ok {
					m[key] = val
				}
				break
			case string:
				if val, ok := util.GetString(keyPrefix + lkey); ok {
					m[key] = val
				}
				break
			default:
				util.Logger.Warning("Unexpected type when bundling: %T", t)
				break
			}
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
	rvalue := c.Get(key)

	bvalue, ok := c.Get(key).(bool)
	if ok {
		return bvalue, true
	}

	svalue, ok := rvalue.(string)
	if ok {
		bvalue, err := strconv.ParseBool(string(svalue))
		if err == nil {
			return bvalue, true
		}
	}

	return false, false
}

// GetInt is a helper method that calls Util.Get() internally and type asserts the value to
// int before returning it.
// If value is not found in any configuration source or the value could not be type asserted to
// int, a zero is returned with ok equal to false.
func (c Util) GetInt(key string) (value int, ok bool) {
	rvalue := c.Get(key)

	// try to assert as any number type
	nvalue, ok := assertAsNumber(rvalue)
	if ok {
		return int(nvalue), true
	}

	// try to assert as string and convert to int
	svalue, ok := rvalue.(string)
	if ok {
		ivalue64, err := strconv.ParseInt(svalue, 0, 64)
		if err == nil {
			return int(ivalue64), true
		}
		fvalue64, err := strconv.ParseFloat(svalue, 64)
		if err == nil {
			return int(fvalue64), true
		}
	}

	return 0, false
}

// GetFloat is a helper method that calls Util.Get() internally and type asserts the value to
// float64 before returning it.
// If value is not found in any configuration source or the value could not be type asserted to
// float64, a zero is returned with ok equal to false.
func (c Util) GetFloat(key string) (value float64, ok bool) {
	rvalue := c.Get(key)

	// try to assert as any number type
	nvalue, ok := assertAsNumber(rvalue)
	if ok {
		return nvalue, true
	}

	// try to assert as string and convert to float64
	svalue, ok := rvalue.(string)
	if ok {
		fvalue64, err := strconv.ParseFloat(svalue, 64)
		if err == nil {
			return fvalue64, true
		}
	}

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

	// can't assert to string, return nil
	return "", false
}

func assertAsNumber(val interface{}) (num float64, ok bool) {
	switch t := val.(type) {
	case int:
		return float64(t), true
	case int8:
		return float64(t), true
	case int16:
		return float64(t), true
	case int32:
		return float64(t), true
	case int64:
		return float64(t), true
	case uint:
		return float64(t), true
	case uint8:
		return float64(t), true
	case uint16:
		return float64(t), true
	case uint32:
		return float64(t), true
	case uint64:
		return float64(t), true

	case float32:
		return float64(t), true
	case float64:
		return float64(t), true
	default:
		return 0, false
	}
}
