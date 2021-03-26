package yamlparser

import (
	"github.com/k8s-practice/octopus/config/parser"
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

func (b *builder) Build() parser.Parser {
	return &yamlParser{}
}

type yamlParser struct{}

func (p *yamlParser) Parse(data []byte, v interface{}) error {
	return yaml.Unmarshal(data, v)
}

func init() {
	parser.Register(&builder{})
}
