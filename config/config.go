package config

import (
	"time"

	"github.com/k8s-practice/octopus/utils/cast"
	// To register jsonparser
	//_ "github.com/k8s-practice/octopus/config/parser/jsonparser"
	// To register tomlparser
	//_ "github.com/k8s-practice/octopus/config/parser/tomlparser"
	// To register yamlparser
	//_ "github.com/k8s-practice/octopus/config/parser/yamlparser"
	// To register localfile datasource
	//_ "github.com/k8s-practice/octopus/config/localfile"
)

// NOTE: All functions must be thread safe.
type Config interface {
	// Get gets value by key.
	Get(key string) interface{}
}

func Get(c Config, key string) interface{} {
	return c.Get(key)
}

func GetBool(c Config, key string) bool {
	return cast.ToBool(c.Get(key))
}

func GetInt(c Config, key string) int {
	return cast.ToInt(c.Get(key))
}

func GetInt32(c Config, key string) int32 {
	return cast.ToInt32(c.Get(key))
}

func GetInt64(c Config, key string) int64 {
	return cast.ToInt64(c.Get(key))
}

func GetIntSlice(c Config, key string) []int {
	return cast.ToIntSlice(c.Get(key))
}

func GetUint(c Config, key string) uint {
	return cast.ToUint(c.Get(key))
}

func GetUint32(c Config, key string) uint32 {
	return cast.ToUint32(c.Get(key))
}

func GetUint64(c Config, key string) uint64 {
	return cast.ToUint64(c.Get(key))
}

func GetFloat32(c Config, key string) float32 {
	return cast.ToFloat32(c.Get(key))
}

func GetFloat64(c Config, key string) float64 {
	return cast.ToFloat64(c.Get(key))
}

func GetString(c Config, key string) string {
	return cast.ToString(c.Get(key))
}

func GetStringSlice(c Config, key string) []string {
	return cast.ToStringSlice(c.Get(key))
}

func GetTime(c Config, key string) time.Time {
	return cast.ToTime(c.Get(key))
}

func GetDuration(c Config, key string) time.Duration {
	return cast.ToDuration(c.Get(key))
}
