# KumuluzEE Go Config

Note: crossed content is work in progress.

KumuluzEE Go Config is an open-source configuration management for the KumuluzEE framework. It is a Go package based on [KumuluzEE Config](https://github.com/kumuluz/kumuluzee-config), configuration management library developed for microservices written in Java programming language. It extends basic configuration framework described [here](https://github.com/kumuluz/kumuluzee/wiki/Configuration).

Package provides support for [environment variables](https://github.com/kumuluz/kumuluzee/wiki/Configuration#environment-variables) and [configuration files](https://github.com/kumuluz/kumuluzee/wiki/Configuration#configuration-files) as well as for additional configuration sources Consul <s>and etcd</s>.

KumuluzEE Node.js Config follows the idea of an unified configuration API for the framework and provides additional configuration sources which can be utilised with a standard KumuluzEE configuration interface.

## Install

You can `go get` this package:

```
$ go get github.com/mc0239/kumuluzee-go-config
```

Or you can get it via a package manager, for example `dep`:

```
$ dep ensure -add github.com/mc0239/kumuluzee-go-config
```

## Setup
In order to connect to Consul and etcd, you must properly set configuration files. For more information check sections **Configuring Consul** and **Configuring etcd**  in [KumuluzEE Config's section Usage](https://github.com/kumuluz/kumuluzee-config#usage).

Properties in Consul and etcd are stored in a specific matter. For more information check sections  **Configuration properties inside Consul** and **Configuration properties inside etcd** in [KumuluzEE Config's section Usage](https://github.com/kumuluz/kumuluzee-config#usage).


**Configuration source priorities**

Each configuration source has its own priority, meaning values from configuration sources with lower priories can be overwritten with values from higher. Properties from configuration files has the lowest priority, which can be overwritten with properties from additional configuration sources Consul and etcd, while properties defined with environmental variables have the highest priority.

## Usage

Properties can be held in a struct using `config.Bundle` or retrieved by using `config.Util` methods.

**config.NewBundle(prefixKey, fields, options)**

**prefixKey** (string): value represents the prefix key for the configuration property keys use "" (empty string) for no prefix.

**fields** (struct pointer): struct that will be populated with configuration properties. Fields in the struct that will be populated must be exported (starting with an upper-case letter). By default, configuration key is equal to field name, but with first letter lower-cased. Fields can use custom key names by specifying `mapstructure` tag.

Fields can be assigned watches by adding a tag `config:"watch"`.

**options** (config.Options): can be used to set an additional configuration source (consul <s>or etcd</s>) or custom configuration file path.

```go
// import package
import "github.com/mc0239/kumuluzee-go-config/config"

// define a struct
type myConfig struct {
    Kumuluz struct {
        Name    string
        Version string
        Env     struct {
            Name string
        }
    } `mapstructure:"kumuluzee"`
    RestConfig struct {
        String  string `mapstructure:"string-property" config:"watch"`
        Boolean bool   `mapstructure:"boolean-property"`
        Integer int    `mapstructure:"integer-property"`
    } `mapstructure:"rest-config"`
}

// make a struct instance & call config.NewBundle with a pointer to it
var myconf myConfig
config.NewBundle("", &myconf, config.Options{})
```

**config.Util**

It is used for retrieving values of configuration parameters from the configuration framework.

***config.NewUtil(options)***

**options** (config.Options): can be used to set an additional configuration source (consul <s>or etcd</s>) or custom configuration file path.

```go
// import package
import "github.com/mc0239/kumuluzee-go-config/config"

// usage
var confUtil config.Util

confUtil = config.Initialize(config.Options{
    Extension: "consul",
})
```

***.Get(key)***

Returns value of a given key. <s>Returned value is a Promise, so you need to `await` for response.</s>
Returned value is of type `interface{}` and should be type asserted before further use. If key does not exist, returned value will be `nil`.

```go
property := confUtil.Get("some-property")
```

There are additional functions available for getting a specific type:

```go
value, ok := confUtil.GetBool(key) // bool
value, ok := confUtil.GetInt(key) // int
value, ok := confUtil.GetFloat(key) // float64
value, ok := confUtil.GetString(key) // string
```

Variable `ok` will evaluate to `true` if key exists and value is successfully type asserted.

**Watches**

Since configuration properties in <s>etcd and</s> Consul can be updated during microservice runtime, they have to be dynamically updated inside the running microservices. This behaviour can be enabled with watches.

If watch is enabled on a field, its value will be dynamically updated on any change in configuration source, as long as new value is of a proper type. For example, if value in configuration store is set to `'string'` type and is changed to a non-string value, field value will not be updated.

While properties can be watched using config.Bundle object by tagging a field with `config:"watch"`, we can use config.Util to subscribe for changes using `subscribe` function.

```go
confUtil.Subscribe(watchKey, func(key string, value string) {
    fmt.Printf("New value for key %s is %s\n", key, value)
})
```

**Retry delays**

<s>Etcd and</s> Consul implementations support retry delays on watch connection errors. Since they use increasing exponential delay, two parameters need to be specified:

* `kumuluzee.config.start-retry-delay-ms`, which sets the retry delay duration in ms on first error - default: 500
* `kumuluzee.config.max-retry-delay-ms`, which sets the maximum delay duration in ms on consecutive errors - default: 900000 (15 min)

<s>

## Changelog

Recent changes can be viewed on Github on the [Releases Page](https://github.com/kumuluz/kumuluzee/releases)

## Contribute

See the [contributing docs](https://github.com/kumuluz/kumuluzee-nodejs-config/blob/master/CONTRIBUTING.md)

When submitting an issue, please follow the [guidelines](https://github.com/kumuluz/kumuluzee-nodejs-config/blob/master/CONTRIBUTING.md#bugs).

When submitting a bugfix, write a test that exposes the bug and fails before applying your fix. Submit the test alongside the fix.

When submitting a new feature, add tests that cover the feature.
</s>

## License

MIT

