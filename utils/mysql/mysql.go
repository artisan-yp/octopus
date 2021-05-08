package mysql

import (
	"errors"
	"fmt"
	"sync"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"

	"github.com/k8s-practice/octopus/utils/prometheus"
)

const (
	defaultMaxOpenConns    = 10
	defaultMaxIdleConns    = 1
	defaultMaxConnLifeTime = 1 * time.Minute
	mysqlInvokeMetric      = "mysql_invoke_total"
	mysqlConnMetric        = "mysql_conn_status"
)

type MysqlDriverOptFunc func(*MysqlConnOpts)

type Limiter interface {
	// user realize
	Allow() error
}

type MysqlConnOpts struct {
	maxOpenConns    int
	maxIdleConns    int
	maxConnLifeTime time.Duration

	// 监控项
	monitor *prometheus.OpenPrometheus
	metrics map[string]*prometheus.Metric

	limiter Limiter // 限频器

	c *MysqlControl
}

type MysqlDriverBasic struct {
	Id string // 连接池id

	UserName string
	Passwd   string
	Host     string
	DB       string
	Port     int
}

type MysqlControl struct {
	*sqlx.DB

	info *MysqlDriverBasic

	opts *MysqlConnOpts
}

type MysqlControls struct {
	Controls map[string]*MysqlControl
	sync.RWMutex
}

var (
	MysqlConnPools MysqlControls

	ErrArgsInvalid = errors.New("args invalid")

	ErrTooManyRequest = errors.New("too many request gt specify qps")
)

func init() {
	MysqlConnPools.Controls = make(map[string]*MysqlControl)
}

func (info *MysqlDriverBasic) validate() error {
	if info == nil {
		return errors.New("args nil")
	} else if len(info.Id) == 0 {
		return errors.New("mysql conn pool init miss id")
	} else if len(info.UserName) == 0 {
		return errors.New("mysql conn pool init miss username")
	} else if len(info.Passwd) == 0 {
		return errors.New("mysql conn pool init miss passwd")
	} else if len(info.Host) == 0 {
		return errors.New("mysql conn pool init miss host")
	} else if info.Port <= 0 || info.Port >= 65535 {
		return errors.New("mysql conn pool init port error")
	}

	return nil
}

func WithMaxOpenConns(n int) MysqlDriverOptFunc {
	return func(o *MysqlConnOpts) { o.maxOpenConns = n }
}

func WithMaxIdleConns(n int) MysqlDriverOptFunc {
	return func(o *MysqlConnOpts) { o.maxIdleConns = n }
}

func WithMaxConnLifeTime(t time.Duration) MysqlDriverOptFunc {
	return func(o *MysqlConnOpts) { o.maxConnLifeTime = t }
}

func WithPrometheus(p int) MysqlDriverOptFunc {
	return func(o *MysqlConnOpts) {
		if p <= 0 || p > 65535 {
			return
		}

		o.monitor = &prometheus.OpenPrometheus{Port: p}

		o.initPrometheus()
	}
}

func WithLimiter(l Limiter) MysqlDriverOptFunc {
	return func(o *MysqlConnOpts) {
		o.limiter = l
	}
}

func (c *MysqlConnOpts) initPrometheus() {
	if c == nil || c.monitor == nil || c.c == nil {
		return
	}

	if c.metrics == nil {
		c.metrics = make(map[string]*prometheus.Metric)
	}

	prometheus.Register(c.monitor.Port)

	c.metrics[mysqlConnMetric] = &prometheus.Metric{
		Name:       mysqlConnMetric,
		Help:       "statics of mysql pool conns",
		MetricType: prometheus.Gauge,
		Labels:     []string{mysqlConnMetric},
	}

	/*
		c.metrics[mysqlInvokeMetric] = &prometheus.Metric{
			// mysql访问量 && 失败率
			Name:       mysqlInvokeMetric,
			Help:       "statics of mysql invoke status",
			MetricType: prometheus.Counter,
			Labels:     []string{mysqlInvokeMetric},
		}
	*/

	for _, v := range c.metrics {
		prometheus.Manager().RegisterMetric(v)
	}

	// 定时监控mysql连接池情况
	go func() {
		for {
			select {
			case <-time.After(time.Second):
				stats := c.c.Stats()
				prometheus.Manager().Set(c.metrics[mysqlConnMetric], float64(stats.OpenConnections), map[string]string{mysqlConnMetric: "openConns"})
			}
		}
	}()
}

func Register(info *MysqlDriverBasic, opts ...MysqlDriverOptFunc) (*MysqlControl, error) {
	if err := info.validate(); err != nil {
		return nil, err
	}

	MysqlConnPools.RLock()
	if _, ok := MysqlConnPools.Controls[info.Id]; ok {
		MysqlConnPools.RUnlock()
		return nil, errors.New(info.Id + " already register done")
	}

	MysqlConnPools.RUnlock()

	ds := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s", info.UserName, info.Passwd, info.Host, info.Port, info.DB)
	db, err := sqlx.Connect("mysql", ds)
	if err != nil {
		return nil, fmt.Errorf("mysql connect error! err: %s", err.Error())
	}

	MysqlConnPools.Lock()
	defer MysqlConnPools.Unlock()

	c := &MysqlControl{
		db,
		info,
		&MysqlConnOpts{
			maxOpenConns:    defaultMaxOpenConns,
			maxIdleConns:    defaultMaxIdleConns,
			maxConnLifeTime: defaultMaxConnLifeTime,
		},
	}

	c.opts.c = c

	for _, o := range opts {
		o(c.opts)
	}

	MysqlConnPools.Controls[info.Id] = c

	return c, nil
}

/*
 * 获取id对应的mysql连接
 */
func Conn(id string) *MysqlControl {
	MysqlConnPools.RLock()
	defer MysqlConnPools.RUnlock()

	if v, ok := MysqlConnPools.Controls[id]; ok {
		return v
	} else {
		return nil
	}
}

/*
 * 带有用户自定义限制如qps控制的mysql链接
 */
func (c *MysqlControl) Conn() (*MysqlControl, error) {
	if c == nil {
		return c, ErrArgsInvalid
	}

	if c.opts != nil && c.opts.limiter != nil {
		if err := c.opts.limiter.Allow(); err != nil {
			return c, err
		}
	}

	return c, nil
}

func (c *MysqlControl) DataBase() string {
	if c == nil || c.info == nil {
		return ""
	}

	return c.info.DB
}
