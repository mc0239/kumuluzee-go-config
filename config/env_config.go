package config

import (
	"os"
	"regexp"
	"strings"

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

	for _, keyName := range getPossibleNames(key) {
		value, exists := os.LookupEnv(keyName)
		if exists {
			return value
		}
	}

	return nil
}

func (c envConfigSource) Subscribe(key string, callback func(key string, value string)) {
	return
}

func (c envConfigSource) Name() string {
	return "Environment"
}

// https://github.com/kumuluz/kumuluzee/blob/master/common/src/main/java/com/kumuluz/ee/configuration/sources/EnvironmentConfigurationSource.java#L224
func getPossibleNames(key string) []string {
	possibleNames := []string{
		// MP Config 1.3: raw key
		key,
		normalizeKey(key),
		normalizeKeyUpper(key),
		parseKeyLegacy1(key),
		parseKeyLegacy2(key),
	}

	return possibleNames
}

// MP Config 1.3: replaces non alpha-numeric characters with '_'
func normalizeKey(key string) string {
	re1 := regexp.MustCompile("[^a-zA-Z0-9]")
	normKey := re1.ReplaceAllString(key, "_")
	return normKey
}

func normalizeKeyUpper(key string) string {
	return strings.ToUpper(normalizeKey(key))
}

// legacy 1: removes characters '[]-' and replaces dots with '_', to uppercase
func parseKeyLegacy1(key string) string {
	return strings.ToUpper(
		strings.Replace(strings.Replace(strings.Replace(strings.Replace(
			key,
			"[", "", -1),
			"]", "", -1),
			"-", "", -1),
			".", "_", -1))
}

// legacy 2: replaces dots with '_', to uppercase
func parseKeyLegacy2(key string) string {
	return strings.ToUpper(strings.Replace(key, ".", "_", -1))
}
