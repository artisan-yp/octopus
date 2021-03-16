package test

import (
	"log"
	"testing"

	// NOTE: Import these packages while using them, for reducing program size.
	"github.com/k8s-practice/octopus/config"
	"github.com/k8s-practice/octopus/config/localfile"
	"github.com/k8s-practice/octopus/config/parser/jsonparser"
	"github.com/k8s-practice/octopus/config/parser/tomlparser"
	"github.com/k8s-practice/octopus/config/parser/yamlparser"
)

func init() {
	log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds | log.Llongfile)
}

func TestNewConfig(t *testing.T) {
	c := config.New(
		config.T().WithScheme(localfile.Scheme()).
			WithPriority(1).
			WithPath("./config.toml").
			WithFormat(tomlparser.Format()),
		config.T().WithScheme(localfile.Scheme()).
			WithPriority(2).
			WithPath("./config.yaml").
			WithFormat(yamlparser.Format()),
		config.T().WithScheme(localfile.Scheme()).
			WithPriority(3).
			WithPath("./config.json").
			WithFormat(jsonparser.Format()),
	)

	log.Println(c.Get("test"))
}
