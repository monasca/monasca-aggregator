## Prerequisites

The aggregation engine requires Apache Kafka 0.9.x or higher and at least one topic
By default, the aggregation engine will look for a topic called ```metrics```
but it can be configured to use any topic.
[Kafka quick start](https://kafka.apache.org/090/documentation.html#quickstart)

Kafka requires a running instance of Apache Zookeeper.
[Zookeeper startup guide](https://zookeeper.apache.org/doc/trunk/zookeeperStarted.html)

## Installing
You can use the ```go build``` command to create a binary of the aggregation engine
if desired. Otherwise, ```go run server.go``` in the top directory of the repo
will compile and run the engine.

## Configuring the Engine
The engine expects a config file called ```config.yaml``` in the same directory it is running. If
running the engine from the repo, a config file is already present at the top directory.

## Configuring the aggregations
See aggregations.md in the docs directory for more information about configuring
aggregations.

## Running the aggregation engine
With either the binary produced from ```go build``` or the command ```go run server.go```
start the aggregation engine. It should log some basic startup information (log level
must be INFO or lower to see this). To verify the engine is running, you can check its metrics
by curling the metric endpoint, configured at ```localhost:8080/metrics``` by default.

## Verifying the Engine
For convenience, a generic kafka producer is included in the ```tools``` directory of
the repo. Use it by running ```go run publisher.go localhost:9092 metrics``` substituting
your own kafka address and topic if necessary. This will produce a metric that matches
one of the default aggregations, so the engine should output one or more aggregated metrics.
If no aggregations are configured, there will be no output.

You can watch a kafka topic with the console consumer
```{kafka install directory}/bin/kafka-console-consumer.sh --zookeeper localhost:2181 --topic metrics```
substituting your own location of zookeeper and topic if necessary.