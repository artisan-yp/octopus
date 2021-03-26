package octopus

import (
	"github.com/k8s-practice/octopus/config"
)

func New() *Octopus {
	return &Octopus{}
}

// Brain is the brain of octopus.
type Octopus struct {
	conf config.Config

	// frameInit is the framework initialize functions slice.
	// It's will be invoked first.
	frameInit []func() error

	// appInit is the application initialize functions slice.
	// It's will be invoded after Run function.
	appInit []func() error
}

func (o *Octopus) WithConfig(c config.Config) *Octopus {
	o.conf = c

	return o
}
