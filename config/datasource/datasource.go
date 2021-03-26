package datasource

import (
	"fmt"
	"log"

	"errors"
)

var (
	// builders is a map from scheme to datasource builder.
	builders = make(map[string]Builder)
)

// Register registers the datasource builder to the datasource map.
func Register(b Builder) {
	if _, ok := builders[b.Scheme()]; ok {
		log.Panicf("Already regestered config scheme [%s].", b.Scheme())
	}

	log.Printf("Register datasource scheme [%s].", b.Scheme())
	builders[b.Scheme()] = b
}

// Build builds an object implements DataSource.
func Build(t Target) (DataSource, error) {
	builder, ok := builders[t.Scheme()]
	if !ok {
		return nil, errors.New(fmt.Sprintf("Unknown data source scheme [%s].", t.Scheme()))
	}

	datasource, err := builder.Build(t)

	return datasource, err
}

// DataSource load data from local file, etcd, consul, env variables, etc.
// NOTE: DataSource must to be thread safe.
type DataSource interface {
	// Load reads data from datasource, and parses it.
	Load() error

	// Get returns the value stored in the path, or nil if no value is present.  // path (e.g. []string{"mysql", "addr"}) means finding "mysql.addr" in // datasource.
	Get(path []string) interface{}
}

// Target helps to store the initialize data required by datasource.
type Target interface {
	// There must be scheme filed, otherwise how to find the datasource.
	WithScheme(scheme string) Target
	Scheme() string

	// Format helps config parser parse the data of the datasource.
	WithFormat(format string) Target
	Format() string

	// Path is the data path of the datasource.
	WithPath(path string) Target
	Path() string

	// Some other fileds use these functions to access.
	WithValue(key string, value interface{}) Target
	Value(key string) interface{}
}

// Builder builds a DataSource.
type Builder interface {
	Build(t Target) (DataSource, error)
	Scheme() string
}
