package config

const CONFIG_PATH = "CONFIG_PATH"

// Configurator gets or sets k/v.
type Configurator interface {
	// Gets value by key, return nil if miss the key.
	Get(key string) interface{}
	// Set value by key, if the key is empty will have no effects.
	Set(key string, value interface{})
}

// Builder creates a Configurator.
type Builder interface {
	Build() (Configurator, error)
	Name() string
}
