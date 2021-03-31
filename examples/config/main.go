package main

import (
	"flag"
	"log"
	"path/filepath"
	"strings"

	"github.com/k8s-practice/octopus"
	"github.com/k8s-practice/octopus/config"
	"github.com/k8s-practice/octopus/config/datasource/localfile"
	"github.com/k8s-practice/octopus/config/parser/jsonparser"
	"github.com/k8s-practice/octopus/config/parser/tomlparser"
	"github.com/k8s-practice/octopus/config/parser/yamlparser"
)

func init() {
	flag.Parse()
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}

func main() {
	conf := loadConfig()
	o := octopus.New().WithConfig(conf)

	var wrapper Wrapper
	if err := o.Load("", &wrapper); err != nil {
		log.Panic(err)
	} else {
		log.Println(wrapper)
	}

	var db Database
	if err := o.Load("database", &db); err != nil {
		log.Panic(err)
	} else {
		log.Println(db)
	}

	var info Info
	if err := o.Load("database.info", &info); err != nil {
		log.Panic(err)
	} else {
		log.Println(info)
	}
}

var configfile = flag.String("c", "./config.toml", "Config file path.")
var configformat = flag.String("f", "", "Config file format.")

// InitConfig initialize base config from local file.
func loadConfig() config.Config {
	format := *configformat
	if format == "" {
		suffix := filepath.Ext(*configfile)
		format = strings.ToLower(strings.TrimPrefix(suffix, "."))
	}

	switch {
	case tomlparser.IsMatchFormat(format):
		if c, err := config.New(config.T().WithScheme(localfile.Scheme()).
			WithPath(*configfile).
			WithFormat(format)); err != nil {
			log.Panic(err)
		} else {
			return c
		}
	case yamlparser.IsMatchFormat(format):
		if c, err := config.New(config.T().WithScheme(localfile.Scheme()).
			WithPath(*configfile).
			WithFormat(format)); err != nil {
			log.Panic(err)
		} else {
			return c
		}
	case jsonparser.IsMatchFormat(format):
		if c, err := config.New(config.T().WithScheme(localfile.Scheme()).
			WithPath(*configfile).
			WithFormat(format)); err != nil {
			log.Panic(err)
		} else {
			return c
		}
	default:
		log.Panic("Unknown config file format.")
	}

	return nil
}

type Wrapper struct {
	Database
}

type Database struct {
	Info
}

type Info struct {
	Addr     string
	Port     int16
	User     string
	Password string
}
