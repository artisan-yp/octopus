package prometheus

import (
	"fmt"
	"math/rand"
	"sync"
	"testing"
)

var (
	port = 8089
	loop = 10

	id         int
	name       = "test"
	buckets    = []float64{0.1, 0.5, 0.9}
	objectives = map[float64]float64{0.1: 0.001, 0.5: 0.005, 0.9: 0.009}
	metric     = &Metric{}
)

func (m *Metric) empty() {
	m.Name, m.MetricType, m.Help = "", 0, ""
	m.ConstLabels, m.Labels, m.Buckets, m.Objectives = nil, nil, nil, nil
}

func Name() string {
	id++
	return fmt.Sprintf("%s%d", name, id)
}

// 顺序处理数据
func dataSerizeProcess(metric *Metric, labels map[string]string, skip bool) (err error) {
	if metric != nil && metric.Name != "" && !skip {
		metric.Name = Name()
	}

	err = Manager().RegisterMetric(metric)

	var val float64

	for i := 0; i < loop; i++ {
		val = float64(rand.Intn(loop))
		if i < loop/2 {
			val = (-1.0) * val
		} else if i == loop/2 {
			val = 0
		}

		Manager().Add(metric, val, labels)
		Manager().Set(metric, val, labels)
	}

	Manager().UnRegisterMetric(metric)

	return
}

// 并发处理数据
func dataConcProcess(metric *Metric, labels map[string]string, skip bool) {
	if metric != nil && metric.Name != "" && !skip {
		metric.Name = Name()
	}

	Manager().RegisterMetric(metric)

	var wg sync.WaitGroup
	wg.Add(10)

	for i := 0; i < 10; i++ {
		go func() {
			var val float64
			for j := 0; j < loop; j++ {
				val = float64(rand.Intn(loop))
				if i < loop/2 {
					val = (-1.0) * val
				}

				Manager().Add(metric, val, labels)
				Manager().Set(metric, val, labels)
			}

			wg.Done()
		}()
	}

	wg.Wait()
	Manager().UnRegisterMetric(metric)
}

func dataProcess(metric *Metric, labels map[string]string) {
	dataSerizeProcess(metric, labels, false)
	dataConcProcess(metric, labels, false)
}

func wrapError(t *testing.T, fn func() error, dst error) {
	if err := fn(); err != dst {
		t.Errorf("excepted %s, but got %s", dst, err)
	}
}

func wrapLog(t *testing.T, fn func() error) {
	if err := fn(); err != nil {
		t.Logf("err is: %s", err)
	}
}

func TestErrs(t *testing.T) {
	// nil metric
	Manager().RegisterMetric(nil)

	// err metric type
	metric.Name, metric.MetricType = "testErr", 10
	wrapError(t, func() error {
		return Manager().RegisterMetric(metric)
	}, ErrMetricType)

	// register same metric repeat
	metric.MetricType = Counter
	for i := 0; i < 2; i++ {
		wrapError(t, func() error { return Manager().RegisterMetric(metric) }, nil)
	}

	// unregister nil metric
	Manager().UnRegisterMetric(nil)

	// unregister not registered metric name
	metric.Name = name
	Manager().UnRegisterMetric(metric)

	// add or set not registered metric name
	wrapError(t, func() error { return Manager().Add(metric, 1, nil) }, nil)
	wrapError(t, func() error { return Manager().Set(metric, 1, nil) }, nil)

	// add same metric value but different metric type
	metric.Name, metric.MetricType = name, Counter
	wrapLog(t, func() error {
		return Manager().RegisterMetric(metric)
	})

	metric.MetricType = Gauge
	wrapLog(t, func() error {
		return dataSerizeProcess(metric, nil, true)
	})

	metric.MetricType = Histogram
	wrapLog(t, func() error {
		return dataSerizeProcess(metric, nil, true)
	})

	metric.MetricType = Summary
	wrapLog(t, func() error {
		return dataSerizeProcess(metric, nil, true)
	})
}

func TestInvalidCounter(t *testing.T) {
	/*
	 * invalid metric conds:
	 * 1. nil metric
	 * 2. empty name
	 * 3. empty metricType
	 */
	metric.empty()
	metric.MetricType = Counter
	dataProcess(metric, nil)

	metric.Name, metric.MetricType = name, 0
	dataProcess(metric, nil)
}

func TestSingleCounter(t *testing.T) {
	/*
	 * single counter, nil labels
	 * 1. constLabels | help | buckets| objectives empty
	 * 2. constLabels not empty
	 * 3. help not empty
	 * 4. buckets not empty
	 * 5. objectives not empty
	 */
	metric.MetricType = Counter
	dataProcess(metric, nil)

	metric.ConstLabels = map[string]string{"server": "test"}
	dataProcess(metric, nil)

	metric.Help = "counter help info"
	dataProcess(metric, nil)

	metric.Buckets = buckets
	dataProcess(metric, nil)

	metric.Objectives = objectives
	dataProcess(metric, nil)
}

func TestVecCounter(t *testing.T) {
	/*
	 * vec counter, not nil labels
	 * 1. single labels, not suitable labels value
	 * 2. single labels, suitable labels value
	 * 3. multi labels, not suitable labels value(error key)
	 * 4. multi labels, suitable labels value
	 */
	metric.Labels = []string{"l1"}
	dataProcess(metric, map[string]string{"l2": "l2"})
	dataProcess(metric, map[string]string{"l1": "l1"})

	metric.Labels = []string{"l1", "l2"}
	dataProcess(metric, map[string]string{"l3": "l3"})
	dataProcess(metric, map[string]string{"l2": "l2"})

	dataProcess(metric, map[string]string{"l1": "l1", "l2": "l2"})

	metric.empty()
}

func TestInvalidGauge(t *testing.T) {
	/*
	 * invalid metric conds:
	 * 1. nil metric
	 * 2. empty name
	 * 3. empty metricType
	 */

	metric.MetricType = Gauge
	dataProcess(metric, nil)
}

func TestSingleGauge(t *testing.T) {
	/*
	 * single counter, nil labels
	 * 1. constLabels | help | buckets| objectives empty
	 * 2. constLabels not empty
	 * 3. help not empty
	 * 4. buckets not empty
	 * 5. objectives not empty
	 */
	metric.Name = name
	dataProcess(metric, nil)

	metric.ConstLabels = map[string]string{"server": "test"}
	dataProcess(metric, nil)

	metric.Help = "gauge help info"
	dataProcess(metric, nil)

	metric.Buckets = buckets
	dataProcess(metric, nil)

	metric.Objectives = objectives
	dataProcess(metric, nil)

}

func TestVecGauge(t *testing.T) {
	/*
	 * vec counter, not nil labels
	 * 1. single labels, not suitable labels value
	 * 2. single labels, suitable labels value
	 * 3. multi labels, not suitable labels value(error key)
	 * 4. multi labels, suitable labels value
	 */
	metric.Labels = []string{"l1"}
	dataProcess(metric, map[string]string{"l2": "l2"})

	dataProcess(metric, map[string]string{"l1": "l1"})

	metric.Labels = []string{"l1", "l2"}
	dataProcess(metric, map[string]string{"l3": "l3"})
	dataProcess(metric, map[string]string{"l2": "l2"})

	dataProcess(metric, map[string]string{"l1": "l1", "l2": "l2"})

	metric.empty()
}

func TestInvalidHistogram(t *testing.T) {
	/*
	 * invalid metric conds:
	 * 1. nil metric
	 * 2. empty name
	 * 3. empty metricType
	 */
	metric.MetricType = Histogram
	dataProcess(metric, nil)

}

func TestSingleHistogram(t *testing.T) {
	/*
	 * single counter, nil labels
	 * 1. constLabels | help | buckets| objectives empty
	 * 2. constLabels not empty
	 * 3. help not empty
	 * 4. buckets not empty
	 * 5. objectives not empty
	 */
	metric.Name = name
	dataProcess(metric, nil)

	metric.ConstLabels = map[string]string{"server": "test"}
	dataProcess(metric, nil)

	metric.Help = "histogram help info"
	dataProcess(metric, nil)

	metric.Buckets = buckets
	dataProcess(metric, nil)

	metric.Objectives = objectives
	dataProcess(metric, nil)

}

func TestVecHistogram(t *testing.T) {
	/*
	 * vec counter, not nil labels
	 * 1. single labels, not suitable labels value
	 * 2. single labels, suitable labels value
	 * 3. multi labels, not suitable labels value(error key)
	 * 4. multi labels, suitable labels value
	 */
	metric.Labels = []string{"l1"}
	dataProcess(metric, map[string]string{"l2": "l2"})

	dataProcess(metric, map[string]string{"l1": "l1"})

	metric.Labels = []string{"l1", "l2"}
	dataProcess(metric, map[string]string{"l3": "l3"})
	dataProcess(metric, map[string]string{"l2": "l2"})

	dataProcess(metric, map[string]string{"l1": "l1", "l2": "l2"})

	metric.empty()
}

func TestInvalidSummary(t *testing.T) {
	/*
	 * invalid metric conds:
	 * 1. nil metric
	 * 2. empty name
	 * 3. empty metricType
	 */
	metric.MetricType = Summary
	dataProcess(metric, nil)

}

func TestSingleSummary(t *testing.T) {
	/*
	 * single counter, nil labels
	 * 1. constLabels | help | buckets| objectives empty
	 * 2. constLabels not empty
	 * 3. help not empty
	 * 4. buckets not empty
	 * 5. objectives not empty
	 */
	metric.Name = name
	dataProcess(metric, nil)

	metric.ConstLabels = map[string]string{"server": "test"}
	dataProcess(metric, nil)

	metric.Help = "summary help info"
	dataProcess(metric, nil)

	metric.Buckets = buckets
	dataProcess(metric, nil)

	metric.Objectives = objectives
	dataProcess(metric, nil)
}

func TestVecSummary(t *testing.T) {
	/*
	 * vec counter, not nil labels
	 * 1. single labels, not suitable labels value
	 * 2. single labels, suitable labels value
	 * 3. multi labels, not suitable labels value(error key)
	 * 4. multi labels, suitable labels value
	 */
	metric.Labels = []string{"l1"}
	dataProcess(metric, map[string]string{"l2": "l2"})

	dataProcess(metric, map[string]string{"l1": "l1"})

	metric.Labels = []string{"l1", "l2"}
	dataProcess(metric, map[string]string{"l3": "l3"})
	dataProcess(metric, map[string]string{"l2": "l2"})
	dataProcess(metric, map[string]string{"l1": "l1", "l2": "l2"})

	metric.empty()
}

func TestMain(m *testing.M) {
	Register(port)
	m.Run()
}
