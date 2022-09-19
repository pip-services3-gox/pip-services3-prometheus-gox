package services

import (
	"io"
	"net/http"

	cref "github.com/pip-services3-gox/pip-services3-commons-gox/refer"
	ccount "github.com/pip-services3-gox/pip-services3-components-gox/count"
	cinfo "github.com/pip-services3-gox/pip-services3-components-gox/info"
	pcount "github.com/pip-services3-gox/pip-services3-prometheus-gox/count"
	rpcservices "github.com/pip-services3-gox/pip-services3-rpc-gox/services"
)

/*
PrometheusMetricsService is service that exposes "/metrics" route for Prometheus to scap performance metrics.

Configuration parameters:

  - dependencies:
    - endpoint:              override for HTTP Endpoint dependency
    - prometheus-counters:   override for PrometheusCounters dependency
  - connection(s):
    - discovery_key:         (optional) a key to retrieve the connection from IDiscovery
    - protocol:              connection protocol: http or https
    - host:                  host name or IP address
    - port:                  port number
    - uri:                   resource URI or connection string with all parameters in it

References:

- *:logger:*:*:1.0         (optional)  ILogger components to pass log messages
- *:counters:*:*:1.0         (optional)  ICounters components to pass collected measurements
- *:discovery:*:*:1.0        (optional)  IDiscovery services to resolve connection
- *:endpoint:http:*:1.0          (optional)  HttpEndpoint reference to expose REST operation
- *:counters:prometheus:*:1.0    PrometheusCounters reference to retrieve collected metrics

See RestService
See RestClient

 Example

    service := NewPrometheusMetricsService();
    service.Configure(cconf.NewConfigParamsFromTuples(
        "connection.protocol", "http",
        "connection.host", "localhost",
        "connection.port", "8080",
    ));

    err := service.Open("123")
    if  err == nil {
        fmt.Println("The Prometheus metrics service is accessible at http://localhost:8080/metrics");
        defer service.Close("")
    }
*/
type PrometheusMetricsService struct {
	rpcservices.RestService
	cachedCounters *ccount.CachedCounters
	source         string
	instance       string
}

// NewPrometheusMetricsService are creates a new instance of c service.
// Returns *PrometheusMetricsService
// pointer on new instance
func NewPrometheusMetricsService() *PrometheusMetricsService {
	c := &PrometheusMetricsService{}
	c.RestService = *rpcservices.InheritRestService(c)
	c.DependencyResolver.Put("cached-counters", cref.NewDescriptor("pip-services", "counters", "cached", "*", "1.0"))
	c.DependencyResolver.Put("prometheus-counters", cref.NewDescriptor("pip-services", "counters", "prometheus", "*", "1.0"))
	return c
}

// SetReferences is sets references to dependent components.
// Parameters:
//   - references cref.IReferences
// references to locate the component dependencies.
func (c *PrometheusMetricsService) SetReferences(references cref.IReferences) {
	c.RestService.SetReferences(references)

	resolv := c.DependencyResolver.GetOneOptional("prometheus-counters")
	c.cachedCounters = resolv.(*pcount.PrometheusCounters).CachedCounters
	if c.cachedCounters == nil {
		resolv = c.DependencyResolver.GetOneOptional("cached-counters")
		c.cachedCounters = resolv.(*ccount.CachedCounters)
	}
	ref := references.GetOneOptional(
		cref.NewDescriptor("pip-services", "context-info", "default", "*", "1.0"))
	contextInfo := ref.(*cinfo.ContextInfo)

	if contextInfo != nil && c.source == "" {
		c.source = contextInfo.Name
	}
	if contextInfo != nil && c.instance == "" {
		c.instance = contextInfo.ContextId
	}
}

// Register method are registers all service routes in HTTP endpoint.
func (c *PrometheusMetricsService) Register() {
	c.RegisterRoute("get", "metrics", nil, func(res http.ResponseWriter, req *http.Request) { c.metrics(res, req) })
}

// Handles metrics requests
//   - req   an HTTP request
//   - res   an HTTP response
func (c *PrometheusMetricsService) metrics(res http.ResponseWriter, req *http.Request) {

	var counters []*ccount.Counter
	if c.cachedCounters != nil {
		counters = c.cachedCounters.GetAll()
	}

	body := pcount.PrometheusCounterConverter.ToString(counters, c.source, c.instance)

	res.Header().Add("content-type", "text/plain")
	res.WriteHeader(200)
	_, wrErr := io.WriteString(res, (string)(body))
	if wrErr != nil {
		c.Logger.Error("PrometheusMetricsService", wrErr, "Can't write response")
	}
}
