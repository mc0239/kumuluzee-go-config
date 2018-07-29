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
	var c fileConfigSource

	joinedPath := path.Join(configPath, "config.yaml")

	bytes, err := ioutil.ReadFile(joinedPath)
	if err != nil {
		fmt.Printf("Failed to read file %s, error: %s\n", joinedPath, err.Error())
		return nil
	}
	//fmt.Printf("Read: %s", bytes)

	err = yaml.Unmarshal(bytes, &c.config)
	if err != nil {
		fmt.Printf("Failed tu unmarshal yaml: %s\n", err.Error())
		return nil
	}

	return c
}

func (c fileConfigSource) Get(key string) interface{} {
	//fmt.Println("[fileConfigSource] Get: " + key)
	tree := strings.Split(key, ".")

	// move deeper into maps for every dot delimiter
	val := c.config
	for i := 0; i < len(tree)-1; i++ {
		if val == nil {
			return nil
		}
		val = val[tree[i]].(map[string]interface{})
		//fmt.Printf("%d ::: %v\n", i, val)
	}

	return val[tree[len(tree)-1]]
}

func (c fileConfigSource) Watch(key string, callback func(key string, value string)) {
	return
}
