package test

import (
	"log"
	"testing"
	"time"

	// NOTE: Import these packages while using them, for reducing program size.
	"github.com/k8s-practice/octopus/config"
	"github.com/k8s-practice/octopus/config/datasource/localfile"
	"github.com/k8s-practice/octopus/config/parser/jsonparser"
	"github.com/k8s-practice/octopus/config/parser/tomlparser"
	"github.com/k8s-practice/octopus/config/parser/yamlparser"
	"github.com/stretchr/testify/assert"
)

func init() {
	log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds | log.Llongfile)
}

func TestNewConfig(t *testing.T) {
	c1, err := config.New(
		config.T().WithScheme(localfile.Scheme()).
			WithPath("./p1.toml").
			WithFormat(tomlparser.Format()),
	)
	assert.Equal(t, err, nil, "Must be successful.")

	c2, err := config.New(
		config.T().WithScheme(localfile.Scheme()).
			WithPath("./p2.yaml").
			WithFormat(yamlparser.Format()),
	)
	assert.Equal(t, err, nil, "Must be successful.")

	c3, err := config.New(
		config.T().WithScheme(localfile.Scheme()).
			WithPath("./p3.json").
			WithFormat(jsonparser.Format()),
	)
	assert.Equal(t, err, nil, "Must be successful.")

	c := config.MultiConfig(c1, c2, c3)

	assert.Equal(t, config.GetString(c, "database.info.addr"), "172.168.0.1",
		"p1 database.addr should be 172.168.0.1")

	assert.Equal(t, config.GetInt(c, "database.info.port"), 0,
		"p1 database.port should be zero value.")

	assert.True(t, true, config.GetTime(c, "database.info.time").
		Equal(time.Date(2021, time.March, 3, 19, 14, 15, 16, time.UTC)))

	assert.True(t, true, config.GetInt(c, "database.port"), 3307)
}
