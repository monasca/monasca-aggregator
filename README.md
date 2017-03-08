# monasca-aggregation

A high-speed aggregation micro-service for Monasca with the following features:

* Read metrics from Kafka.

* Write aggregated metrics to Kafka.

* Filter metrics by metric name.

* Filter metrics by dimension (name, value) pairs.

* Group metrics by dimension names.

* Write aggregated metrics with a specified aggregated name.

* Aggregate on specified window sizes. Support any window size. E.g. 10 seconds or one hour.

* Time window aligned. Time window aggregations occur at boundaries aligned at the start of the epoch. E.g. if a one hour aggregation window size is specified it will start at the top of the hour, not randomly in the middle.

* Lag time. Produce aggregations at a specified lag time from the end of the time window. The time at which the aggregation starts is set based on a "lag" time that is the duration past the end of the time window. For example, 10 minutes past the hour. This can be set to any value, such as 10 hours if desired.

* Continuous aggregations. Totals are stored in memory, not metrics. Therefore the metrics don't need to be pulled into memory and operated on in a batch.

* Event time window processing. Aggregations for metrics are processed based on the timestamp of the metric in event time, and not the process time or time at which it is being processed.

* Stop/start, crash/restarts handling. Kafka offsets are manually stored and set on start-up to the proper point in-time to allow processing to start from the last successful aggregation. Therefore, aggregations are computed with no data loss.

* DSL. A nice DSL for describing aggregations in a config.yaml. See, https://github.hpe.com/UNCLE/monasca-aggregation/blob/master/config.yaml.

* Performance. > 50K metrics/sec, but we're not exactly sure how fast it is. It is possible it is greater than 100K metrics/sec, but we'll need a different testing strategy to verify.

* Testable. Due to it's lightweight design and footprint, as well as ability to specify small windows sizes, it is very easy to test.

* Instantaneous start-up times. Due to it's lightweight design and use of Go, start-up times are extremely fast.

* Low cpu and memory footprint. Since processing is continuous and only the aggregations are stored in memory, the memory footprint is very small.

* Written in Go.

* Instrumented using the [Prometheus Go Client Library](https://github.com/prometheus/client_golang) and [https://github.com/sirupsen/logrus](https://github.com/sirupsen/logrus).

* Configured using [Viper](https://github.com/spf13/viper). Viper supports many configuration options, but we use it for yaml config files.
