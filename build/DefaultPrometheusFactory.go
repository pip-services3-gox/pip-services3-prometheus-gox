package build

import (
	cref "github.com/pip-services3-gox/pip-services3-commons-gox/refer"
	"github.com/pip-services3-gox/pip-services3-components-gox/build"
	cbuild "github.com/pip-services3-gox/pip-services3-components-gox/build"
	pcount "github.com/pip-services3-gox/pip-services3-prometheus-gox/count"
	pservices "github.com/pip-services3-gox/pip-services3-prometheus-gox/services"
)

// DefaultPrometheusFactory creates Prometheus components by their descriptors.
// See: Factory
// See: PrometheusCounters
// See: PrometheusMetricsService
type DefaultPrometheusFactory struct {
	*cbuild.Factory
}

// NewDefaultPrometheusFactory are create a new instance of the factory.
func NewDefaultPrometheusFactory() *DefaultPrometheusFactory {
	c := DefaultPrometheusFactory{}
	c.Factory = build.NewFactory()

	prometheusCountersDescriptor := cref.NewDescriptor("pip-services", "counters", "prometheus", "*", "1.0")
	prometheusMetricsServiceDescriptor := cref.NewDescriptor("pip-services", "metrics-service", "prometheus", "*", "1.0")

	c.RegisterType(prometheusCountersDescriptor, pcount.NewPrometheusCounters)
	c.RegisterType(prometheusMetricsServiceDescriptor, pservices.NewPrometheusMetricsService)
	return &c
}
