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
	"fmt"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/monasca/monasca-aggregator/models"
	"os"
	"strconv"
	"time"
)

func main() {

	if len(os.Args) != 3 {
		fmt.Fprintf(os.Stderr, "Usage: %s <broker> <topic>\n",
			os.Args[0])
		os.Exit(1)
	}

	broker := os.Args[1]
	topic := os.Args[2]

	p, err := kafka.NewProducer(&kafka.ConfigMap{"bootstrap.servers": broker})
	defer p.Close()

	if err != nil {
		fmt.Printf("Failed to create producer: %s\n", err)
		os.Exit(1)
	}

	fmt.Printf("Created Producer %v\n", p)

	//go func() {
	//	for e := range p.Events() {
	//		switch ev := e.(type) {
	//		case *kafka.Message:
	//			m := ev
	//			if m.TopicPartition.Error != nil {
	//				fmt.Printf("Delivery failed: %v\n", m.TopicPartition.Error)
	//			} else {
	//				fmt.Printf("Delivered message to topic %s [%d] at offset %v\n",
	//					*m.TopicPartition.Topic, m.TopicPartition.Partition, m.TopicPartition.Offset)
	//			}
	//
	//		default:
	//			fmt.Printf("Ignored event: %s\n", ev)
	//		}
	//	}
	//}()

	for {
		for i := 0; i < 3; i++ {
			for j := 0; j < 2; j++ {

				dimensions := map[string]string{"service": strconv.Itoa(i), "hostname": strconv.Itoa(j)}

				var metricEnvelope = models.MetricEnvelope{
					Metric: models.Metric{
						Name:       "metric2",
						Dimensions: dimensions,
						Timestamp:  float64(time.Now().Unix()) * 1000,
						Value:      2.0,
						ValueMeta:  map[string]string{}},
					Meta:         map[string]string{},
					CreationTime: int64(time.Now().Unix() * 1000)}

				value, _ := json.Marshal(metricEnvelope)

				p.ProduceChannel() <- &kafka.Message{TopicPartition: kafka.TopicPartition{Topic: &topic, Partition: kafka.PartitionAny}, Value: value}
			}
		}

		time.Sleep(time.Second)

	}
}
