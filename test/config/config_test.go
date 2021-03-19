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

	"github.com/stretchr/testify/assert"
)

func init() {
	log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds | log.Llongfile)
}

func TestNewConfig(t *testing.T) {
	c := config.New(
		config.T().WithScheme(localfile.Scheme()).
			WithPriority(1).
			WithPath("./p1.toml").
			WithFormat(tomlparser.Format()),
		config.T().WithScheme(localfile.Scheme()).
			WithPriority(2).
			WithPath("./p2.yaml").
			WithFormat(yamlparser.Format()),
		config.T().WithScheme(localfile.Scheme()).
			WithPriority(3).
			WithPath("./p3.json").
			WithFormat(jsonparser.Format()),
	)

	assert.Equal(t, c.Get("database.addr"), "172.168.0.1",
		"p1 database.addr should be 172.168.0.1")

	c.Set("database.addr", "127.0.0.1")
	assert.Equal(t, c.Get("database.addr"), "127.0.0.1",
		"after set, database.addr should be 127.0.0.1")

	c.Set("database.addr", nil)
	assert.Equal(t, c.Get("database.addr"), "172.168.0.1",
		"after delete, database.addr should be 172.168.0.1")

	assert.Equal(t, c.Get("database.port"), 3307,
		"p2 database.port should be 3307")
}
