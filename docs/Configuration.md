# Configuration Guide <br/>

Configuration structure follows the 
[standard configuration](https://github.com/pip-services/pip-services3-container-node/doc/Configuration.md) 
structure. 

### <a name="count_prometheus"></a> Prometheus

Prometheus counters has the following configuration properties:
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

Example:
```yaml
- descriptor: "pip-services:counters:prometheus:default:1.0"
  source: "test"
  connection:
    protocol: "http"
    host: "localhost"
    port: 8080
```

Prometheus counters service has the following configuration properties:
- dependencies:
  - endpoint:              override for HTTP Endpoint dependency
  - prometheus-counters:   override for PrometheusCounters dependency
- connection(s):
  - discovery_key:         (optional) a key to retrieve the connection from IDiscovery
  - protocol:              connection protocol: http or https
  - host:                  host name or IP address
  - port:                  port number
  - uri:                   resource URI or connection string with all parameters in it

Example:
```yaml
- descriptor: "pip-services:service:prometheus:default:1.0"
  connection:
    protocol: "http"
    host: "localhost"
    port: 8080
```

For more information on this section read 
[Pip.Services Configuration Guide](https://github.com/pip-services/pip-services3-container-node/doc/Configuration.md#deps)