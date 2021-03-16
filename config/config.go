package config

import (
	"log"
	// NOTE: Import these packages while using them, for reducing program size.
	// To register jsonparser
	// _ "github.com/k8s-practice/octopus/config/parser/jsonparser"
	// To register tomlparser
	// _ "github.com/k8s-practice/octopus/config/parser/tomlparser"
	// To register yamlparser
	// _ "github.com/k8s-practice/octopus/config/parser/yamlparser"
	// To register localfile datasource
	// _ "github.com/k8s-practice/octopus/config/localfile"
)

type Config interface {
	Get(key string) interface{}
	Set(key string, value interface{})
}

func T() Target {
	return make(target)
}

func New(targets ...Target) Config {
	c := &config{
		delimiter: DEFAULT_KEY_DELIMITER,
	}

	for _, t := range targets {
		if builder, ok := SearchBuilder(t.Scheme()); ok {
			if ds, err := builder.Build(t); err != nil {
				log.Panic(err)
			} else {
				c.addDataSource(ds)
			}
		} else {
			log.Panicf("Unsupported config scheme [%s].", t.Scheme())
		}
	}

	return c
}
