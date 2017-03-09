# monasca-aggregation ![Monasca](https://www.dropbox.com/sh/yigfsgvpzxz4t72/AABHPm5FOoBFv2Q-div_j9RXa?dl=0&preview=OpenStack_Project_Monasca_mascot.png)

A high-speed near real-time aggregation micro-service for Monasca with the following features:

* Read metrics from Kafka.

* Write aggregated metrics to Kafka.

* Filter metrics by metric name.

* Filter metrics by dimension (name, value) pairs.

* Group metrics by dimension names.

* Write aggregated metrics with a specified aggregated name.

* Supported aggregations include the following:
 
  * sum
  
  * others will be supported in the near future

* Aggregate on specified window sizes. Support any window size. E.g. 10 seconds or one hour.

* Aggregations aligned to time window boundaries.
Time window aggregations occur at boundaries aligned to the start of the epoch.
E.g. if a one hour aggregation window size is specified aggregations will start on the hour, not randomly in the middle.

* Lag time. Produce aggregations at a specified lag time past the end of the time window.
The time at which the aggregations start is set based on a "lag" time, which is the duration past the end of the time window.
For example, 10 minutes past the hour. This can be set to any value, such as 10 hours if desired.

* Continuous near real-time aggregations.
Aggregations are stored in memory, not metrics.
Therefore the metrics don't need to be pulled into memory and operated on in a batch.

* Event time window processing.
Aggregations for metrics are processed based on the timestamp of the metric in event time, and not the process time or time at which it is being processed.

* Stop/start, crash/restarts handling.
Kafka offsets are manually committed after an aggregation is produced to allow processing to start from where the last successful aggregation ended. Therefore, aggregations are computed with no data loss.

* Domain Specific Language (DSL).
A nice DSL for specifying aggregations.
See, [aggregations.yaml](aggregations.yaml).

* Performance. > 50K metrics/sec, but we're not exactly sure how fast it is.
It is possible it is greater than 100K metrics/sec, but we'll need a different testing strategy to verify.

* Testable.
Due to it's lightweight design and footprint, as well as ability to specify small windows sizes, it is very easy to test. For example, for testing purpose it is possible to aggregate with 10 second window sizes.

* Written in Go.

* Dependencies:

  * Dependent on only a few Go libraries as follows:
  
    * [Confluent's Apache Kafka client for Golang](https://github.com/confluentinc/confluent-kafka-go)
    
    * [Prometheus Go Client Library](https://github.com/prometheus/client_golang)
    
    * [logrus](https://github.com/sirupsen/logrus)
    
    * [Viper](https://github.com/spf13/viper)
  
  * No additional runtime requirements, beyond Apache Kafka, such as Apache Spark and Apache Storm.
  
  * No additional databases required. For example, Kafka offsets are stored in Kafka and does not require an external database, such as MySQL.

* Instantaneous start-up times.
Due to it's lightweight design and use of Go, start-up times are extremely fast.

* Easily deployed and configured.
Due to the use of Go and small set of dependencies, can be easily deployed.

* Low cpu and memory footprint.
Since processing is continuous and only the aggregations are stored in memory, such as the sum, the memory footprint is very small.

* Instrumented using the [Prometheus Go Client Library](https://github.com/prometheus/client_golang) and [logrus](https://github.com/sirupsen/logrus).

* Configured using [Viper](https://github.com/spf13/viper).
Viper supports many configuration options, but we use it for yaml config files.
See [config.yaml](config.yaml)
