package test_count

import (
	"os"
	"testing"

	cconf "github.com/pip-services3-gox/pip-services3-commons-gox/config"
	pcount "github.com/pip-services3-gox/pip-services3-prometheus-gox/count"
	pfixture "github.com/pip-services3-gox/pip-services3-prometheus-gox/test/fixture"
)

func TestPrometheusCounters(t *testing.T) {
	var counters *pcount.PrometheusCounters
	var fixture *pfixture.CountersFixture

	host := os.Getenv("PUSHGATEWAY_SERVICE_HOST")
	if host == "" {
		host = "localhost"
	}
	port := os.Getenv("PUSHGATEWAY_SERVICE_PORT")
	if port == "" {
		port = "9091"
	}
	counters = pcount.NewPrometheusCounters()
	fixture = pfixture.NewCountersFixture(counters.CachedCounters)

	config := cconf.NewConfigParamsFromTuples(
		"source", "test",
		"connection.host", host,
		"connection.port", port,
		"connection.protocol", "http",
	)
	counters.Configure(config)

	counters.Open("")

	defer counters.Close("")

	t.Run("Simple Counters", fixture.TestSimpleCounters)
	t.Run("Measure Elapsed Time", fixture.TestMeasureElapsedTime)
}
