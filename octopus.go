package octopus

import (
	"flag"
	"log"
	"path/filepath"
	"strings"

	"github.com/k8s-practice/octopus/config"
	"github.com/k8s-practice/octopus/config/localfile"
	"github.com/k8s-practice/octopus/config/parser/jsonparser"
	"github.com/k8s-practice/octopus/config/parser/tomlparser"
	"github.com/k8s-practice/octopus/config/parser/yamlparser"
)

var configfile = flag.String("c", "./config.toml", "Config file path.")
var configformat = flag.String("f", "", "Config file format.")

// InitConfig initialize base config from local file.
func InitConfig() {
	if *configformat == "" {
		suffix := filepath.Ext(*configfile)
		*configformat = strings.ToLower(strings.TrimPrefix(suffix, "."))
	}

	switch *configformat {
	case tomlparser.Format():
		config.New(config.T().WithScheme(localfile.Scheme()).
			WithPath(*configfile).
			WithFormat(tomlparser.Format()))
	case yamlparser.Format():
		config.New(config.T().WithScheme(localfile.Scheme()).
			WithPath(*configfile).
			WithFormat(tomlparser.Format()))
	case jsonparser.Format():
		config.New(config.T().WithScheme(localfile.Scheme()).
			WithPath(*configfile).
			WithFormat(tomlparser.Format()))
	default:
		log.Panic("Unknown config file format.")
	}
}

// Brain is the brain of octopus.
type Brain struct {
	conf config.Config

	// frameInit is the framework initialize functions slice.
	// It's will be invoked first.
	frameInit []func() error

	// appInit is the application initialize functions slice.
	// It's will be invoded after Run function.
	appInit []func() error
}
