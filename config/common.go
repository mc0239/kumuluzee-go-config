package config

func loadServiceConfiguration(conf Util) (envName, name, version string, startRD, maxRD int64) {
	if e, ok := conf.GetString("kumuluzee.env.name"); ok {
		envName = e
	} else {
		envName = "dev"
	}

	name, _ = conf.GetString("kumuluzee.name")

	if v, ok := conf.GetString("kumuluzee.version"); ok {
		version = v
	} else {
		version = "1.0.0"
	}

	if sdl, ok := conf.GetInt("kumuluzee.config.start-retry-delay-ms"); ok {
		startRD = int64(sdl)
	} else {
		startRD = 500
	}

	if mdl, ok := conf.GetInt("kumuluzee.config.max-retry-delay-ms"); ok {
		maxRD = int64(mdl)
	} else {
		maxRD = 900000
	}

	return
}

func assertAsNumber(val interface{}) (num float64, ok bool) {
	switch t := val.(type) {
	case int:
		return float64(t), true
	case int8:
		return float64(t), true
	case int16:
		return float64(t), true
	case int32:
		return float64(t), true
	case int64:
		return float64(t), true
	case uint:
		return float64(t), true
	case uint8:
		return float64(t), true
	case uint16:
		return float64(t), true
	case uint32:
		return float64(t), true
	case uint64:
		return float64(t), true

	case float32:
		return float64(t), true
	case float64:
		return float64(t), true
	default:
		return 0, false
	}
}
