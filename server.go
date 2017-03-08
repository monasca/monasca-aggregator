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
	"encoding/json"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	log "github.com/Sirupsen/logrus"

	"github.hpe.com/UNCLE/monasca-aggregation/models"
	"github.com/spf13/viper"
)

var windowSize time.Duration
var windowLag time.Duration
var aggregationSpecifications []models.AggregationSpecification
var timeWindowAggregations = map[int64]map[string]models.Metric{}
var offsetCache = map[int64]map[int32]int64{}

func initLogging() {
	// Log as JSON instead of the default ASCII formatter.
	log.SetFormatter(&log.JSONFormatter{})
	log.SetOutput(os.Stdout)
	log.SetLevel(log.InfoLevel)
}

func initConfig() {
	viper.SetDefault("windowSize", 10)
	viper.SetDefault("windowLag", 2)
	viper.SetDefault("consumerTopic", "metrics")
	viper.SetDefault("producerTopic", "metrics")
	viper.SetDefault("kafka.bootstrap.servers", "localhost:9092")
	viper.SetDefault("kafka.group.id", "monasca-aggregation")
	viper.SetConfigName("config")
	viper.AddConfigPath(".")
	err := viper.ReadInConfig()

	if err != nil {
		log.Fatalf("Fatal error config file: %s \n", err)
	}
}

func initConsumer(consumerTopic, groupId, bootstrapServers string) *kafka.Consumer {
	c, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers":               bootstrapServers,
		"group.id":                        groupId,
		"session.timeout.ms":              6000,
		"go.events.channel.enable":        true,
		"go.application.rebalance.enable": true,
		"enable.auto.commit":		   false,
		"default.topic.config":            kafka.ConfigMap{"auto.offset.reset": "earliest"},
	})

	if err != nil {
		log.Fatalf("Failed to create consumer: %s", err)
	}

	log.Infof("Started monasca-aggregation %v", c)

	err = c.Subscribe(consumerTopic, nil)

	if err != nil {
		log.Fatalf("Failed to subscribe to topics %c", err)
	}
	log.Infof("Subscribed to topic %s as group %s", consumerTopic, groupId)

	return c
}

func initProducer(bootstrapServers string) *kafka.Producer {
	p, err := kafka.NewProducer(&kafka.ConfigMap{"bootstrap.servers": bootstrapServers})

	if err != nil {
		log.Fatalf("Failed to create producer: %s", err)
	}

	log.Infof("Created Producer %v", p)

	return p
}

func handleProducerEvents(p *kafka.Producer) {
	outer:
	for e := range p.Events() {
		switch ev := e.(type) {
		case *kafka.Message:
			m := ev
			if m.TopicPartition.Error != nil {
				log.Errorf("Delivery failed: %v\n", m.TopicPartition.Error)
			} else {
				log.Infof("Delivered message to topic %s [%d] at offset %v\n",
					*m.TopicPartition.Topic, m.TopicPartition.Partition, m.TopicPartition.Offset)
			}
			break outer

		default:
			log.Debugf("Ignored event: %s\n", ev)
		}
	}
}

// Return a timer for when the first window should be processed
// TODO: Check this math to account for all boundary conditions and large lag times
func firstTick() *time.Timer {
	now := time.Now().Unix()
	completed := now % int64(windowSize.Seconds()) - int64(windowLag.Seconds())
	remaining := int64(windowSize.Seconds()) - completed
	firstTick := time.NewTimer(time.Duration(remaining * 1e9))
	return firstTick
}

func publishAggregations(outbound chan *kafka.Message, topic *string, c *kafka.Consumer) {
	// TODO make timestamp assignment the beginning of the aggregation window
	log.Debug(timeWindowAggregations)
	var currentTimeWindow = int64(time.Now().Unix()) / int64(windowSize.Seconds())
	var windowLagCount = int64(windowLag.Seconds() / windowSize.Seconds()) - 1
	var activeTimeWindow = currentTimeWindow + windowLagCount
	log.Infof("currentTimeWindow: %d", currentTimeWindow)
	log.Infof("activeTimeWindow: %d", activeTimeWindow)
	log.Info(timeWindowAggregations)

	for t, windowAggregations := range timeWindowAggregations {
		if t > activeTimeWindow {
			continue
		}
		for _, value := range windowAggregations {
			var metricEnvelope = models.MetricEnvelope{
				value,
				map[string]string{},
				int64(time.Now().Unix() * 1000)}

			value, _ := json.Marshal(metricEnvelope)

			outbound <- &kafka.Message{TopicPartition: kafka.TopicPartition{Topic: topic, Partition: kafka.PartitionAny}, Value: []byte(value)}
		}
	}

	// TODO: Confirm messages published before committing offsets
	offsetList, activeTimeWindow := getMinOffsets(activeTimeWindow, topic)
	commitOffsets(offsetList, c)
	deleteInactiveTimeWindows(activeTimeWindow)
}
// Get the min offsets in Kafka for each time window.
func getMinOffsets(activeTimeWindow int64, topic *string) (map[int32]kafka.TopicPartition, int64) {
	var offsetList = map[int32]kafka.TopicPartition{}
	for eventWindow, partitions := range offsetCache {
		if eventWindow > activeTimeWindow {
			continue
		}
		for partition, offset := range partitions {
			if int64(offsetList[partition].Offset) <= 0 || offset <= int64(offsetList[partition].Offset) {
				new_offset, err := kafka.NewOffset(offset + 1)
				if err != nil {
					log.Fatalf("Failed to update kafka offset %s[%d]@%d", topic, partition, offset)
				}
				offsetList[partition] = kafka.TopicPartition{
					Topic:     topic,
					Partition: partition,
					Offset:    new_offset}
			}
		}
	}
	return offsetList, activeTimeWindow
}
// Commit the Kafka offsets.
func commitOffsets(offsetList map[int32]kafka.TopicPartition, c *kafka.Consumer) {
	if len(offsetList) > 0 {
		finalOffsets := make([]kafka.TopicPartition, len(offsetList))
		idx := 0
		for _, value := range offsetList {
			finalOffsets[idx] = value
			idx++
		}
		log.Info(finalOffsets)
		_, err := c.CommitOffsets(finalOffsets)
		if err != nil {
			log.Errorf("Consumer errors submitting offsets %v", err)
		}
	}
}
// Delete time window aggregations for inactive time windows.
func deleteInactiveTimeWindows(activeTimeWindow int64) {
	for timeWindow := range timeWindowAggregations {
		if timeWindow <= activeTimeWindow {
			log.Debugf("Delete time window %d", timeWindow)
			delete(timeWindowAggregations, timeWindow)
			delete(offsetCache, timeWindow)
		}
	}
}

// TODO: Add Prometheus Client library and report metrics
// TODO: Create Helm Charts
// TODO: Add support for consuming/publishing intermediary aggregations. For example, publish a (sum, count) to use in an avg aggregation
// TODO: Guarantee at least once publishing of aggregated metrics
// TODO: Allow start/end consumer offsets to be specified as parameters.
// TODO: Allow start/end aggregation period to be specified.
func main() {
	initLogging()
	initConfig()

	windowSize = time.Duration(viper.GetInt("WindowSize") * 1e9)
	windowLag = time.Duration(viper.GetInt("WindowLag") * 1e9)
	consumerTopic := viper.GetString("consumerTopic")
	producerTopic := viper.GetString("producerTopic")
	err := viper.UnmarshalKey("aggregationSpecifications", &aggregationSpecifications)

	if err != nil {
		log.Fatalf("unable to decode into struct, %v", err)
	}

	bootstrapServers := viper.GetString("kafka.bootstrap.servers")
	groupId := viper.GetString("kafka.group.id")


	sigchan := make(chan os.Signal)
	signal.Notify(sigchan, syscall.SIGINT, syscall.SIGTERM)

	c := initConsumer(consumerTopic, groupId, bootstrapServers)
	defer c.Close()

	p := initProducer(bootstrapServers)
	defer p.Close()

	go handleProducerEvents(p)

	// align to time boundaries?
	firstTick := firstTick()
	var ticker *time.Ticker = new(time.Ticker)

	run := true
	processed_msg_count := 0

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
				var eventTimeWindow = int64(metric.Timestamp) / (1000 * int64(windowSize.Seconds()))

				for _, aggregationSpecification := range aggregationSpecifications {
					if models.MatchMetric(aggregationSpecification, metric) {
						var windowAggregations = timeWindowAggregations[eventTimeWindow]

						if windowAggregations == nil {
							timeWindowAggregations[eventTimeWindow] = make(map[string]models.Metric)
							windowAggregations = timeWindowAggregations[eventTimeWindow]
						}

						var aggregationKey = aggregationSpecification.AggregatedMetricName

						// make the key unique for the supplied groupings
						if aggregationSpecification.GroupedDimensions != nil {
							for _, key := range aggregationSpecification.GroupedDimensions {
								aggregationKey += "," + key + ":" + metric.Dimensions[key]
							}
						}
						log.Debugf("Storing key %s", aggregationKey)

						currentMetric := windowAggregations[aggregationKey]

						// create a new metric is one did not exist
						if currentMetric.Name == "" {
							currentMetric.Name = aggregationSpecification.AggregatedMetricName
							currentMetric.Dimensions = aggregationSpecification.FilteredDimensions
							if currentMetric.Dimensions == nil {
								currentMetric.Dimensions = map[string]string{}
							}
							for _, key := range aggregationSpecification.GroupedDimensions {
								currentMetric.Dimensions[key] = metric.Dimensions[key]
							}
							currentMetric.Value = metric.Value
							currentMetric.Timestamp = float64(eventTimeWindow * 1000 * int64(windowSize.Seconds()))
							windowAggregations[aggregationKey] = currentMetric
						} else {
							currentMetric.Value += metric.Value
							windowAggregations[aggregationKey] = currentMetric
						}

						offsetTimeWindow := offsetCache[eventTimeWindow]
						partition := e.TopicPartition.Partition
						offset := int64(e.TopicPartition.Offset)
						if offsetTimeWindow == nil {
							log.Infof("Initialized offset window %d", eventTimeWindow)
							offsetCache[eventTimeWindow] = make(map[int32]int64)
							offsetCache[eventTimeWindow][partition] = offset
						} else if offsetTimeWindow[partition] <= 0 || offset <= offsetTimeWindow[partition]{
							log.Infof("Updated offset for partition %d to %d", partition, offset)
							offsetTimeWindow[partition] = offset
						}
					}
				}
				log.Debug(metricEnvelope)
				processed_msg_count += 1
			case kafka.OffsetsCommitted:
				log.Infof("Commited offsets ", e)
			case kafka.PartitionEOF:
				//log.Infof("%% Reached %v", e)
			case kafka.Error:
				log.Errorf("%% Error: %v", e)
				run = false
			}

		case <-firstTick.C:
			ticker = time.NewTicker(windowSize)
			log.Infof("Processed %d messages", processed_msg_count)
			processed_msg_count = 0
			publishAggregations(p.ProduceChannel(), &producerTopic, c)

		case <-ticker.C:
			log.Infof("Processed %d messages", processed_msg_count)
			processed_msg_count = 0
			publishAggregations(p.ProduceChannel(), &producerTopic, c)
		}
	}

	log.Info("Stopped monasca-aggregation")
}
