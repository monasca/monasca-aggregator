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
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/spf13/viper"

	"github.com/monasca/monasca-aggregator/aggregation"
	"github.com/monasca/monasca-aggregator/models"
)

var windowSize time.Duration
var windowLag time.Duration

var offsetCache = make(map[int64]map[int32]int64)

//init global references to counters
var inCounter = prometheus.NewCounter(
	prometheus.CounterOpts{
		Name: "in_messages",
		Help: "Number of messages read"})
var outCounter = prometheus.NewCounter(
	prometheus.CounterOpts{
		Name: "out_messages",
		Help: "Number of messages written"})

// This is a workaround to handle https://issues.apache.org/jira/browse/KAFKA-3593
var lastMessage time.Time = time.Now()

var config = initConfig()
var aggregations = initAggregationSpecs()

var aggregationRules = initAggregationRules(aggregations)

func init() {
	// Log as JSON instead of the default ASCII formatter.
	log.SetFormatter(&log.JSONFormatter{})
	//log.SetOutput(os.Stdout)
	//log.SetLevel(log.InfoLevel)
	logfile := config.GetString("logging.file")
	if logfile != "" {
		f, err := os.Create(logfile)
		if err != nil {
			log.SetOutput(os.Stdout)
			log.Fatalf("Failed to create log file: %v", err)
		}
		log.SetOutput(f)
	} else {
		log.SetOutput(os.Stdout)
	}

	loglevel := config.GetString("logging.level")
	switch strings.ToUpper(loglevel) {
	case "DEBUG":
		log.SetLevel(log.DebugLevel)
	case "INFO":
		log.SetLevel(log.InfoLevel)
	case "WARN":
		log.SetLevel(log.WarnLevel)
	case "ERROR":
		log.SetLevel(log.ErrorLevel)
	default:
		log.SetLevel(log.WarnLevel)
	}

	prometheus.MustRegister(inCounter)
	prometheus.MustRegister(outCounter)
}

func initConfig() *viper.Viper {
	config := viper.New()
	config.SetDefault("windowSize", 10)
	config.SetDefault("windowLag", 2)
	config.SetDefault("consumerTopic", "metrics")
	config.SetDefault("producerTopic", "metrics")
	config.SetDefault("kafka.bootstrap.servers", "localhost:9092")
	config.SetDefault("kafka.group.id", "monasca-aggregation")
	config.SetDefault("prometheus.endpoint", "localhost:8080")
	config.SetConfigName("config")
	config.AddConfigPath(".")
	err := config.ReadInConfig()

	if err != nil {
		log.Fatalf("Fatal error reading config file: %s", err)
	}
	return config
}

func initAggregationSpecs() []models.AggregationSpecification {
	config := viper.New()
	config.SetConfigName("aggregation-specifications")
	config.AddConfigPath(".")
	err := config.ReadInConfig()
	if err != nil {
		log.Fatalf("Fatal error reading aggregations: %s", err)
	}
	var aggregations []models.AggregationSpecification
	err = config.UnmarshalKey("aggregationSpecifications", &aggregations)
	if err != nil {
		log.Fatalf("Failed to parse aggregations: %s", err)
	}

	return aggregations
}

func initAggregationRules(specifications []models.AggregationSpecification) []aggregation.Rule {
	var rules = make([]aggregation.Rule, len(specifications))
	i := 0
	for _, spec := range specifications {
		rules[i] = aggregation.NewAggregationRule(spec)
		i++
		log.Infof("New Spec: %v", spec)
	}
	return rules
}

func initConsumer(consumerTopic, groupID, bootstrapServers string) *kafka.Consumer {
	c, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers":               bootstrapServers,
		"group.id":                        groupID,
		"session.timeout.ms":              6000,
		"go.events.channel.enable":        true,
		"go.application.rebalance.enable": true,
		"enable.auto.commit":              false,
		"default.topic.config":            kafka.ConfigMap{"auto.offset.reset": "earliest"},
	})

	if err != nil {
		log.Fatalf("Failed to create consumer: %s", err)
	}

	log.Infof("Created kafka consumer %v", c)

	err = c.Subscribe(consumerTopic, nil)

	if err != nil {
		log.Fatalf("Failed to subscribe to topics %c", err)
	}
	log.Infof("Subscribed to topic %s as group %s", consumerTopic, groupID)

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
	for e := range p.Events() {
		switch ev := e.(type) {
		case *kafka.Message:
			m := ev
			if m.TopicPartition.Error != nil {
				log.Errorf("Delivery failed: %v\n", m.TopicPartition.Error)
			} else {
				log.Debugf("Delivered message to topic %s [%d] at offset %v\n",
					*m.TopicPartition.Topic, m.TopicPartition.Partition, m.TopicPartition.Offset)
			}

		default:
			log.Debugf("Ignored event: %s\n", ev)
		}
	}
}

// Return a timer for when the first window should be processed
// TODO: Check this math to account for all boundary conditions and large lag times
func firstTick() *time.Timer {
	now := time.Now().Unix()
	completed := now%int64(windowSize.Seconds()) - int64(windowLag.Seconds())
	remaining := int64(windowSize.Seconds()) - completed
	firstTick := time.NewTimer(time.Duration(remaining * 1e9))
	return firstTick
}

func publishAggregations(outbound chan *kafka.Message, topic *string, c *kafka.Consumer) {
	var currentTimeWindow = int64(time.Now().Unix()) / int64(windowSize.Seconds())
	// roof(windowLag / windowSize) i.e. number of windows in the lag time
	var windowLagCount = int64(windowLag.Seconds()/windowSize.Seconds()) + 1
	var activeTimeWindow = currentTimeWindow - windowLagCount
	log.Debugf("currentTimeWindow: %d", currentTimeWindow)
	log.Debugf("Publishing metrics in window %d", activeTimeWindow)

	for _, rule := range aggregationRules {
	windowLoop:
		for windowTime := range rule.Windows {
			if windowTime > activeTimeWindow {
				continue windowLoop
			}

			for _, metric := range rule.GetMetrics(windowTime) {
				metric.CreationTime = time.Now().Unix() * 1000
				value, _ := json.Marshal(metric)

				outbound <- &kafka.Message{TopicPartition: kafka.TopicPartition{Topic: topic, Partition: kafka.PartitionAny}, Value: []byte(value)}
				outCounter.Inc()
			}
		}
	}

	// TODO: Confirm messages published before committing offsets
	offsetList := maxOffsets(activeTimeWindow)
	commitOffsets(offsetList, topic, c)

	deleteInactiveTimeWindows(activeTimeWindow)
}

func maxOffsets(eventTime int64) map[int32]int64 {
	output := make(map[int32]int64)

	for windowTime, window := range offsetCache {
		if windowTime > eventTime {
			continue
		}

		for partition, offset := range window {
			oldOffset, exists := output[partition]
			if !exists || offset > oldOffset {
				output[partition] = offset
			}
		}
	}
	return output
}

// Commit the Kafka offsets.
func commitOffsets(offsetList map[int32]int64, topic *string, c *kafka.Consumer) {
	if len(offsetList) <= 0 {
		log.Debug("No new offsets, skipping commit")
		return
	}
	//convert ints to kafka structures
	finalOffsets := make([]kafka.TopicPartition, len(offsetList))
	idx := 0
	for partition, value := range offsetList {
		newOffset, err := kafka.NewOffset(value + 1)
		if err != nil {
			log.Fatalf("Failed to update kafka offset %s[%d]@%d", *topic, partition, value)
		}
		finalOffsets[idx] = kafka.TopicPartition{
			Topic:     topic,
			Partition: partition,
			Offset:    newOffset}
		idx++
	}

	log.Infof("Committing offsets: %v", finalOffsets)
	_, err := c.CommitOffsets(finalOffsets)
	if err != nil {
		log.Errorf("Consumer errors submitting offsets %v", err)
	}
}

// Delete time window aggregations for inactive time windows.
func deleteInactiveTimeWindows(activeTimeWindow int64) {
	log.Debugf("Deleteing windows older than %d", activeTimeWindow)
	for _, rule := range aggregationRules {
		for windowTime := range rule.Windows {
			if windowTime <= activeTimeWindow {
				delete(rule.Windows, windowTime)
			}
		}
	}
	for windowTime := range offsetCache {
		if windowTime <= activeTimeWindow {
			delete(offsetCache, windowTime)
		}
	}
}

func processMessage(e *kafka.Message) {
	metricEnvelope := models.MetricEnvelope{}
	err := json.Unmarshal([]byte(e.Value), &metricEnvelope)
	if err != nil {
		log.Warnf("%% Invalid metric envelope on %s:%s", e.TopicPartition, string(e.Value))
		return
	}
	log.Debugf("Processing metric: %v", metricEnvelope)

	for _, rule := range aggregationRules {
		if rule.MatchesMetric(metricEnvelope) {
			rule.AddMetric(metricEnvelope, windowSize)
		}
	}

	eventTime := int64(metricEnvelope.Metric.Timestamp) / (1000 * int64(windowSize.Seconds()))
	_, exists := offsetCache[eventTime]
	if !exists {
		offsetCache[eventTime] = make(map[int32]int64)
	}
	//we can assume each new message read is the latest offset for its partition
	offsetCache[eventTime][e.TopicPartition.Partition] = int64(e.TopicPartition.Offset)

	inCounter.Inc()
	lastMessage = time.Now()
}

// TODO: Add validation for aggregation rules (and metrics?)
// TODO Allow easy grouping on 'all' dimensions
// TODO: Allow start/end consumer offsets to be specified as parameters.
// TODO: Allow start/end aggregation period to be specified.
func main() {
	windowSize = time.Duration(config.GetInt("WindowSize") * 1e9)
	windowLag = time.Duration(config.GetInt("WindowLag") * 1e9)
	consumerTopic := config.GetString("consumerTopic")
	producerTopic := config.GetString("producerTopic")

	bootstrapServers := config.GetString("kafka.bootstrap.servers")
	groupID := config.GetString("kafka.group.id")

	prometheusEndpoint := config.GetString("prometheus.endpoint")

	sigchan := make(chan os.Signal)
	signal.Notify(sigchan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		sig := <-sigchan
		log.Fatalf("Caught signal %v: terminating", sig)
	}()

	c := initConsumer(consumerTopic, groupID, bootstrapServers)
	defer c.Close()

	p := initProducer(bootstrapServers)
	defer p.Close()

	go handleProducerEvents(p)

	log.Info("Started monasca-aggregation")

	// align to time boundaries?
	firstTick := firstTick()
	var ticker = new(time.Ticker)

	go func() {
		// Start prometheus endpoint
		http.Handle("/metrics", promhttp.Handler())
		log.Fatal(http.ListenAndServe(prometheusEndpoint, nil))
	}()
	log.Infof("Serving metrics on %s/metrics", prometheusEndpoint)

	for true {
		select {
		case <-firstTick.C:
			ticker = time.NewTicker(windowSize)
			publishAggregations(p.ProduceChannel(), &producerTopic, c)
		case <-ticker.C:
			publishAggregations(p.ProduceChannel(), &producerTopic, c)
			// if the window size has passed and no messages have been received, die in case
			// https://issues.apache.org/jira/browse/KAFKA-3593 is occurring.
			if time.Now().Sub(lastMessage) > windowSize {
				log.Fatalf("No messages received since %v", lastMessage)
			}

		case ev := <-c.Events():
			switch e := ev.(type) {
			case kafka.AssignedPartitions:
				log.Infof("%% %v", e)
				c.Assign(e.Partitions)

			case kafka.RevokedPartitions:
				log.Infof("%% %v", e)
				c.Unassign()

			case *kafka.Message:
				processMessage(e)

			case kafka.OffsetsCommitted:
				log.Infof("Commited offsets ", e)

			case kafka.PartitionEOF:
				//log.Infof("%% Reached %v", e)

			case kafka.Error:
				log.Fatalf("%% Error: %v", e)
			}
		}
	}

	log.Info("Stopped monasca-aggregation")
}
