// Copyright 2017 Hewlett Packard Enterprise Development LP
//
//    Licensed under the Apache License, Version 2.0 (the "License"); you may
//    not use this file except in compliance with the License. You may obtain
//    a copy of the License at
//
//         http://www.apache.org/licenses/LICENSE-2.0
//
//    Unless required by applicable law or agreed to in writing, software
//    distributed under the License is distributed on an "AS IS" BASIS, WITHOUT
//    WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the
//    License for the specific language governing permissions and limitations
//    under the License.

package main

import (
	log "github.com/Sirupsen/logrus"
	"os"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"os/signal"
	"syscall"
	"github.hpe.com/UNCLE/monasca-aggregation/models"
	"encoding/json"
	"time"
)

const windowSize = time.Minute/6 // 10 seconds

var aggregationSpecifications = []models.AggregationSpecification{
	{"Aggregation0", "metric0", "aggregated-metric0"},
	{"Aggregation1", "metric1", "aggregated-metric1"},
	{"Aggregation2", "metric2", "aggregated-metric2"},
}

var timedWindowAggregations = map[int64]map[string]float64{}

func initLogging() {
	// Log as JSON instead of the default ASCII formatter.
	log.SetFormatter(&log.JSONFormatter{})
	log.SetOutput(os.Stdout)
	log.SetLevel(log.InfoLevel)
}

func publishAggregations() {
	log.Debug(timedWindowAggregations)
	var previousTimedWindow = int64(time.Now().Unix())/int64(windowSize.Seconds()) - 1
	var windowedAggregations = timedWindowAggregations[previousTimedWindow]
	log.Infof("previousTimedWindow: %d", previousTimedWindow)
	log.Info(windowedAggregations)

	// TODO: Publish the aggreations to Kafka
	// TODO: Advance the Kafka offsets
	// TODO: Delete windowedAggregations for the current window Id that was just published
	// delete(timedWindowAggregations, previousTimedWindow)
}

// TODO: Read in kafka configuration parameters from yaml file
// TODO: Read in aggregation period and aggregation specifications from yaml file
// TODO: Publish aggregated metrics to Kafka
// TODO: Manually update Kafka offsets such that if a crash occurs, processing re-starts at the correct offset
// TODO: Potentially, restrict metrics to a previous, current and next time windowed aggregation.
// TODO: Add support for grouping on dimensions
// TODO: Publish aggregations at window boundaries + lag time. For example, 10 minutes past the hour.
// TODO: Add Prometheus Client library and report metrics
// TODO: Create Helm Charts
// TODO: Add support for consuming/publishing intermediary aggregations. For example, publish a (sum, count) to use in an avg aggregation
// TODO: Guarantee at least once publishing of aggregated metrics
// TODO: Handle start/stop, fail/restart
func main() {
	initLogging()

	if len(os.Args) < 4 {
		log.Errorf("Usage: %s <broker> <group> <topics..>", os.Args[0])
	}

	broker := os.Args[1]
	group := os.Args[2]
	topics := os.Args[3:]

	sigchan := make(chan os.Signal)
	signal.Notify(sigchan, syscall.SIGINT, syscall.SIGTERM)

	c, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers":               broker,
		"group.id":                        group,
		"session.timeout.ms":              6000,
		"go.events.channel.enable":        true,
		"go.application.rebalance.enable": true,
		"default.topic.config":            kafka.ConfigMap{"auto.offset.reset": "earliest"}})

	if err != nil {
		log.Fatalf("Failed to create consumer: %s", err)
	}

	log.Infof("Started monasca-aggregation %v", c)

	err = c.SubscribeTopics(topics, nil)

	aggregationTicker := time.NewTicker(windowSize)

	run := true

	for run == true {
		select {
		case sig := <-sigchan:
			log.Infof("Caught signal %v: terminating", sig)
			run = false

		case ev := <-c.Events():
			switch e := ev.(type) {
			case kafka.AssignedPartitions:
				log.Infof("%% %v", e)
				c.Assign(e.Partitions)
			case kafka.RevokedPartitions:
				log.Infof("%% %v", e)
				c.Unassign()
			case *kafka.Message:
				metricEnvelope := models.MetricEnvelope{}
				err = json.Unmarshal([]byte(e.Value), &metricEnvelope)
				if err != nil {
					log.Warnf("%% Invalid metric envelope on %s:%s", e.TopicPartition, string(e.Value))
					continue
				}
				var metric = metricEnvelope.Metric
				var eventTimedWindow = metric.Timestamp/(1000*int64(windowSize.Seconds()))

				for _, aggregationSpecification := range aggregationSpecifications {
					if metric.Name == aggregationSpecification.FilteredMetricName {
						var windowAggregations = timedWindowAggregations[eventTimedWindow]
						if windowAggregations == nil {
							timedWindowAggregations[eventTimedWindow] = make(map[string]float64)
							windowAggregations = timedWindowAggregations[eventTimedWindow]
						}
						windowAggregations[aggregationSpecification.AggregatedMetricName] += metric.Value
					}
				}
				log.Debug(metricEnvelope)
			case kafka.PartitionEOF:
				//log.Infof("%% Reached %v", e)
			case kafka.Error:
				log.Errorf("%% Error: %v", e)
				run = false
			}

		case <-aggregationTicker.C:
			publishAggregations()
		}
	}

	log.Info("Stopped monasca-aggregation")
	c.Close()
}
