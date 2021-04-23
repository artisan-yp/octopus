package prometheus

import (
	"errors"
	"fmt"
	"net/http"
	"sync"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const (
	Empty     = iota // int默认值, 无用类型
	Counter          // 计数器, 只能增加或重启清零
	Gauge            // 测量器
	Histogram        // 累计直方图
	Summary          // 采样点分位图
)

var (
	ErrArgNil     = errors.New("metric nil or name、type not specify")
	ErrMetricType = errors.New("metric type not correct")
)

var (
	prometheusManager *ExporterManager = nil
	prometheusReg                      = prometheus.NewRegistry()
	initHttpOnce      sync.Once        // 单个服务只用初始化http handler一次
	initDataOnce      sync.Once        // 总的数据结构初始化控制
)

type OpenPrometheus struct {
	Port int // 用于普罗米修斯http汇报的端口
}

type Metric struct {
	Name        string            // 指标名称, 全局唯一
	Help        string            // 指标帮助信息
	MetricType  int               // 监控指标类型
	ConstLabels map[string]string // 固定标签
	Labels      []string          // 指标数据标签

	Buckets    []float64           // histogram buckets细分区间, 值为统计值落入桶的判断
	Objectives map[float64]float64 // summary提前指定的分位数及可允许误差
}

// -----------------------------------------------对外提供服务接口----------------------------------------------
type ExporterManager struct {
	counterVecInfo   *counterVecDesc
	gaugeVecInfo     *gaugeVecDesc
	histogramVecInfo *histogramVecDesc
	summaryVecInfo   *summaryVecDesc
}

func pickDesc(metricType int) metricDesc {
	switch metricType {
	case Counter:
		return prometheusManager.counterVecInfo
	case Gauge:
		return prometheusManager.gaugeVecInfo
	case Histogram:
		return prometheusManager.histogramVecInfo
	case Summary:
		return prometheusManager.summaryVecInfo
	default:
		return &emptyDesc{}
	}
}

func Manager() *ExporterManager {
	initDataOnce.Do(func() {
		prometheusManager = &ExporterManager{
			counterVecInfo: &counterVecDesc{
				descs: make(map[string]*counterVecMetric),
			},

			gaugeVecInfo: &gaugeVecDesc{
				descs: make(map[string]*gaugeVecMetric),
			},

			histogramVecInfo: &histogramVecDesc{
				descs: make(map[string]*histogramVecMetric),
			},

			summaryVecInfo: &summaryVecDesc{
				descs: make(map[string]*summaryVecMetric),
			},
		}
	})

	return prometheusManager
}

// @params: name && metricType is nessary
func (m *ExporterManager) RegisterMetric(metric *Metric) error {
	if m == nil || metric == nil {
		return ErrArgNil
	}

	return pickDesc(metric.MetricType).registerMetric(metric)
}

/*
 *@func: 取消指标监控
 *@params: Name
 */
func (m *ExporterManager) UnRegisterMetric(metric *Metric) {
	if m == nil || metric == nil {
		return
	}

	/*
	 * Note that even after unregistering, it will not be possible to
	 * register a new Collector that is inconsistent with the unregistered
	 * Collector, e.g. a Collector collecting metrics with the same name but
	 * a different help string. The rationale here is that the same registry
	 * instance must only collect consistent metrics throughout its
	 * lifetime.
	 */
	pickDesc(metric.MetricType).unregisterMetric(metric)
}

/*
 *@func: 原子性指标值+n
 *@params: Name, MetricType; optional: labels
 */
func (m *ExporterManager) Add(metric *Metric, val float64, labels map[string]string) error {
	if m == nil || metric == nil || metric.Name == "" {
		return ErrArgNil
	}

	return pickDesc(metric.MetricType).add(metric, val, labels)
}

func (m *ExporterManager) Set(metric *Metric, val float64, labels map[string]string) error {
	if m == nil || metric == nil || metric.Name == "" {
		return ErrArgNil
	}

	return pickDesc(metric.MetricType).set(metric, val, labels)
}

func Register(port int) {
	initHttpOnce.Do(func() {
		// 考虑之后通过consul+prometheus自动发现新起监控节点
		addr := fmt.Sprintf(":%d", port)

		gathers := prometheus.Gatherers{
			prometheus.DefaultGatherer,
			prometheusReg,
		}

		continueHandler := promhttp.HandlerFor(gathers, promhttp.HandlerOpts{
			ErrorHandling: promhttp.ContinueOnError,
			Registry:      prometheusReg,
		})

		http.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
			continueHandler.ServeHTTP(w, r)
		})

		go func() {
			panic(http.ListenAndServe(addr, nil))
		}()
	},
	)
}
