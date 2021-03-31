package tomlparser

import (
	"github.com/BurntSushi/toml"
	"github.com/k8s-practice/octopus/config/parser"
)

const (
	format = "toml"
)

func Format() string {
	return format
}

func IsMatchFormat(fmt string) bool {
	return fmt == format
}

type builder struct{}

func (b *builder) Format() []string {
	return []string{format}
}

func (b *builder) Build() parser.Parser {
	return &tomlParser{}
}

type tomlParser struct{}

func (p *tomlParser) Parse(data []byte, v interface{}) error {
	return toml.Unmarshal(data, v)
}

func init() {
	parser.Register(&builder{})
}
