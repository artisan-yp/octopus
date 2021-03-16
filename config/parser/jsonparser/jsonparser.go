package jsonparser

import (
	"encoding/json"

	cp "github.com/k8s-practice/octopus/config/parser"
)

const (
	format = "json"
)

func Format() string {
	return format
}

type builder struct{}

func (b *builder) Build() cp.Parser {
	return &parser{}
}

func (b *builder) Format() []string {
	return []string{format}
}

type parser struct{}

func (p *parser) Parse(data []byte, v interface{}) error {
	return json.Unmarshal(data, v)
}

func init() {
	cp.Register(&builder{})
}
