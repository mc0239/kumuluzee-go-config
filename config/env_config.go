package config

import "os"

type envConfigSource struct {
}

func initEnvConfigSource() configSource {
	var c envConfigSource
	return c
}

func (c envConfigSource) Get(key string) interface{} {
	value, exists := os.LookupEnv(key)
	if exists {
		return value
	}
	return nil
}

func (c envConfigSource) Watch(key string, callback func(key string, value string)) {
	return
}
