// ------------------------------------------------------ {COPYRIGHT-TOP} ---
// Licensed Materials - Property of IBM
// 5900-AEO
//
// Copyright IBM Corp. 2020, 2021. All Rights Reserved.
//
// US Government Users Restricted Rights - Use, duplication, or
// disclosure restricted by GSA ADP Schedule Contract with IBM Corp.
// ------------------------------------------------------ {COPYRIGHT-END} ---
package producer

import (
	"context"
	"crypto/sha512"
	"crypto/tls"
	"crypto/x509"
	"hash"
	"log"
	"math"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/Shopify/sarama"
	"github.com/cloudevents/sdk-go/protocol/kafka_sarama/v2"
	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/gocarina/gocsv"
	"github.com/google/uuid"
	"github.com/xdg/scram"

	"github.ibm.com/automation-base-pak/abp-demo-cartridge/pkg/config"
)

const (
	baiTimeFormat = "1/2/06 15:04" // Or if Excel didn't mess it up: "2006-01-02 15:04:05"
)

// Start kafka producer
func Start(cfg *config.Config) {
	log.Println("Starting Producer Service")
	brokers := strings.Split(cfg.BootstrapServers, ",")
	var messagesPerGroup, secondsToPause, sequenceRepititions int
	var err error = nil
	if messagesPerGroup, err = strconv.Atoi(cfg.MessagesPerGroup); err != nil {
		messagesPerGroup = math.MaxInt32
	}
	if secondsToPause, err = strconv.Atoi(cfg.SecondsToPause); err != nil {
		secondsToPause = 0
	}
	if sequenceRepititions, err = strconv.Atoi(cfg.SequenceRepititions); err != nil {
		sequenceRepititions = 1
	}
	config := newConfig(cfg)

	logFile, err := os.OpenFile("/var/log/demoproducer.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.Fatalf("Failed to open log file: %s", err.Error())
	}
	sarama.Logger = log.New(logFile, "[sarama] ", log.LstdFlags)

	sender, err := kafka_sarama.NewSender(brokers, config, cfg.KafkaTopic)
	defer sender.Close(context.Background())
	if err != nil {
		log.Fatalf("failed to create protocol: %s", err.Error())
	}

	c, err := cloudevents.NewClient(sender, cloudevents.WithTimeNow(), cloudevents.WithUUIDs())
	if err != nil {
		log.Fatalf("failed to create client, %v", err)
	}

	// Send sequences of events to Kafka
	for i := 0; i != sequenceRepititions; i++ {
		sendSequence(c, messagesPerGroup, secondsToPause)
		log.Println("Sequence complete.")
	}

	// Wait forever
	wg := sync.WaitGroup{}
	wg.Add(1)
	wg.Wait()
}

func sendSequence(c cloudevents.Client, messagesPerGroup int, secondsToPause int) {
	data, err := os.Open("sample.csv")
	if err != nil {
		log.Fatal(err)
	}
	defer data.Close()
	log.Println("messagesPerGroup is ", messagesPerGroup)
	log.Println("secondsToPause is ", secondsToPause)
	samples := []*BaiMessage{}
	err = gocsv.UnmarshalFile(data, &samples)
	if err != nil {
		log.Fatal(err)
	}

	numSent := 0
	for _, s := range samples {
		// Get the time of the event for inclusion in the CloudEvent
		t, err := time.Parse(baiTimeFormat, s.DateTime)
		if err != nil {
			log.Fatal(err)
		}

		// Ensure a consistent format inside the actual message:
		s.DateTime = t.Format(time.RFC3339)

		sendMessage(c, "bai.events.sample", t, s)
		numSent++
		if numSent >= messagesPerGroup {
			log.Println("After sending", numSent, "messages, pausing for", secondsToPause, "seconds")
			time.Sleep(time.Duration(secondsToPause) * time.Second)
			numSent = 0
		}
	}
}

type BaiMessage struct {
	Req_Line_ID             string `csv:"Req_Line_ID"`
	Order_Line_ID           string `csv:"Order_Line_ID"`
	Goods_ID                string `csv:"Goods_ID"`
	Invoice_ID              string `csv:"Invoice_ID"`
	Activity                string `csv:"Activity"`
	DateTime                string `csv:"DateTime"`
	Resource                string `csv:"Resource"`
	Role                    string `csv:"Role"`
	Requisition_Vendor      string `csv:"Requisition_Vendor"`
	Order_Vendor            string `csv:"Order_Vendor"`
	Invoice_Vendor          string `csv:"Invoice_Vendor"`
	Pay_Vendor              string `csv:"Pay_Vendor"`
	Requisition_Type        string `csv:"Requisition_Type"`
	Order_Type              string `csv:"Order_Type"`
	Purchasing_Group        string `csv:"Purchasing_Group"`
	Purchasing_Organization string `csv:"Purchasing_Organization"`
	Material_Group          string `csv:"Material_Group"`
	Material_Number         string `csv:"Material_Number"`
	Plant                   string `csv:"Plant"`
	Good_ReferenceNumber    string `csv:"Good_ReferenceNumber"`
	Requisition_Header      string `csv:"Requisition_Header"`
	Order_Header            string `csv:"Order_Header"`
	Invoice_Header          string `csv:"Invoice_Header"`
	ClearDoc_Header         string `csv:"ClearDoc_Header"`
	Good_Year               string `csv:"Good_Year"`
	Invoice_Year            string `csv:"Invoice_Year"`
	Order_Line_Amount       string `csv:"Order_Line_Amount"`
	Invoice_Amount          string `csv:"Invoice_Amount"`
	Paid_Amount             string `csv:"Paid_Amount"`
	Invoice_Document_Date   string `csv:"Invoice_Document_Date"`
	Invoice_Due_Date        string `csv:"Invoice_Due_Date"`
	Pay_Type                string `csv:"Pay_Type"`
	Pay_Delay               string `csv:"Pay_Delay"`
	UserType                string `csv:"UserType"`
	Invoice_Is_Overdue      string `csv:"Invoice_Is_Overdue"`
}

func sendMessage(c cloudevents.Client, evType string, t time.Time, data interface{}) {
	e := cloudevents.NewEvent()
	e.SetID(uuid.New().String())
	e.SetType(evType)
	e.SetSource("https://github.ibm.com/automation-base-pak/abp-demo-cartridge")
	e.SetTime(t)
	_ = e.SetData(cloudevents.ApplicationJSON, data)

	if result := c.Send(
		// Set the producer message key
		kafka_sarama.WithMessageKey(context.Background(), sarama.StringEncoder(e.ID())),
		e,
	); cloudevents.IsUndelivered(result) {
		log.Printf("failed to send %s: %v", e.ID(), result)
	} else {
		log.Printf("sent: %s, accepted: %t", e.ID(), cloudevents.IsACK(result))
	}
}

func newConfig(cfg *config.Config) *sarama.Config {
	config := sarama.NewConfig()
	config.Version = sarama.V2_3_0_0
	config.Producer.RequiredAcks = sarama.WaitForAll // Wait for all in-sync replicas to ack the message
	config.Producer.Retry.Max = 10                   // Retry up to 10 times to produce the message
	config.Producer.Return.Successes = true

	tlsConfig := &tls.Config{}
	if len(cfg.KafkaCaCertPem) > 0 {
		certPool := x509.NewCertPool()
		certPool.AppendCertsFromPEM([]byte(cfg.KafkaCaCertPem))
		tlsConfig = &tls.Config{RootCAs: certPool}
	}

	config.Net.TLS.Enable = true
	config.Net.TLS.Config = tlsConfig

	config.Net.SASL.Enable = true
	config.Net.SASL.User = cfg.KafkaUsername
	config.Net.SASL.Password = cfg.KafkaPassword
	config.Net.SASL.SCRAMClientGeneratorFunc = scramClientGeneratorFunc
	config.Net.SASL.Mechanism = sarama.SASLTypeSCRAMSHA512

	return config
}

func scramClientGeneratorFunc() sarama.SCRAMClient {
	var SHA512 scram.HashGeneratorFcn = func() hash.Hash { return sha512.New() }
	return &XDGSCRAMClient{HashGeneratorFcn: SHA512}
}

// See https://github.com/Shopify/sarama/blob/ceadf4f6b74eb2ca0b6108fd96778032b0fa404f/examples/sasl_scram_client/scram_client.go
type XDGSCRAMClient struct {
	*scram.Client
	*scram.ClientConversation
	scram.HashGeneratorFcn
}

func (x *XDGSCRAMClient) Begin(userName, password, authzID string) (err error) {
	x.Client, err = x.HashGeneratorFcn.NewClient(userName, password, authzID)
	if err != nil {
		return err
	}
	x.ClientConversation = x.Client.NewConversation()
	return nil
}

func (x *XDGSCRAMClient) Step(challenge string) (response string, err error) {
	response, err = x.ClientConversation.Step(challenge)
	return
}

func (x *XDGSCRAMClient) Done() bool {
	return x.ClientConversation.Done()
}
