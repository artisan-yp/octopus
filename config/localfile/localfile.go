package localfile

import (
	"log"
	"os"
	"sync/atomic"

	"github.com/k8s-practice/octopus/config"
	cp "github.com/k8s-practice/octopus/config/parser"
	cs "github.com/k8s-practice/octopus/internal/configsearch"
)

const (
	scheme = "localfile"
)

func init() {
	config.RegisterBuilder(&builder{})
}

func Scheme() string {
	return scheme
}

type builder struct{}

// Build builds a config.DataSource by config.Target.
func (b *builder) Build(t config.Target) (config.DataSource, error) {
	d := &datasource{
		priority: t.Priority(),
		filepath: t.Path(),
		format:   t.Format(),
	}
	d.config.Store(make(map[string]interface{}))

	return d, d.Load()
}

func (b *builder) Scheme() string {
	return Scheme()
}

// datasource implements config.Configurator interface.
type datasource struct {
	// filepath is the path of the datasource file,
	// absolute or relative path.
	filepath string

	// format is the format of the datasource file,
	// it could be json, toml or yaml, etc.
	format string

	// priority is the priority of configuration from this datasource.
	// The higher the value, the higher the priority.
	priority int32

	// config contains all configurations.
	// Value store type is map[string]interface{}
	config atomic.Value
}

func (d *datasource) Load() (err error) {
	if data, err := os.ReadFile(d.filepath); err != nil {
		log.Println(err)
	} else {
		config := make(map[string]interface{})
		if err = cp.Parse(d.format, data, &config); err != nil {
			log.Println(err)
		} else {
			d.config.Store(config)
		}
	}

	return err
}

func (d *datasource) Get(path []string) interface{} {
	m := d.config.Load()
	if m == nil {
		return nil
	}

	return cs.SearchPathInMap(m.(map[string]interface{}), path)
}

func (d *datasource) Priority() int32 {
	return d.priority
}

/*
func (d *datasource) findConfigFile() (string, error) {
	for _, dir := range d.dirs {
		// Don't care the target is a symlink or not.
		filePath := path.Join(dir, d.fileName+"."+d.fileType)
		fileInfo, err := os.Stat(filePath)
		// Ignore all errors.
		// If target is a symlink, fileInfo is the FileInfo of final target.
		if err == nil && !fileInfo.IsDir() {
			return filePath, nil
		}
	}

	// Finally, if search dirs is empty, search in current folder.
	if len(d.dirs) == 0 {
		filePath := path.Join("./", d.fileName+"."+d.fileType)
		fileInfo, err := os.Stat(filePath)
		// Ignore all errors.
		// If target is a symlink, fileInfo is the FileInfo of final target.
		if err == nil && !fileInfo.IsDir() {
			return filePath, nil
		}
	}

	return "", errors.New("Config file not found.")
}*/
