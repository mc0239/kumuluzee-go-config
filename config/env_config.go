package config

import "os"

type envConfigSource struct {
}

func initEnvConfigSource() configSource {
	LogV("Initializing EnvConfigSource")
	var c envConfigSource
	LogV("Initialized EnvConfigSource")
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

func (c envConfigSource) Name() string {
	return "Environment"
}
