package config

import (
	"strings"

	"github.com/k8s-practice/octopus/config/datasource"
	"github.com/k8s-practice/octopus/utils/cast"
)

const (
	DEFAULT_KEY_DELIMITER = "."

	KEY_SCHEME = "scheme"
	KEY_FORMAT = "format"
	KEY_PATH   = "path"
)

var (
	// delim separates the nested key.
	delim string = DEFAULT_KEY_DELIMITER
)

func T() datasource.Target {
	return make(target)
}

func New(t datasource.Target) (Config, error) {
	if ds, err := datasource.Build(t); err != nil {
		return nil, err
	} else {
		return &config{ds: ds}, nil
	}
}

// config implements the interface of Config.
type config struct {
	ds datasource.DataSource
}

// Get gets value by key, it's thread safe.
func (c *config) Get(key string) interface{} {
	if v := c.ds.Get([]string{key}); v != nil {
		return v
	}

	path := strings.Split(key, delim)
	if len(path) == 1 {
		return nil
	}

	if v := c.ds.Get(path); v != nil {
		return v
	}

	return nil
}

// target implements the interface of datasource.Target.
type target map[string]interface{}

func (t target) WithScheme(scheme string) datasource.Target {
	t[KEY_SCHEME] = strings.ToLower(scheme)
	return t
}

func (t target) Scheme() string {
	if v, ok := t[KEY_SCHEME]; ok {
		return cast.ToString(v)
	} else {
		return ""
	}
}

func (t target) WithFormat(format string) datasource.Target {
	t[KEY_FORMAT] = strings.ToLower(format)
	return t
}

func (t target) Format() string {
	if v, ok := t[KEY_FORMAT]; ok {
		return cast.ToString(v)
	} else {
		return ""
	}
}

func (t target) WithPath(path string) datasource.Target {
	t[KEY_PATH] = strings.ToLower(path)
	return t
}

func (t target) Path() string {
	if v, ok := t[KEY_PATH]; ok {
		return cast.ToString(v)
	} else {
		return ""
	}
}

func (t target) WithValue(key string, value interface{}) datasource.Target {
	t[key] = value
	return t
}

func (t target) Value(key string) interface{} {
	return t[key]
}
