package prometheus

import (
	"fmt"
	"sync"

	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
)

/*
 *@moduel: 监控模块中未区分出*Desc, *VecDesc类型, 统一使用*VecDesc, 降低数据类型复杂度
 */
type metricDesc interface {
	registerMetric(metric *Metric) error
	unregisterMetric(metric *Metric)

	add(*Metric, float64, prometheus.Labels) error
	set(*Metric, float64, prometheus.Labels) error
}

type counterVecMetric struct {
	counter *prometheus.CounterVec
}

type gaugeVecMetric struct {
	gauge *prometheus.GaugeVec
}

type histogramVecMetric struct {
	histogram *prometheus.HistogramVec
}

type summaryVecMetric struct {
	summary *prometheus.SummaryVec
}

type counterVecDesc struct {
	sync.Mutex
	descs map[string]*counterVecMetric
}

type gaugeVecDesc struct {
	sync.Mutex
	descs map[string]*gaugeVecMetric
}

type histogramVecDesc struct {
	sync.Mutex
	descs map[string]*histogramVecMetric
}

type summaryVecDesc struct {
	sync.Mutex
	descs map[string]*summaryVecMetric
}

type emptyDesc struct{}

func (desc *emptyDesc) registerMetric(metric *Metric) error {
	return ErrMetricType
}

func (desc *emptyDesc) unregisterMetric(metric *Metric) {
	return
}

func (desc *emptyDesc) add(*Metric, float64, prometheus.Labels) error {
	return ErrMetricType
}

func (desc *emptyDesc) set(*Metric, float64, prometheus.Labels) error {
	return ErrMetricType
}

func (desc *counterVecDesc) registerMetric(metric *Metric) (err error) {
	if desc == nil || metric == nil || metric.Name == "" {
		return ErrArgNil
	}

	defer func() {
		desc.Unlock()

		// prometheus重复注册相同metric报错
		if e := recover(); e != nil {
			err = errors.Wrapf(fmt.Errorf("%s", e), "register counter metric %+v failed", *metric)
			desc.unregisterMetric(metric)
		}
	}()

	desc.Lock()

	if _, ok := desc.descs[metric.Name]; !ok {
		desc.descs[metric.Name] = &counterVecMetric{
			counter: prometheus.NewCounterVec(prometheus.CounterOpts{
				Name:        metric.Name,
				Help:        metric.Help,
				ConstLabels: metric.ConstLabels,
			}, metric.Labels),
		}

		prometheusReg.MustRegister(desc.descs[metric.Name].counter)

	}

	return
}

func (desc *counterVecDesc) unregisterMetric(metric *Metric) {
	if desc == nil || metric == nil || metric.Name == "" {
		return
	}

	desc.Lock()
	defer desc.Unlock()

	if _, ok := desc.descs[metric.Name]; ok {
		prometheus.Unregister(desc.descs[metric.Name].counter)
		delete(desc.descs, metric.Name)
	}
}

func (desc *counterVecDesc) add(metric *Metric, val float64, labels prometheus.Labels) (err error) {
	if val <= 0 || metric == nil || metric.Name == "" || desc == nil {
		return
	}

	desc.Lock()
	defer func() {
		desc.Unlock()
		if e := recover(); e != nil {
			err = errors.Wrapf(fmt.Errorf("%s", e), "try add counter metric %+v failed", *metric)
		}
	}()

	if _, ok := desc.descs[metric.Name]; ok {
		desc.descs[metric.Name].counter.With(labels).Add(val)
	}

	return
}

func (desc *counterVecDesc) set(*Metric, float64, prometheus.Labels) error {
	return nil
}

func (desc *gaugeVecDesc) registerMetric(metric *Metric) (err error) {
	if desc == nil || metric == nil || metric.Name == "" {
		return ErrArgNil
	}

	defer func() {
		desc.Unlock()

		// prometheus重复注册相同metric报错
		if e := recover(); e != nil {
			err = errors.Wrapf(fmt.Errorf("%s", e), "register gauge metric %+v failed", *metric)
			desc.unregisterMetric(metric)
		}
	}()

	desc.Lock()

	if _, ok := desc.descs[metric.Name]; !ok {

		desc.descs[metric.Name] = &gaugeVecMetric{
			gauge: prometheus.NewGaugeVec(prometheus.GaugeOpts{
				Name:        metric.Name,
				Help:        metric.Help,
				ConstLabels: metric.ConstLabels,
			}, metric.Labels),
		}

		prometheusReg.MustRegister(desc.descs[metric.Name].gauge)

	}

	return
}

func (desc *gaugeVecDesc) unregisterMetric(metric *Metric) {
	if desc == nil || metric == nil || metric.Name == "" {
		return
	}

	desc.Lock()
	defer desc.Unlock()
	if _, ok := desc.descs[metric.Name]; ok {
		prometheus.Unregister(desc.descs[metric.Name].gauge)
		delete(desc.descs, metric.Name)
	}
}

func (desc *gaugeVecDesc) add(metric *Metric, val float64, labels prometheus.Labels) (err error) {
	if desc == nil || metric == nil || metric.Name == "" {
		return
	}

	desc.Lock()
	defer func() {
		desc.Unlock()
		if e := recover(); e != nil {
			err = errors.Wrapf(fmt.Errorf("%s", e), "try add gauge metric %+v failed", *metric)
		}
	}()

	if _, ok := desc.descs[metric.Name]; ok {
		desc.descs[metric.Name].gauge.With(labels).Add(val)
	}

	return
}

func (desc *gaugeVecDesc) set(metric *Metric, val float64, labels prometheus.Labels) (err error) {
	if desc == nil || metric == nil || metric.Name == "" {
		return
	}

	desc.Lock()
	defer func() {
		desc.Unlock()
		if e := recover(); e != nil {
			err = errors.Wrapf(fmt.Errorf("%s", e), "try set gauge metric %+v failed", *metric)
		}
	}()

	if _, ok := desc.descs[metric.Name]; ok {
		desc.descs[metric.Name].gauge.With(labels).Set(val)
	}

	return
}

func (desc *histogramVecDesc) registerMetric(metric *Metric) (err error) {
	if desc == nil || metric == nil || metric.Name == "" || len(metric.Buckets) == 0 {
		return ErrArgNil
	}

	defer func() {
		desc.Unlock()

		// prometheus重复注册相同metric报错
		if e := recover(); e != nil {
			err = errors.Wrapf(fmt.Errorf("%s", e), "register histogram metric %+v failed", *metric)
			desc.unregisterMetric(metric)
		}
	}()

	desc.Lock()

	if _, ok := desc.descs[metric.Name]; !ok {

		desc.descs[metric.Name] = &histogramVecMetric{
			histogram: prometheus.NewHistogramVec(prometheus.HistogramOpts{
				Name:        metric.Name,
				Help:        metric.Help,
				ConstLabels: metric.ConstLabels,
				Buckets:     metric.Buckets,
			}, metric.Labels),
		}

		prometheusReg.MustRegister(desc.descs[metric.Name].histogram)

	}

	return
}

func (desc *histogramVecDesc) unregisterMetric(metric *Metric) {
	if desc == nil || metric == nil || metric.Name == "" {
		return
	}

	desc.Lock()
	defer desc.Unlock()
	if _, ok := desc.descs[metric.Name]; ok {
		prometheus.Unregister(desc.descs[metric.Name].histogram)
		delete(desc.descs, metric.Name)
	}
}

/*
 *@func: histogram metric add func == observe add val into buckets
 */
func (desc *histogramVecDesc) add(metric *Metric, val float64, labels prometheus.Labels) (err error) {
	if desc == nil || metric == nil || metric.Name == "" {
		return
	}

	desc.Lock()
	defer func() {
		desc.Unlock()
		if e := recover(); e != nil {
			err = errors.Wrapf(fmt.Errorf("%s", e), "try add histogram metric %+v failed", *metric)
		}
	}()

	if _, ok := desc.descs[metric.Name]; ok {
		desc.descs[metric.Name].histogram.With(labels).Observe(val)
	}

	return
}

func (desc *histogramVecDesc) set(metric *Metric, val float64, labels prometheus.Labels) error {
	return nil
}

func (desc *summaryVecDesc) registerMetric(metric *Metric) (err error) {
	if desc == nil || metric == nil || metric.Name == "" || len(metric.Objectives) == 0 {
		return ErrArgNil
	}

	defer func() {
		desc.Unlock()

		// prometheus重复注册相同metric报错
		if e := recover(); e != nil {
			err = errors.Wrapf(fmt.Errorf("%s", e), "register summary metric %+v failed", *metric)
			desc.unregisterMetric(metric)
		}
	}()

	desc.Lock()

	if _, ok := desc.descs[metric.Name]; !ok {

		desc.descs[metric.Name] = &summaryVecMetric{
			summary: prometheus.NewSummaryVec(prometheus.SummaryOpts{
				Name:        metric.Name,
				Help:        metric.Help,
				ConstLabels: metric.ConstLabels,
				Objectives:  metric.Objectives,
			}, metric.Labels),
		}

		prometheusReg.MustRegister(desc.descs[metric.Name].summary)

	}

	return
}

func (desc *summaryVecDesc) unregisterMetric(metric *Metric) {
	if desc == nil || metric == nil || metric.Name == "" {
		return
	}

	desc.Lock()
	defer desc.Unlock()
	if _, ok := desc.descs[metric.Name]; ok {
		prometheus.Unregister(desc.descs[metric.Name].summary)
		delete(desc.descs, metric.Name)
	}
}

func (desc *summaryVecDesc) add(metric *Metric, val float64, labels prometheus.Labels) (err error) {
	if desc == nil || metric == nil || metric.Name == "" {
		return
	}

	desc.Lock()
	defer func() {
		desc.Unlock()
		if e := recover(); e != nil {
			err = errors.Wrapf(fmt.Errorf("%s", e), "try add summary metric %+v failed", *metric)
		}
	}()

	if _, ok := desc.descs[metric.Name]; ok {
		desc.descs[metric.Name].summary.With(labels).Observe(val)
	}

	return
}

func (desc *summaryVecDesc) set(metric *Metric, val float64, labels prometheus.Labels) error {
	return nil
}
