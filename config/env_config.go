package config

import "os"

type envConfigSource struct {
}

func initEnvConfigSource() configSource {
	lgr.logV("Initializing EnvConfigSource")
	var c envConfigSource
	lgr.logV("Initialized EnvConfigSource")
	return c
}

func (c envConfigSource) ordinal() int {
	return 300
}

func (c envConfigSource) Get(key string) interface{} {
	value, exists := os.LookupEnv(key)
	if exists {
		return value
	}
	return nil
}

func (c envConfigSource) Subscribe(key string, callback func(key string, value string)) {
	return
}

func (c envConfigSource) Name() string {
	return "Environment"
}
