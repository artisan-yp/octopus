package config

import (
	"log"
	"sort"
	"strings"
	"sync"
)

const (
	DEFAULT_KEY_DELIMITER = "."

	KEY_SCHEME   = "scheme"
	KEY_PRIORITY = "priority"
	KEY_FORMAT   = "format"
	KEY_PATH     = "path"
	KEY_OBSERVER = "observer"
)

// config implements Configurator interface.
// It combines overrides, config file, key/value store, etc.
type config struct {
	// delimiter separates the nested key.
	delimiter string

	// override stores user settings with the highest priority.
	override sync.Map
	// defaults stores default settings with the lowest priority.
	defaults sync.Map

	// dss is an unique priority queue of DataSource
	// NOTE: Priority determines the datasource order when retrival a key.
	dss uniquePriorityQueue
}

// Get gets value by key, it's thread safe.
func (c *config) Get(key string) interface{} {
	// Search in the highest priority map.
	if v, ok := c.override.Load(key); ok {
		return v
	}

	// Search in datasource by priority.
	path := strings.Split(key, c.delimiter)
	for _, ds := range c.dss {
		if v := ds.datasource.Get(path); v != nil {
			return v
		}
	}

	// Search in the lowest priority map.
	if v, ok := c.defaults.Load(key); ok {
		return v
	}

	return nil
}

func (c *config) Set(key string, value interface{}) {
	if value != nil {
		c.override.Store(key, value)
	} else {
		c.override.Delete(key)
	}
}

func (c *config) SetDefault(key string, value interface{}) {
	c.defaults.Store(key, value)
}

// addDataSource inserts an datasource in upq.
func (c *config) addDataSource(d DataSource) {
	i := &item{
		priority:   d.Priority(),
		datasource: d,
	}
	c.dss = c.dss.insert(i)
}

type item struct {
	priority   int32
	datasource DataSource
}

type uniquePriorityQueue []*item

// insert inserts an item in an uniquePriorityQueue.
// If the same priority already exists in the uniquePriorityQueue, panic.
func (upq uniquePriorityQueue) insert(it *item) uniquePriorityQueue {
	idx := sort.Search(len(upq), func(i int) bool {
		return upq[i].priority >= it.priority
	})
	if idx < len(upq) && upq[idx].priority == it.priority {
		log.Panic("The same priority already exists.")
	}

	upq = append(upq, it)
	sort.Slice(upq, func(i, j int) bool {
		return upq[i].priority < upq[j].priority
	})

	return upq
}

// Observer watch change happens on the datasource.
type Observer func()

// Target helps to store the initialize data required by datasource.
type Target interface {
	// There must be scheme filed, otherwise how to find the datasource.
	WithScheme(scheme string) Target
	Scheme() string

	// There must be priority filed, otherwise how to sort datasource.
	WithPriority(priority int32) Target
	Priority() int32

	// Format helps config parser parse the data of the datasource.
	WithFormat(format string) Target
	Format() string

	// Path is the data path of the datasource.
	WithPath(path string) Target
	Path() string

	// Observer is the observer to find configuration changes.
	WithObserver(observers ...Observer) Target
	Observer() []Observer

	// Some other fileds use these functions to access.
	WithValue(key string, value interface{}) Target
	Value(key string) interface{}
}

type target map[string]interface{}

func (t target) WithScheme(scheme string) Target {
	t[KEY_SCHEME] = scheme
	return t
}

func (t target) Scheme() string {
	return t[KEY_SCHEME].(string)
}

func (t target) WithPriority(priority int32) Target {
	t[KEY_PRIORITY] = priority
	return t
}

func (t target) Priority() int32 {
	return t[KEY_PRIORITY].(int32)
}

func (t target) WithFormat(format string) Target {
	t[KEY_FORMAT] = format
	return t
}

func (t target) Format() string {
	return t[KEY_FORMAT].(string)
}

func (t target) WithPath(path string) Target {
	t[KEY_PATH] = path
	return t
}

func (t target) Path() string {
	return t[KEY_PATH].(string)
}

func (t target) WithObserver(observers ...Observer) Target {
	if v, ok := t[KEY_OBSERVER]; ok {
		if w, ok := v.([]Observer); ok {
			t[KEY_OBSERVER] = append(w, observers...)
		} else {
			log.Panicf("Key observer already exists, but it's not observer slice.")
		}
	} else {
		t[KEY_OBSERVER] = observers
	}

	return t
}

func (t target) Observer() []Observer {
	return t[KEY_OBSERVER].([]Observer)
}

func (t target) WithValue(key string, value interface{}) Target {
	t[key] = value
	return t
}

func (t target) Value(key string) interface{} {
	return t[key]
}
