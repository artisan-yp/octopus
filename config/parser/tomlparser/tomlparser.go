package tomlparser

import (
	"github.com/BurntSushi/toml"
	cp "github.com/k8s-practice/octopus/config/parser"
)

const (
	format = "toml"
)

func Format() string {
	return format
}

type builder struct{}

func (b *builder) Format() []string {
	return []string{format}
}

func (b *builder) Build() cp.Parser {
	return &parser{}
}

type parser struct{}

func (p *parser) Parse(data []byte, v interface{}) error {
	return toml.Unmarshal(data, v)
}

func init() {
	cp.Register(&builder{})
}
