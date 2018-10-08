package config

import (
	"context"
	"fmt"
	"path"
	"strings"
	"time"

	"github.com/mc0239/logm"

	"go.etcd.io/etcd/client"
)

type etcdConfigSource struct {
	client          *client.Client
	startRetryDelay int64
	maxRetryDelay   int64
	path            string
	logger          *logm.Logm
}

func newEtcdConfigSource(localConfig configSource, lgr *logm.Logm) configSource {
	var etcdConfig etcdConfigSource
	lgr.Verbose("Initializing %s config source", etcdConfig.Name())
	etcdConfig.logger = lgr

	// Initialize etcd client
	clientConfig := client.Config{}

	if etcdAddress := localConfig.Get("kumuluzee.config.etcd.hosts"); etcdAddress != nil {
		clientConfig.Endpoints = []string{etcdAddress.(string)}
	}

	cl, err := client.New(clientConfig)
	if err != nil {
		lgr.Error("Failed to create etcd client: %s", err.Error())
		return nil
	}
	lgr.Info("etcd client address set to %v", clientConfig.Endpoints)
	etcdConfig.client = &cl

	// Load service config from file
	envName := getOrDefault(localConfig, "kumuluzee.env.name", "dev")
	name := getOrDefault(localConfig, "kumuluzee.name", nil)
	version := getOrDefault(localConfig, "kumuluzee.version", "1.0.0")

	if sdl, ok := localConfig.Get("kumuluzee.config.start-retry-delay-ms").(int64); ok {
		etcdConfig.startRetryDelay = sdl
	} else {
		etcdConfig.startRetryDelay = 500
	}

	if mdl, ok := localConfig.Get("kumuluzee.config.max-retry-delay-ms").(int64); ok {
		etcdConfig.maxRetryDelay = mdl
	} else {
		etcdConfig.maxRetryDelay = 900000
	}

	etcdConfig.path = fmt.Sprintf("environments/%s/services/%s/%s/config", envName, name, version)

	lgr.Info("etcd key-value namespace: %s", etcdConfig.path)

	lgr.Verbose("Initialized %s config source", etcdConfig.Name())
	return etcdConfig
}

func (c etcdConfigSource) Get(key string) interface{} {
	kv := client.NewKeysAPI(*c.client)

	key = strings.Replace(key, ".", "/", -1)
	//fmt.Printf("KV path: %s\n", path.Join(c.path, key))

	resp, err := kv.Get(context.Background(), path.Join(c.path, key), nil)
	if err != nil {
		c.logger.Warning("Error getting value: %v", err)
		return nil
	}

	return resp.Node.Value
}

func (c etcdConfigSource) Subscribe(key string, callback func(key string, value string)) {
	c.logger.Info("Creating a watch for key %s, source: %s", key, c.Name())
	go c.watch(key, "", c.startRetryDelay, callback)
}

func (c etcdConfigSource) Name() string {
	return "etcd"
}

func (c etcdConfigSource) ordinal() int {
	return 150
}

//

func (c etcdConfigSource) watch(key string, previousValue string, retryDelay int64, callback func(key string, value string)) {

	// TODO: have a parameter for watch duration, (likely reads from config.yaml?)
	t := 10 * time.Minute
	c.logger.Verbose("Set a watch on key %s with %s wait time", key, t)
	// TODO: where is timeout set????

	key = strings.Replace(key, ".", "/", -1)
	kv := client.NewKeysAPI(*c.client)

	watcher := kv.Watcher(path.Join(c.path, key), nil)

	resp, err := watcher.Next(context.Background())
	if err != nil {
		c.logger.Warning("Watch on %s failed with error: %s, retry delay: %d ms", key, err.Error(), retryDelay)

		// sleep for current delay
		time.Sleep(time.Duration(retryDelay) * time.Millisecond)

		// exponentially extend retry delay, but keep it at most maxRetryDelay
		newRetryDelay := retryDelay * 2
		if newRetryDelay > c.maxRetryDelay {
			newRetryDelay = c.maxRetryDelay
		}
		c.watch(key, "", newRetryDelay, callback)
		return
	}

	c.logger.Verbose("Wait time (%s) on watch for key %s reached.", key, t)

	if string(resp.Node.Value) != previousValue {
		callback(key, string(resp.Node.Value))
	}
	c.watch(key, string(resp.Node.Value), c.startRetryDelay, callback)
}
