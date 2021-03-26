package config

import (
	"log"
	"strings"
	"sync/atomic"

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

func New(targets ...datasource.Target) Config {
	c := &config{}

	for _, t := range targets {
		if err := c.AddTarget(t); err != nil {
			log.Panic(err)
		}
	}

	return c
}

// config implements the interface of Config.
type config struct {
	// datasource contains impl of the datasource.DataSource.
	datasource atomic.Value
}

// Get gets value by key, it's thread safe.
func (c *config) Get(key string) interface{} {
	// Search in the remote datasource if it exists.
	if v := c.datasource.Load(); v != nil {
		if i := v.(datasource.DataSource).Get([]string{key}); i != nil {
			return i
		}

		path := strings.Split(key, delim)
		if i := v.(datasource.DataSource).Get(path); i != nil {
			return i
		}
	}

	return nil
}

func (c *config) AddTarget(t datasource.Target) error {
	ds, err := datasource.Build(t)
	if err == nil {
		c.datasource.Store(ds)
	}

	return err
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
