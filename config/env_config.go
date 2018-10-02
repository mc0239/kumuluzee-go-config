package config

import (
	"os"

	"github.com/mc0239/logm"
)

type envConfigSource struct {
}

func newEnvConfigSource(lgr *logm.Logm) configSource {
	lgr.Verbose("Initializing EnvConfigSource")
	var c envConfigSource
	lgr.Verbose("Initialized EnvConfigSource")
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
