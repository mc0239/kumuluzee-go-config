package config

import (
	"fmt"
	"strconv"
)

// Util is used for retrieving config values from available sources. Util can be initialized with
// config.Initialize() function
type Util struct {
	configSources []configSource
	Logger        logger
}

/*type Bundle struct {
	prefixKey string
	fields    interface{}
	conf      Util
}*/

// Options struct is used when instantiating a new Util, it allows configuring extension (currently
// supported: "consul"), path to config file and log level of initialized Util
type Options struct {
	Extension string // supported options: consul
	FilePath  string // path to config.yaml. If not specified, assumes current directory
	LogLevel  int
}

var lgr *logger

type configSource interface {
	Name() string
	ordinal() int
	Get(key string) interface{}
	Subscribe(key string, callback func(key string, value string))
}

func Initialize(options Options) Util {

	lgr = &logger{
		LogLevel: options.LogLevel,
	}

	configs := make([]configSource, 0)

	envConfigSource := initEnvConfigSource()
	if envConfigSource != nil {
		configs = append(configs, envConfigSource)
	}

	fileConfigSource := initFileConfigSource(options.FilePath)
	if fileConfigSource != nil {
		configs = append(configs, fileConfigSource)
	}

	if options.Extension == "consul" {
		extConfigSource := initConsulConfigSource(fileConfigSource)
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
		*lgr,
	}

	return k
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
			lgr.logV(fmt.Sprintf("Found value for key %s, source: %s", key, cs.Name()))
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
