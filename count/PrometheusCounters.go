package count

import (
	"bytes"
	"net/http"
	"os"
	"time"

	cconf "github.com/pip-services3-gox/pip-services3-commons-gox/config"
	cerr "github.com/pip-services3-gox/pip-services3-commons-gox/errors"
	cref "github.com/pip-services3-gox/pip-services3-commons-gox/refer"
	ccount "github.com/pip-services3-gox/pip-services3-components-gox/count"
	cinfo "github.com/pip-services3-gox/pip-services3-components-gox/info"
	clog "github.com/pip-services3-gox/pip-services3-components-gox/log"
	rpcconnect "github.com/pip-services3-gox/pip-services3-rpc-gox/connect"
)

/*
PrometheusCounters performance counters that send their metrics to Prometheus service.

The component is normally used in passive mode conjunction with PrometheusMetricsService.
Alternatively when connection parameters are set it can push metrics to Prometheus PushGateway.

Configuration parameters:

  - connection(s):
    - discovery_key:         (optional) a key to retrieve the connection from connect.idiscovery.html IDiscovery
    - protocol:              connection protocol: http or https
    - host:                  host name or IP address
    - port:                  port number
    - uri:                   resource URI or connection string with all parameters in it
  - options:
    - retries:               number of retries (default: 3)
    - connect_timeout:       connection timeout in milliseconds (default: 10 sec)
    - timeout:               invocation timeout in milliseconds (default: 10 sec)

References:

- *:logger:*:*:1.0         (optional) ILogger components to pass log messages
- *:counters:*:*:1.0         (optional) ICounters components to pass collected measurements
- *:discovery:*:*:1.0        (optional)  IDiscovery services to resolve connection

See:  RestService
See:  CommandableHttpService

Example:

    counters = NewPrometheusCounters();
    counters.Configure(cconf.NewConfigParamsFromTuples(
        "connection.protocol", "http",
        "connection.host", "localhost",
        "connection.port", 8080
    ));

	counters.Open("123")

    counters.Increment("mycomponent.mymethod.calls");
    timing := counters.BeginTiming("mycomponent.mymethod.exec_time");
        ...
    timing.EndTiming();

    counters.Dump();
*/
type PrometheusCounters struct {
	*ccount.CachedCounters
	logger             *clog.CompositeLogger
	connectionResolver *rpcconnect.HttpConnectionResolver
	opened             bool
	source             string
	instance           string
	client             *http.Client
	requestRoute       string
	timeout            int
	retries            int
	connectTimeout     int
	uri                string
}

// NewPrometheusCounters is creates a new instance of the performance counters.
// Returns *PrometheusCounters
// pointer on new instance
func NewPrometheusCounters() *PrometheusCounters {
	c := PrometheusCounters{}
	c.CachedCounters = ccount.InheritCacheCounters(&c)
	c.logger = clog.NewCompositeLogger()
	c.connectionResolver = rpcconnect.NewHttpConnectionResolver()
	c.opened = false
	c.timeout = 10000
	c.retries = 3
	c.connectTimeout = 10000
	return &c
}

// Configure method are configures component by passing configuration parameters.
// Parameters:
// - config   *cconf.ConfigParams
// configuration parameters to be set.
func (c *PrometheusCounters) Configure(config *cconf.ConfigParams) {
	c.CachedCounters.Configure(config)
	c.connectionResolver.Configure(config)

	c.source = config.GetAsStringWithDefault("source", c.source)
	c.instance = config.GetAsStringWithDefault("instance", c.instance)
	c.retries = config.GetAsIntegerWithDefault("options.retries", c.retries)
	c.connectTimeout = config.GetAsIntegerWithDefault("options.connectTimeout", c.connectTimeout)
	c.timeout = config.GetAsIntegerWithDefault("options.timeout", c.timeout)
}

// SetReferences method are sets references to dependent components.
// - references  cref.IReferences
// references to locate the component dependencies.
func (c *PrometheusCounters) SetReferences(references cref.IReferences) {
	c.logger.SetReferences(references)
	c.connectionResolver.SetReferences(references)
	ref := references.GetOneOptional(
		cref.NewDescriptor("pip-services", "context-info", "default", "*", "1.0"))
	contextInfo, _ := ref.(*cinfo.ContextInfo)
	if contextInfo != nil && c.source == "" {
		c.source = contextInfo.Name
	}
	if contextInfo != nil && c.instance == "" {
		c.instance = contextInfo.ContextId
	}
}

// IsOpen method are checks if the component is opened.
// Returns true if the component has been opened and false otherwise.
func (c *PrometheusCounters) IsOpen() bool {
	return c.opened
}

// Open method are opens the component.
// - correlationId 	string
// (optional) transaction id to trace execution through call chain.
// Returns error
//	error or nil, if no errors occured.
func (c *PrometheusCounters) Open(correlationId string) (err error) {
	if c.opened {
		return nil
	}

	c.opened = true
	connection, _, err := c.connectionResolver.Resolve(correlationId)

	if err != nil {
		c.client = nil
		c.logger.Warn(correlationId, "Connection to Prometheus server is not configured: "+err.Error())
		return nil
	}

	c.uri = connection.Uri()

	job := c.source
	if job == "" {
		job = "unknown"
	}

	instance := c.instance
	if instance == "" {
		host, _ := os.Hostname()
		instance = host
	}
	c.requestRoute = "/metrics/job/" + job + "/instance/" + instance

	localClient := http.Client{}
	localClient.Timeout = (time.Duration)(c.timeout) * time.Millisecond
	c.client = &localClient
	if c.client == nil {
		ex := cerr.NewConnectionError(correlationId, "CANNOT_CONNECT", "Connection to REST service failed").WithDetails("url", c.uri)
		return ex
	}

	return nil
}

// Close method are closes component and frees used resources.
// Parameters:
// - correlationId string
//	(optional) transaction id to trace execution through call chain.
// Returns error
// error or nil, if no errors occured.
func (c *PrometheusCounters) Close(correlationId string) error {
	c.opened = false
	c.client = nil
	c.requestRoute = ""
	return nil
}

// Save method are saves the current counters measurements.
// - counters   []*ccount.Counter
//    current counters measurements to be saves.
// Retruns error
// error or nil, if no errors occured.
func (c *PrometheusCounters) Save(counters []*ccount.Counter) (err error) {
	if c.client == nil {
		return nil
	}

	url := c.uri + c.requestRoute
	body := PrometheusCounterConverter.ToString(counters, "", "")

	req, reqErr := http.NewRequest(http.MethodPut, url, bytes.NewBuffer([]byte(body)))
	if reqErr != nil {
		err = cerr.NewUnknownError("PrometheusCounters", "UNSUPPORTED_METHOD", "Method is not supported by REST client").WithDetails("verb", "PUT").WithCause(reqErr)
		return err
	}
	// Set headers
	req.Header.Set("Accept", "text/html")
	retries := c.retries
	var resp *http.Response
	var respErr error

	for retries > 0 {
		// Try send request
		resp, respErr = c.client.Do(req)
		if respErr != nil {

			retries--
			if retries == 0 {
				err = cerr.NewUnknownError("PrometheusCounters", "COMMUNICATION_ERROR", "Unknown communication problem on REST client").WithCause(respErr)
				return err
			}
			continue

		}
		break
	}
	if resp != nil {
		defer resp.Body.Close()
		if resp.StatusCode >= 204 && resp.StatusCode < 300 {
			return nil
		}
		c.logger.Error("prometheus-counters", respErr, "Failed to push metrics to prometheus")
	}

	return respErr
}
