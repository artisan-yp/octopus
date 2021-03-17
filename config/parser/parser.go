package configparser

import (
	"fmt"
	"log"
	"strings"
)

var (
	// parsers is a map from format to config parser builder.
	parsers = make(map[string]Builder)
)

// Parser parses configurations from input data.
type Parser interface {
	Parse(data []byte, v interface{}) error
}

// Builder builds a config parser to parse configiration.
type Builder interface {
	Build() Parser
	// Format returns the format of parsers built by this builder.
	// It will be used to pick config parsers.
	Format() []string
}

func Register(b Builder) {
	log.Printf("Register paser [%s]\n", b.Format())
	for _, format := range b.Format() {
		parsers[strings.ToLower(format)] = b
	}
}

// Parse uses registered parser to parse the coming data.
// - format is used to search parser.
// - data is the data need to parse.
// - v is output parameter.
func Parse(format string, data []byte, v interface{}) error {
	if builder, ok := parsers[strings.ToLower(format)]; ok {
		return builder.Build().Parse(data, v)
	}

	return fmt.Errorf("Unsupported parse format [%s].", format)
}
