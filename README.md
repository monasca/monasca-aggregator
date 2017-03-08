# monasca-aggregation

An aggregation micro-service for Monasca with the following features:

* Read metrics from Kafka
* Write metrics to Kafka
* Filter metrics by metric name
* Filter metrics by dimension name/values
* Group metrics by dimension keys
* Write aggregated metrics with a new aggregated name
* Arbitrary time window aggregations. Support any time period specifiable, such as 10 seconds or one hour.
* Time windows are aligned to the epoch. For example, a one hour aggregation window size will start at the top of the hour, not in the middle.
* Arbitrary lag times. The time at which the aggregation starts is set based on a "lag" time past the end of the collection period. For example, 10 minutes past the hour. This can be set to any value, such as 10 hours if desired.
* Continuous aggregations. Totals are stored in memory, not metrics. Therefore the metrics don't need to be pulled into memory and operated on in a batch.
* Event time window processing. Aggregations for metrics are performed based on the timestamp of the metric in event time, time it was created, and not the process time, time at which it is being processed.
* Stop/starts, crash/restarts handling. Kafka offsets are manually stored and set on start-up to the proper point in-time to allow processing to start from the last successful aggregation. Therefore, no data loss.
* A nice DSL for describing aggregations in a config.yaml. See, https://github.hpe.com/UNCLE/monasca-aggregation/blob/master/config.yaml
* Performance. > 50K metrics/sec, but I'm not sure how high. It is possible it is greater than 100K, but I'll need a different testing strategy to verify.
* Testable.
* Instantaneous start-up times.
* Low cpu and memory footprint.
* Written in Go.
