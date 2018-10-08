package config

import (
	"fmt"
	"path"
	"strings"
	"time"

	"github.com/mc0239/logm"

	"github.com/hashicorp/consul/api"
)

type consulConfigSource struct {
	client          *api.Client
	startRetryDelay int64
	maxRetryDelay   int64
	path            string
	logger          *logm.Logm
}

func newConsulConfigSource(conf Util, lgr *logm.Logm) configSource {
	var consulConfig consulConfigSource
	lgr.Verbose("Initializing %s config source", consulConfig.Name())
	consulConfig.logger = lgr

	// Initialize consul client
	clientConfig := api.DefaultConfig()

	if consulAddress, ok := conf.GetString("kumuluzee.config.consul.hosts"); ok {
		clientConfig.Address = consulAddress
	}

	client, err := api.NewClient(clientConfig)
	if err != nil {
		lgr.Error("Failed to create Consul client: %s", err.Error())
		return nil
	}
	lgr.Info("Consul client address set to %v", clientConfig.Address)
	consulConfig.client = client

	// Load service config from file
	envName := getOrDefault(conf, "kumuluzee.env.name", "dev")
	name := getOrDefault(conf, "kumuluzee.name", nil)
	version := getOrDefault(conf, "kumuluzee.version", "1.0.0")

	if sdl, ok := conf.GetInt("kumuluzee.config.start-retry-delay-ms"); ok {
		consulConfig.startRetryDelay = int64(sdl)
	} else {
		consulConfig.startRetryDelay = 500
	}

	if mdl, ok := conf.GetInt("kumuluzee.config.max-retry-delay-ms"); ok {
		consulConfig.maxRetryDelay = int64(mdl)
	} else {
		consulConfig.maxRetryDelay = 900000
	}

	lgr.Verbose("start-retry-delay-ms=%d, max-retry-delay-ms=%d", consulConfig.startRetryDelay, consulConfig.maxRetryDelay)

	consulConfig.path = fmt.Sprintf("environments/%s/services/%s/%s/config", envName, name, version)

	//
	lgr.Info("%s key-value namespace: %s", consulConfig.Name(), consulConfig.path)
	lgr.Verbose("Initialized %s config source", consulConfig.Name())
	return consulConfig
}

func (c consulConfigSource) Get(key string) interface{} {
	//fmt.Println("[consulConfigSource] Get: " + key)
	kv := c.client.KV()

	key = strings.Replace(key, ".", "/", -1)
	//fmt.Printf("KV path: %s\n", path.Join(c.path, key))

	pair, _, err := kv.Get(path.Join(c.path, key), nil)
	if err != nil {
		c.logger.Warning("Error getting value: %v", err)
		return nil
	}

	//fmt.Printf("Pair received: %v\n", pair)
	if pair == nil {
		return nil
	}
	// pair.Value is type []byte
	return string(pair.Value)
}

func (c consulConfigSource) Subscribe(key string, callback func(key string, value string)) {
	c.logger.Info("Creating a watch for key %s, source: %s", key, c.Name())
	go c.watch(key, "", c.startRetryDelay, callback, 0)
}

func (c consulConfigSource) Name() string {
	return "consul"
}

func (c consulConfigSource) ordinal() int {
	return 150
}

//

func (c consulConfigSource) watch(key string, previousValue string, retryDelay int64, callback func(key string, value string), waitIndex uint64) {

	q := api.QueryOptions{
		WaitIndex: waitIndex,
		WaitTime:  10 * time.Minute,
	}

	c.logger.Verbose("Setting a watch on key %s with %s wait time", key, q.WaitTime)
	key = strings.Replace(key, ".", "/", -1)
	pair, meta, err := c.client.KV().Get(path.Join(c.path, key), &q)

	if err != nil {
		c.logger.Warning("Watch on %s failed with error: %s, retry delay: %d ms", key, err.Error(), retryDelay)

		// sleep for current delay
		time.Sleep(time.Duration(retryDelay) * time.Millisecond)

		// exponentially extend retry delay, but keep it at most maxRetryDelay
		newRetryDelay := retryDelay * 2
		if newRetryDelay > c.maxRetryDelay {
			newRetryDelay = c.maxRetryDelay
		}
		c.watch(key, "", newRetryDelay, callback, 0)
		return
	}

	c.logger.Verbose("Wait time (%s) on watch for key %s reached.", q.WaitTime, key)

	if pair != nil {
		if string(pair.Value) != previousValue {
			callback(key, string(pair.Value))
		}
		c.watch(key, string(pair.Value), c.startRetryDelay, callback, meta.LastIndex)
	} else {
		if previousValue != "" {
			callback(key, "")
		}
		var lastIndex uint64
		if meta != nil {
			lastIndex = meta.LastIndex
		}
		c.watch(key, "", c.startRetryDelay, callback, lastIndex)
	}
}
