package config

import (
	"fmt"
	"path"
	"strings"
	"time"

	"github.com/hashicorp/consul/api"
)

type consulConfigSource struct {
	client *api.Client
	path   string
}

func initConsulConfigSource(localConfig configSource) configSource {
	var consulConfig consulConfigSource

	client, err := api.NewClient(api.DefaultConfig())
	if err != nil {
		fmt.Printf("Couldn't create consul client: %s\n", err.Error())
	}

	consulConfig.client = client

	envName := localConfig.Get("kumuluzee.env.name")
	name := localConfig.Get("kumuluzee.name")
	version := localConfig.Get("kumuluzee.version")

	consulConfig.path = fmt.Sprintf("environments/%s/services/%s/%s/config", envName, name, version)
	fmt.Printf("consul KV namespace: %s\n", consulConfig.path)

	return consulConfig
}

func (c consulConfigSource) Get(key string) interface{} {
	//fmt.Println("[consulConfigSource] Get: " + key)
	kv := c.client.KV()

	key = strings.Replace(key, ".", "/", -1)
	//fmt.Printf("KV path: %s\n", path.Join(c.path, key))

	pair, _, err := kv.Get(path.Join(c.path, key), nil)
	if err != nil {
		fmt.Printf("Error getting value: %v\n", err)
		return nil
	}

	//fmt.Printf("Pair received: %v\n", pair)
	if pair == nil {
		return nil
	}
	return pair.Value
}

func (c consulConfigSource) Watch(key string, callback func(key string, value string)) {
	go c.watch(key, "", callback, 0)
}

func (c consulConfigSource) watch(key string, previousValue string, callback func(key string, value string), waitIndex uint64) {
	t, err := time.ParseDuration("10s")
	if err != nil {
		fmt.Printf("Couldn't parse duration for WaitTime: %s\n", err.Error())
		return
	}

	q := api.QueryOptions{
		WaitIndex: waitIndex,
		WaitTime:  t,
	}

	key = strings.Replace(key, ".", "/", -1)
	pair, meta, err := c.client.KV().Get(path.Join(c.path, key), &q)

	fmt.Printf("Key: %s\nPair:\n%v err?: %v\n", key, pair, err)

	if pair != nil {
		if string(pair.Value) != previousValue {
			callback(key, string(pair.Value))
		}
		c.watch(key, string(pair.Value), callback, meta.LastIndex)
	} else {
		if previousValue != "" {
			callback(key, "")
		}
		c.watch(key, "", callback, meta.LastIndex)
	}
}
