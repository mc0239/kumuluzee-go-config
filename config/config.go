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

// Util is used for retrieving config values from available sources. Util can be initialized with
// config.Initialize() function
type Util struct {
	configSources []configSource
	Logger        logm.Logm
}

type Bundle struct {
	prefixKey string
	fields    interface{}
	conf      Util
	Logger    logm.Logm
}

// Options struct is used when instantiating a new Util, it allows configuring extension (currently
// supported: "consul"), path to config file and log level of initialized Util
type Options struct {
	Extension string // supported options: consul
	FilePath  string // path to config.yaml. If not specified, assumes current directory
	LogLevel  int
}

type configSource interface {
	Name() string
	ordinal() int
	Get(key string) interface{}
	Subscribe(key string, callback func(key string, value string))
}

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

func (c Util) Subscribe(key string, callback func(key string, value string)) {

	// find extension configSource and deploy a watch
	for _, cs := range c.configSources {
		if cs.Name() == "consul" || cs.Name() == "etcd" {
			cs.Subscribe(key, callback)
			break
		}
	}

}

/*
Get returns value for a given key, stored in configSource. ConfigSources are iterated by their
importance from most to least important, and the value is returned from the first configSource it
was found in.
Value is of type interface{} and might require type assertion
*/
func (c Util) Get(key string) interface{} {
	// iterate through configSources and try to get some value ...
	var val interface{}

	for _, cs := range c.configSources {
		val = cs.Get(key)
		if val != nil {
			c.Logger.LogV(fmt.Sprintf("Found value for key %s, source: %s", key, cs.Name()))
			break
		}
	}
	return val
}

/*
GetString calls Config.Get() function to retrieve the value and tries to type assert or type
cast the value to type string.
*/
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

/*
GetInt calls Config.Get() function to retrieve the value and tries to type assert or type
cast the value to type int.
*/
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
