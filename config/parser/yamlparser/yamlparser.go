package yamlparser

import (
	cp "github.com/k8s-practice/octopus/config/parser"
	"gopkg.in/yaml.v2"
)

const (
	format       = "yaml"
	format_alias = "yml"
)

func Format() string {
	return format
}

type builder struct{}

func (b *builder) Format() []string {
	return []string{format, format_alias}
}

func (b *builder) Build() cp.Parser {
	return &parser{}
}

type parser struct{}

func (p *parser) Parse(data []byte, v interface{}) error {
	return yaml.Unmarshal(data, v)
}

func init() {
	cp.Register(&builder{})
}
