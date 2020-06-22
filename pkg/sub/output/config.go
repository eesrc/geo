package output

// Config is a config map containing an output configuration
type Config map[string]interface{}

// GetStringWithDefault gets a string value from output if present, otherwise returns default value
func (config *Config) GetStringWithDefault(key, def string) string {
	if (*config)[key] == nil {
		return def
	}

	v, ok := (*config)[key].(string)

	if !ok {
		return def
	}

	return v
}

// GetIntWithDefault gets a string value from output if present, otherwise returns default value
func (config *Config) GetIntWithDefault(key string, def int64) int64 {
	if (*config)[key] == nil {
		return def
	}

	v, ok := (*config)[key].(int64)

	if !ok {
		return def
	}

	return v
}
