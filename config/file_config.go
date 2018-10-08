package config

import (
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/ghodss/yaml"
	"github.com/mc0239/logm"
)

type fileConfigSource struct {
	config map[string]interface{}
	logger *logm.Logm
}

func newFileConfigSource(configPath string, lgr *logm.Logm) configSource {
	var c fileConfigSource
	lgr.Verbose("Initializing %s config source", c.Name())
	c.logger = lgr

	var joinedPath string
	if configPath == "" {
		// set default
		joinedPath = "config/config.yaml"
	} else {
		joinedPath = configPath
	}

	lgr.Verbose(fmt.Sprintf("Config file path: %s\n", joinedPath))

	bytes, err := ioutil.ReadFile(joinedPath)
	if err != nil {
		lgr.Error(fmt.Sprintf("Failed to read file on path: %s, error: %s", joinedPath, err.Error()))
		return nil
	}
	//fmt.Printf("Read: %s", bytes)

	err = yaml.Unmarshal(bytes, &c.config)
	if err != nil {
		lgr.Error(fmt.Sprintf("Failed tu unmarshal yaml: %s", err.Error()))
		return nil
	}

	lgr.Verbose("Initialized %s config source", c.Name())
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

func (c fileConfigSource) Subscribe(key string, callback func(key string, value string)) {
	return
}

func (c fileConfigSource) Name() string {
	return "file"
}

func (c fileConfigSource) ordinal() int {
	return 100
}

//
