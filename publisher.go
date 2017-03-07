package main

import (
	"fmt"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"os"
	"github.hpe.com/UNCLE/monasca-aggregation/models"
	"time"
	"encoding/json"
	"strconv"
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

	if err != nil {
		fmt.Printf("Failed to create producer: %s\n", err)
		os.Exit(1)
	}

	fmt.Printf("Created Producer %v\n", p)

	doneChan := make(chan bool)

	go func() {
	outer:
		for e := range p.Events() {
			switch ev := e.(type) {
			case *kafka.Message:
				m := ev
				if m.TopicPartition.Error != nil {
					fmt.Printf("Delivery failed: %v\n", m.TopicPartition.Error)
				} else {
					fmt.Printf("Delivered message to topic %s [%d] at offset %v\n",
						*m.TopicPartition.Topic, m.TopicPartition.Partition, m.TopicPartition.Offset)
				}
				break outer

			default:
				fmt.Printf("Ignored event: %s\n", ev)
			}
		}

		close(doneChan)
	}()

	for {
		for i := 0; i < 3; i++ {
			for j := 0; j < 2; j++ {

				dimensions := map[string]string{"service": strconv.Itoa(i), "hostname": strconv.Itoa(j)}

				var metricEnvelope = models.MetricEnvelope{
					models.Metric{"metric2", dimensions, float64(time.Now().Unix()) * 1000, 1.0, map[string]string{}},
					map[string]string{},
					int64(time.Now().Unix() * 1000)}

				value, _ := json.Marshal(metricEnvelope)

				p.ProduceChannel() <- &kafka.Message{TopicPartition: kafka.TopicPartition{Topic: &topic, Partition: kafka.PartitionAny}, Value: []byte(value)}
			}
		}

		time.Sleep(time.Second)

		// wait for delivery report goroutine to finish
		_ = <-doneChan

		//fmt.Printf("Done waiting\n")

	}

	p.Close()
}
