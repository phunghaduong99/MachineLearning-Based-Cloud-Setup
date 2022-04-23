package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"

	"github.com/Shopify/sarama"
)

//PolicymsConf... The configmap of the policyms
type PolicymsConf struct {
	Cds        string `json:"cds"`
	Kafka      string `json:"kafka"`
	KafkaTopic string `json:"kafkaTopic"`
	CNFIp      string `json:"cnfIp"`
	CNFPort    string `json:"cnfPort"`
}

type cdsPayload struct {
	ActionIdentifiers struct {
		Mode             string `json:"mode"`
		BlueprintName    string `json:"blueprintName"`
		BlueprintVersion string `json:"blueprintVersion"`
		ActionName       string `json:"actionName"`
	} `json:"actionIdentifiers"`
	Payload struct {
		StreamCountConfigEditRequest struct {
			StreamCountConfigEditProperties struct {
				PnfID             string `json:"pnf-id"`
				PnfIpv4Address    string `json:"pnf-ipv4-address"`
				NetconfPassword   string `json:"netconf-password"`
				NetconfUsername   string `json:"netconf-username"`
				NetconfServerPort string `json:"netconf-server-port"`
				VfwEditPayload    struct {
					ActiveStreams string `json:"active-streams"`
				} `json:"vfwEditPayload"`
			} `json:"stream-count-config-edit-properties"`
		} `json:"stream-count-config-edit-request"`
	} `json:"payload"`
	CommonHeader struct {
		SubRequestID string `json:"subRequestId"`
		RequestID    string `json:"requestId"`
		OriginatorID string `json:"originatorId"`
	} `json:"commonHeader"`
}

// Sarama configuration options
var (
	brokers  = "10.99.114.79:9092"
	version  = ""
	group    = "CG1"
	topics   = "unauthenticated.DCAE_CL_OUTPUT"
	assignor = ""
	oldest   = true
	verbose  = false
	conf     = PolicymsConf{}
)

//go run main.go -brokers="http://10.99.114.79" -topics="unauthenticated.SEC_MEASUREMENT_OUTPUT" -group="CG1"

func init() {
	configFile, err := os.Open("/opt/emco/config/policyms.conf")
	if err != nil {
		log.Printf("Failed to read config service configuration")
		return
	}
	defer configFile.Close()

	// Read the configuration json
	byteValue, _ := ioutil.ReadAll(configFile)
	json.Unmarshal(byteValue, &conf)

	log.Printf("Config %s\n", conf)
	flag.StringVar(&brokers, "brokers", conf.Kafka, "Kafka bootstrap brokers to connect to, as a comma separated list")
	flag.StringVar(&group, "group", "CG1", "Kafka consumer group definition")
	flag.StringVar(&version, "version", "2.1.1", "Kafka cluster version")
	flag.StringVar(&topics, "topics", conf.KafkaTopic, "Kafka topics to be consumed, as a comma separated list")
	flag.StringVar(&assignor, "assignor", "range", "Consumer group partition assignment strategy (range, roundrobin, sticky)")
	flag.BoolVar(&oldest, "oldest", true, "Kafka consumer consume initial offset from oldest")
	flag.BoolVar(&verbose, "verbose", false, "Sarama logging")
	flag.Parse()
}

func apiPost(jsonLoad []byte, url string) (interface{}, []byte, error) {
	// prepare and POST API
	request, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonLoad))
	if err != nil {
		return nil, nil, err
	}
	// --header 'Content-Type: application/json'
	// --header 'Accept: application/json'
	// --header 'Authorization: Basic Y2NzZGthcHBzOmNjc2RrYXBwcw=='
	request.Header.Add("Content-Type", "application/json")
	request.Header.Add("Accept", "application/json")
	request.Header.Add("Authorization", "Basic Y2NzZGthcHBzOmNjc2RrYXBwcw==")
	client := http.Client{}
	resp, err := client.Do(request)
	if err != nil {
		return nil, nil, err
	}
	defer resp.Body.Close()

	// Prepare the response
	data, _ := ioutil.ReadAll(resp.Body)

	return resp.StatusCode, data, nil
}

func main() {

	log.Println("Starting a new Sarama consumer")

	if verbose {
		sarama.Logger = log.New(os.Stdout, "[sarama] ", log.LstdFlags)
	}

	version, err := sarama.ParseKafkaVersion(version)
	if err != nil {
		log.Panicf("Error parsing Kafka version: %v", err)
	}

	/**
	 * Construct a new Sarama configuration.
	 * The Kafka cluster version has to be defined before the consumer/producer is initialized.
	 */
	config := sarama.NewConfig()
	config.Version = version

	switch assignor {
	case "sticky":
		config.Consumer.Group.Rebalance.Strategy = sarama.BalanceStrategySticky
	case "roundrobin":
		config.Consumer.Group.Rebalance.Strategy = sarama.BalanceStrategyRoundRobin
	case "range":
		config.Consumer.Group.Rebalance.Strategy = sarama.BalanceStrategyRange
	default:
		log.Panicf("Unrecognized consumer group partition assignor: %s", assignor)
	}

	if oldest {
		config.Consumer.Offsets.Initial = sarama.OffsetOldest
	}

	/**
	 * Setup a new Sarama consumer group
	 */
	consumer := Consumer{
		ready: make(chan bool),
	}

	ctx, cancel := context.WithCancel(context.Background())
	log.Printf("broker %s", brokers)
	client, err := sarama.NewConsumerGroup(strings.Split(brokers, ","), group, config)
	if err != nil {
		log.Panicf("Error creating consumer group client: %v", err)
	}

	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			// `Consume` should be called inside an infinite loop, when a
			// server-side rebalance happens, the consumer session will need to be
			// recreated to get the new claims
			if err := client.Consume(ctx, strings.Split(topics, ","), &consumer); err != nil {
				log.Panicf("Error from consumer: %v", err)
			}
			// check if context was cancelled, signaling that the consumer should stop
			if ctx.Err() != nil {
				return
			}
			consumer.ready = make(chan bool)
		}
	}()

	<-consumer.ready // Await till the consumer has been set up
	log.Println("Sarama consumer up and running!...")

	sigterm := make(chan os.Signal, 1)
	signal.Notify(sigterm, syscall.SIGINT, syscall.SIGTERM)
	select {
	case <-ctx.Done():
		log.Println("terminating: context cancelled")
	case <-sigterm:
		log.Println("terminating: via signal")
	}
	cancel()
	wg.Wait()
	if err = client.Close(); err != nil {
		log.Panicf("Error closing client: %v", err)
	}
}

// Consumer represents a Sarama consumer group consumer
type Consumer struct {
	ready chan bool
}

// Setup is run at the beginning of a new session, before ConsumeClaim
func (consumer *Consumer) Setup(sarama.ConsumerGroupSession) error {
	// Mark the consumer as ready
	close(consumer.ready)
	return nil
}

// Cleanup is run at the end of a session, once all ConsumeClaim goroutines have exited
func (consumer *Consumer) Cleanup(sarama.ConsumerGroupSession) error {
	return nil
}

// ConsumeClaim must start a consumer loop of ConsumerGroupClaim's Messages().
func (consumer *Consumer) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	// NOTE:
	// Do not move the code below to a goroutine.
	// The `ConsumeClaim` itself is called within a goroutine, see:
	// https://github.com/Shopify/sarama/blob/main/consumer_group.go#L27-L29
	for message := range claim.Messages() {
		log.Printf("Message claimed: value = %s, timestamp = %v, topic = %s", string(message.Value), message.Timestamp, message.Topic)
		session.MarkMessage(message, "")
		schemaPayload := cdsPayload{}

		schemaPayload.ActionIdentifiers.Mode = "sync"
		schemaPayload.ActionIdentifiers.BlueprintVersion = "1.0.0"
		schemaPayload.ActionIdentifiers.BlueprintName = "vfw_netconf"
		schemaPayload.ActionIdentifiers.ActionName = "stream-count-config-edit"

		schemaPayload.Payload.StreamCountConfigEditRequest.StreamCountConfigEditProperties.NetconfPassword = "admin"
		schemaPayload.Payload.StreamCountConfigEditRequest.StreamCountConfigEditProperties.NetconfUsername = "admin"
		schemaPayload.Payload.StreamCountConfigEditRequest.StreamCountConfigEditProperties.NetconfServerPort = conf.CNFPort
		schemaPayload.Payload.StreamCountConfigEditRequest.StreamCountConfigEditProperties.PnfIpv4Address = conf.CNFIp
		schemaPayload.Payload.StreamCountConfigEditRequest.StreamCountConfigEditProperties.PnfID = "Packet Generator"
		schemaPayload.Payload.StreamCountConfigEditRequest.StreamCountConfigEditProperties.VfwEditPayload.ActiveStreams = "5"

		schemaPayload.CommonHeader.SubRequestID = "143748f9-3cd5-4910-81c9-a4601ff2ea58"
		schemaPayload.CommonHeader.OriginatorID = "e5eb1f1e-3386-435d-b290-d49d8af8db4c"
		schemaPayload.CommonHeader.RequestID = "SDNC_DG"

		//url := "http://" + c.CdsEndpoint + "/api/v1/execution-service/process"
		url := "http://" + conf.Cds + "/api/v1/execution-service/process"
		jsonLoad, _ := json.Marshal(schemaPayload)
		retCode, data, err := apiPost(jsonLoad, url)
		log.Printf("data %s, retCode %d ", data, retCode)
		if err != nil {

			log.Printf("data %s, retCode %d", data, retCode)
		}
	}

	return nil
}

/*
{
	"actionIdentifiers": {
			"mode": "sync",
			"blueprintName": "vfw_netconf",
			"blueprintVersion": "1.0.0",
			"actionName": "stream-count-config-edit"
	},
	"payload": {
			"stream-count-config-edit-request": {
					"stream-count-config-edit-properties": {
							"pnf-id": "Packet Generator",
							"pnf-ipv4-address": "192.168.102.81",
							"netconf-password": "admin",
							"netconf-username": "admin",
							"netconf-server-port": "30831",
							"vfwEditPayload": {
									"active-streams": "5"
							}
					}
			}

	},
	"commonHeader": {
			"subRequestId": "143748f9-3cd5-4910-81c9-a4601ff2ea58",
			"requestId": "e5eb1f1e-3386-435d-b290-d49d8af8db4c",
			"originatorId": "SDNC_DG"
	}
}*/
