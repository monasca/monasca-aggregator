# monasca-aggregation

![Monasca](https://photos-3.dropbox.com/t/2/AABUtCKREPgNoxyDzwPS9R2zACBW2i8lhO3QRHykGthlvw/12/442004266/png/32x32/3/1489042800/0/2/OpenStack_Project_Monasca_mascot.png/EPrB40IYj54HIAIoAg/nxOqRLZpYcMepsb_bitQEmj0VGCfwUnnwTOP-tjNFGs?dl=0&size=1600x1200&size_mode=3)

## Introduction

A high-speed near real-time continuous aggregation micro-service for Monasca with the following features:

* Read metrics from Kafka.

* Write aggregated metrics to Kafka.

* Filter metrics by metric name.

* Filter metrics by dimension (name, value) pairs.

* Group by metric dimension names.

* Write aggregated metrics using a specified aggregated name.

* Supported aggregations include the following:
 
  * sum
  
  * others will be supported in the near future

* Aggregate on specified window sizes. Support any window size. E.g. 10 seconds or one hour.

* Aggregations aligned to time window boundaries.
Time window aggregations occur at boundaries aligned to the start of the epoch.
E.g. If a one hour window size is specified, time window aggregations will start on the hour, not randomly in the middle based on when the process is started.

* Lag time. Aggregations are produced at a specified lag time past the end of the time window.
The time at which the aggregations start is specified based on a "lag" time, which is the duration past the end of the time window.
E.g. 10 minutes past the hour. This can be set to any value, such as 10 hours if desired.

* Continuous near real-time aggregations.
Aggregations are stored in memory only.
Therefore, metrics don't need to be pulled into memory and operated on in a batch operation.
E.g. When perfoming a sum operation for a series only the running total is kept in memory for each series.

* Event time window processing.
Aggregations for metrics are processed based on the timestamp of the metric in event time, and not the process time or time at which the metric is being processed.

* Stop/start, crash/restarts handling.
Kafka offsets are manually committed after an aggregation is produced to allow processing to start off from where the last successful aggregation completed.
Therefore, aggregations are computed with no data loss.
If for any reason  processing stops in the middle of a time window the Kafka offsets will not be committed for that time window.
When re-started, the Kafka offsets are read from Kafka and processing starts off from the last succesful commit.
This implies that metrics may be read from Kafka multiple times in the event of a re-start, but there is no data loss.

* Domain Specific Language (DSL).
A simple expressive DSL for specifying aggregations.
See, [aggregation-specifications.yaml](aggregation-specifications.yaml).

* Performance. > 50K metrics/sec, but we're not exactly sure how fast it is.
It is possible it is greater than 100K metrics/sec, but we'll need a different testing strategy to verify.

* Written in Go.

* Dependencies: Dependent on only the following Go libraries:

  * [Confluent's Apache Kafka client for Golang](https://github.com/confluentinc/confluent-kafka-go)

  * [Prometheus Go Client Library](https://github.com/prometheus/client_golang)

  * [logrus](https://github.com/sirupsen/logrus)

  * [Viper](https://github.com/spf13/viper)

* No additional runtime requirements, beyond Apache Kafka, such as Apache Spark and Apache Storm.
In addition, no additional databases required.
For example, Kafka offsets are stored in Kafka and do not require an external database, such as MySQL.

* Instantaneous start-up times.
Due to it's lightweight design and use of Go, start-up times are extremely fast.

* Easily deployed and configured.
Due to the use of Go and small set of dependencies, can be easily deployed.

* Low cpu and memory footprint.
Since processing is continuous and only the aggregations are stored in memory, such as the sum, the memory footprint is very small.

* Testable.
Due to it's lightweight design and footprint, as well as ability to specify small windows sizes, it is very easy to test.
For example, when testing it is possible to aggregate with 10 second window sizes.
In addition, due to Go and a small set of dependencies, it is possible to run monasca-aggregation on a laptop without any additional runtime environment, other than Kafka.

* Instrumented using the [Prometheus Go Client Library](https://github.com/prometheus/client_golang) and [logrus](https://github.com/sirupsen/logrus).

* Configured using [Viper](https://github.com/spf13/viper).
Viper supports many configuration options, but we use it for yaml config files.
See [config.yaml](config.yaml) and [aggregation-specifications.yaml](aggregation-specifications.yaml)

## References

Several of the concepts, such as time windows, continuous aggregations, event time processing, are best described in the following references.

### Kafka Streams

Although Kafka Stream isn't used here, it serves as excellent background on stream processing.
One of the main concepts that Kafka Streams introduces is a time windowed key/value store that can be used to store aggregations.
If used wisely, this can help address more complicated scenarios, such as fail/re-start, without having to manually manage and commit Kafka offsets.
Kafka Streams is a really exciting technology, but we didn't use it here, as it is only available in Java.
However, several of the concepts in Kafka Streams are used here.
Hopefully, Kafka Streams is ported to Go someday.

* [Introducing Kafka Streams: Stream Processing Made Simple](https://www.confluent.io/blog/introducing-kafka-streams-stream-processing-made-simple/)

* [Introduction to Streaming Data and Stream Processing with Apache Kafka](https://www.confluent.io/apache-kafka-talk-series/introduction-to-stream-processing-with-apache-kafka/})

* [Kafka Streams](http://docs.confluent.io/3.0.0/streams/)

### Google and Apache Beam

Although Apache Beam isn't used here, Tyler Akidau et al's seminal paper, which led to the Apache Beam project, is an excellent reference for understanding event and process time windowing.

* [The Dataflow Model: A Practical Approach to Balancing
 Correctness, Latency, and Cost in Massive-Scale,
 Unbounded, Out-of-Order Data Processing](http://www.vldb.org/pvldb/vol8/p1792-Akidau.pdf)
 
* [The world beyond batch: Streaming 101](https://www.oreilly.com/ideas/the-world-beyond-batch-streaming-101)
 
* [The world beyond batch: Streaming 102](https://www.oreilly.com/ideas/the-world-beyond-batch-streaming-102)

* [MillWheel: Fault-Tolerant Stream Processing at Internet Scale](https://research.google.com/pubs/pub41378.html)

### Misc
 
* [Building Scalable Stateful Services by Caitie McCaffrey](https://www.youtube.com/watch?v=H0i_bXKwujQ&feature=youtu.be&a)