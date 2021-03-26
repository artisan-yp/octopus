package jsonparser

import (
	"encoding/json"

	"github.com/k8s-practice/octopus/config/parser"
)

const (
	format = "json"
)

func Format() string {
	return format
}

type builder struct{}

func (b *builder) Build() parser.Parser {
	return &jsonParser{}
}

func (b *builder) Format() []string {
	return []string{format}
}

type jsonParser struct{}

func (p *jsonParser) Parse(data []byte, v interface{}) error {
	return json.Unmarshal(data, v)
}

func init() {
	parser.Register(&builder{})
}
