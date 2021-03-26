package localfile

import (
	"os"
	"sync/atomic"

	"github.com/k8s-practice/octopus/config/datasource"
	"github.com/k8s-practice/octopus/config/parser"
	"github.com/k8s-practice/octopus/internal/configsearch"
)

const (
	scheme = "localfile"
)

func init() {
	datasource.Register(&builder{})
}

func Scheme() string {
	return scheme
}

type builder struct{}

// Build builds a datasource.DataSource by datasource.Target.
func (b *builder) Build(t datasource.Target) (datasource.DataSource, error) {
	d := &localfile{
		filepath: t.Path(),
		format:   t.Format(),
	}
	d.config.Store(make(map[string]interface{}))

	return d, d.Load()
}

func (b *builder) Scheme() string {
	return Scheme()
}

// localfile implements the interface of datasource.DataSource.
type localfile struct {
	// filepath is the path of the datasource file,
	// absolute or relative path.
	filepath string

	// format is the format of the datasource file,
	// it could be json, toml or yaml, etc.
	format string

	// config contains all configurations.
	// Value store type is map[string]interface{}
	config atomic.Value
}

func (d *localfile) Load() error {
	data, err := os.ReadFile(d.filepath)
	if err != nil {
		return err
	}

	config := make(map[string]interface{})
	if err = parser.Parse(d.format, data, &config); err != nil {
		return err
	}
	d.config.Store(config)

	return err
}

func (d *localfile) Get(path []string) interface{} {
	m := d.config.Load()
	if m == nil {
		return nil
	}

	return configsearch.SearchPathInMap(m.(map[string]interface{}), path)
}

/*
func (d *localfile) findConfigFile() (string, error) {
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
