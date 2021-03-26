package config

type multiConfig struct {
	allConfig []Config
}

// MultiConfig combines configurations.
// Config order is the priority of each configurations.
func MultiConfig(configSlice ...Config) Config {
	allConfig := make([]Config, 0, len(configSlice))
	for _, c := range configSlice {
		if mc, ok := c.(*multiConfig); ok {
			allConfig = append(allConfig, mc.allConfig...)
		} else {
			allConfig = append(allConfig, c)
		}
	}

	return &multiConfig{allConfig}
}

func (mc *multiConfig) Get(key string) interface{} {
	for _, c := range mc.allConfig {
		if i := c.Get(key); i != nil {
			return i
		}
	}

	return nil
}
