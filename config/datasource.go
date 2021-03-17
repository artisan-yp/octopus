package config

import (
	"log"
)

var (
	// builders is a map from scheme to datasource builder.
	builders = make(map[string]Builder)
)

// Register registers the datasource builder to the datasource map.
func RegisterBuilder(b Builder) {
	if _, ok := builders[b.Scheme()]; ok {
		log.Panicf("Already regestered config scheme [%s].", b.Scheme())
	}

	log.Printf("Register datasource scheme [%s].", b.Scheme())
	builders[b.Scheme()] = b
}

// SearchBuilder searches the datasource builder from the datasource map.
func SearchBuilder(scheme string) (Builder, bool) {
	builder, ok := builders[scheme]
	return builder, ok
}

// DataSource load data from local file, etcd, consul, env variables, etc.
// NOTE: DataSource must to be thread safe.
type DataSource interface {
	// Load reads data from datasource, and parses it.
	Load() error

	// Priority represents datasource priority.
	// NOTE: The higher the value, the higher the priority.
	// NOTE: The priority must be unique, or it causes panic.
	Priority() int32

	// Get returns the value stored in the path, or nil if no value is present.
	// path (e.g. []string{"mysql", "addr"}) means finding "mysql.addr" in
	// datasource.
	Get(path []string) interface{}
}

// Builder creates a DataSource.
type Builder interface {
	Build(t Target) (DataSource, error)
	Scheme() string
}
