package config

func getOrDefault(c Util, key string, defaultValue interface{}) interface{} {
	val := c.Get(key)
	if val == nil {
		val = defaultValue
	}
	return val
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
