package main

import (
	"flag"
	"fmt"
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

type SqlConfig struct {
	Driver   string `db:"driver"`
	TenantId uint32 `db:"tenant_id"`
	Host     string `db:"host"`
	Port     uint16 `db:"port"`
	User     string `db:"user"`
	Password string `db:"password"`
	DbName   string `db:"db_name"`
}

func (sc *SqlConfig) String() string {
	return fmt.Sprintf("Driver=>%s, TenantId=>%d, Host=>%s, Port=>%d, User=>%s, Password=>%s, DbName=>%s", sc.Driver, sc.TenantId, sc.Host, sc.Port, sc.User, sc.Password, sc.DbName)
}

func init() {
	flag.Parse()
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}

func main() {
	cfg := loadConfig()
	o := octopus.New().WithConfig(cfg)

	var sqlcfg SqlConfig
	if err := o.Load("database", &sqlcfg); err != nil {
		log.Panic(err)
	} else {
		log.Println(sqlcfg.String())
	}

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
