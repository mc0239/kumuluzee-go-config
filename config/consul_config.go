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
	lgr.logV("Initializing ConsulConfigSource")
	var consulConfig consulConfigSource

	clientConfig := api.DefaultConfig()

	consulAddress := localConfig.Get("kumuluzee.config.consul.hosts")
	if consulAddress != nil {
		clientConfig.Address = consulAddress.(string)
	}

	client, err := api.NewClient(clientConfig)
	if err != nil {
		lgr.logE(fmt.Sprintf("Failed to create Consul client: %s", err.Error()))
		return nil
	}

	lgr.logI(fmt.Sprintf("Consul client address set to %v", clientConfig.Address))

	consulConfig.client = client

	envName := localConfig.Get("kumuluzee.env.name")
	name := localConfig.Get("kumuluzee.name")
	version := localConfig.Get("kumuluzee.version")

	consulConfig.path = fmt.Sprintf("environments/%s/services/%s/%s/config", envName, name, version)

	lgr.logI(fmt.Sprintf("Consul key-value namespace: %s", consulConfig.path))

	lgr.logV("Initialized ConsulConfigSource")
	return consulConfig
}

func (c consulConfigSource) ordinal() int {
	return 150
}

func (c consulConfigSource) Get(key string) interface{} {
	//fmt.Println("[consulConfigSource] Get: " + key)
	kv := c.client.KV()

	key = strings.Replace(key, ".", "/", -1)
	//fmt.Printf("KV path: %s\n", path.Join(c.path, key))

	pair, _, err := kv.Get(path.Join(c.path, key), nil)
	if err != nil {
		lgr.logW(fmt.Sprintf("Error getting value: %v", err))
		return nil
	}

	//fmt.Printf("Pair received: %v\n", pair)
	if pair == nil {
		return nil
	}
	return pair.Value
}

func (c consulConfigSource) Subscribe(key string, callback func(key string, value string)) {
	lgr.logI(fmt.Sprintf("Creating a watch for key %s, source: %s", key, c.Name()))
	go c.watch(key, "", callback, 0)
}

func (c consulConfigSource) watch(key string, previousValue string, callback func(key string, value string), waitIndex uint64) {

	// TODO: have a parameter for watch duration, (likely reads from config.yaml?)
	t, err := time.ParseDuration("10m")
	if err != nil {
		lgr.logW(fmt.Sprintf("Failed to parse duration for WaitTime: %s, using default value: %s", err.Error(), t))
		return
	}

	lgr.logV(fmt.Sprintf("Set a watch on key %s with %s wait time", key, t))

	q := api.QueryOptions{
		WaitIndex: waitIndex,
		WaitTime:  t,
	}

	key = strings.Replace(key, ".", "/", -1)
	pair, meta, err := c.client.KV().Get(path.Join(c.path, key), &q)

	//fmt.Printf("Key: %s\nPair:\n%v err?: %v\n", key, pair, err)
	lgr.logV(fmt.Sprintf("Watch on key %s hit %s wait time", key, t))

	if pair != nil {
		if string(pair.Value) != previousValue {
			callback(key, string(pair.Value))
		}
		c.watch(key, string(pair.Value), callback, meta.LastIndex)
	} else {
		if previousValue != "" {
			callback(key, "")
		}
		var lastIndex uint64
		if meta != nil {
			lastIndex = meta.LastIndex
		}
		c.watch(key, "", callback, lastIndex)
	}
}

func (c consulConfigSource) Name() string {
	return "consul"
}
