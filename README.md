# Prom-generator

Prom-generator is an automated generator for Prometheus metrics and exemplars. You can use this to generate load,
test performance or feature implementation for your Prometheus infrastructure, optionally with a remote-storage system.

Note:
1. Runs on `http://localhost:9001`
2. Telemetry endpoint is `/metrics`
3. If you want to see exemplar, you will have to provide `"Accept: application/openmetrics-text"` as header in the GET
   request to `/metrics` endpoint
   
```shell
[hsingh@localhost exemplars-generator]$ ./prom-generator -h
Usage of ./prom-generator:
  -evaluate-every duration
        Frequency of evaluation of metrics and exemplar. (default 1s)
  -num-counters int
        Number of counters to be generated. (default 1)
  -num-counters-with-exemplars int
        Number of counters to be generated with exemplars. (default 1)
  -num-gauges int
        Number of gauges to be generated. (default 1)
  -num-histograms int
        Number of histograms to be generated. (default 1)
  -num-histograms-with-exemplars int
        Number of histograms to be generated with exemplars. (default 1)

```