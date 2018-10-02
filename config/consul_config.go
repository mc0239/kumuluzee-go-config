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

func newConsulConfigSource(localConfig configSource, lgr *logm.Logm) configSource {
	lgr.Verbose("Initializing ConsulConfigSource")
	var consulConfig consulConfigSource
	consulConfig.logger = lgr

	clientConfig := api.DefaultConfig()

	consulAddress := localConfig.Get("kumuluzee.config.consul.hosts")
	if consulAddress != nil {
		clientConfig.Address = consulAddress.(string)
	}

	client, err := api.NewClient(clientConfig)
	if err != nil {
		lgr.Error("Failed to create Consul client: %s", err.Error())
		return nil
	}

	lgr.Info("Consul client address set to %v", clientConfig.Address)

	consulConfig.client = client

	envName := localConfig.Get("kumuluzee.env.name")
	if envName == nil {
		envName = "dev"
	}
	name := localConfig.Get("kumuluzee.name")
	version := localConfig.Get("kumuluzee.version")
	if version == nil {
		version = "1.0.0"
	}

	startRetryDelay, ok := localConfig.Get("kumuluzee.config.start-retry-delay-ms").(float64)
	if !ok {
		lgr.Warning("Failed to assert value kumuluzee.config.start-retry-delay-ms as float64. Using default value 500.")
		startRetryDelay = 500
	}
	consulConfig.startRetryDelay = int64(startRetryDelay)

	maxRetryDelay, ok := localConfig.Get("kumuluzee.config.max-retry-delay-ms").(float64)
	if !ok {
		lgr.Warning("Failed to assert value kumuluzee.config.max-retry-delay-ms as float64. Using default value 900000.")
		maxRetryDelay = 900000
	}
	consulConfig.maxRetryDelay = int64(maxRetryDelay)

	consulConfig.path = fmt.Sprintf("environments/%s/services/%s/%s/config", envName, name, version)

	lgr.Info("Consul key-value namespace: %s", consulConfig.path)

	lgr.Verbose("Initialized ConsulConfigSource")
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

func (c consulConfigSource) watch(key string, previousValue string, retryDelay int64, callback func(key string, value string), waitIndex uint64) {

	// TODO: have a parameter for watch duration, (likely reads from config.yaml?)
	t := 10 * time.Minute
	c.logger.Verbose("Set a watch on key %s with %s wait time", key, t)

	q := api.QueryOptions{
		WaitIndex: waitIndex,
		WaitTime:  10 * time.Minute,
	}

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

	c.logger.Verbose("Wait time (%s) on watch for key %s reached.", key, t)

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

func (c consulConfigSource) Name() string {
	return "consul"
}
