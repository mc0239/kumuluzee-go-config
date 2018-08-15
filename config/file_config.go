package config

import (
	"fmt"
	"io/ioutil"
	"path"
	"strings"

	"github.com/ghodss/yaml"
)

type fileConfigSource struct {
	config map[string]interface{}
}

func initFileConfigSource(configPath string) configSource {
	lgr.logV("Initializing FileConfigSource")
	var c fileConfigSource

	joinedPath := path.Join(configPath, "config.yaml")

	lgr.logV(fmt.Sprintf("Config file path: %s\n", joinedPath))

	bytes, err := ioutil.ReadFile(joinedPath)
	if err != nil {
		lgr.logE(fmt.Sprintf("Failed to read file on path: %s, error: %s", joinedPath, err.Error()))
		return nil
	}
	//fmt.Printf("Read: %s", bytes)

	err = yaml.Unmarshal(bytes, &c.config)
	if err != nil {
		lgr.logE(fmt.Sprintf("Failed tu unmarshal yaml: %s", err.Error()))
		return nil
	}

	lgr.logV("Initialized FileConfigSource")
	return c
}

func (c fileConfigSource) Get(key string) interface{} {
	//fmt.Println("[fileConfigSource] Get: " + key)
	tree := strings.Split(key, ".")

	// move deeper into maps for every dot delimiter
	val := c.config
	var assertOk bool
	for i := 0; i < len(tree)-1; i++ {
		if val == nil {
			return nil
		}
		val, assertOk = val[tree[i]].(map[string]interface{})

		if !assertOk {
			return nil
		}
		//fmt.Printf("%d ::: %v\n", i, val)
	}

	return val[tree[len(tree)-1]]
}

func (c fileConfigSource) Watch(key string, callback func(key string, value string)) {
	return
}

func (c fileConfigSource) Name() string {
	return "File"
}
