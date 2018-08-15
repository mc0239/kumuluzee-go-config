package config

import (
	"fmt"
	"strconv"
)

// consul value -> local value -> nil

// on init: read all values and store a map with everything
//          make an option to make value watches (update
//			values while app is running)

// order of priority:
// 1 - environment variables
// 2 - configuration files
// 3 - extension (consul, etcd ...)

type ConfigUtil struct {
	configSources []configSource
	Logger        logger
}

var lgr *logger

type configSource interface {
	Name() string
	Get(key string) interface{}
	Watch(key string, callback func(key string, value string))
}

func Initialize(ext string, configPath string, logLevel int) ConfigUtil {

	lgr = &logger{
		LogLevel: logLevel,
	}

	var envConfigSource, fileConfigSource, extConfigSource configSource

	envConfigSource = initEnvConfigSource()
	fileConfigSource = initFileConfigSource(configPath)

	if ext == "consul" {
		extConfigSource = initConsulConfigSource(fileConfigSource)
	} else if ext == "etcd" {
		// TODO:
	} else {
		// TODO: invalid ext
	}

	k := ConfigUtil{
		[]configSource{envConfigSource, extConfigSource, fileConfigSource},
		*lgr,
	}

	return k
}

func (c ConfigUtil) Watch(key string, callback func(key string, value string)) {

	// iterate through configSources and deploy watches
	// note: env and file configSources don't actually have a watch
	for _, cs := range c.configSources {
		cs.Watch(key, callback)
	}

}

/*
Get returns value for a given key, stored in configSource. ConfigSources are iterated by their
importance from most to least important, and the value is returned from the first configSource it
was found in.
Value is of type interface{} and might require type assertion
*/
func (c ConfigUtil) Get(key string) interface{} {
	// iterate through configSources and try to get some value ...
	var val interface{}

	for _, cs := range c.configSources {
		if cs == nil { // TODO: temporary
			continue
		}
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
func (c ConfigUtil) GetString(key string) (value string, ok bool) {
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
func (c ConfigUtil) GetInt(key string) (value int, ok bool) {
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
