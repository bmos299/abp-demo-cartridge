// ------------------------------------------------------ {COPYRIGHT-TOP} ---
// Licensed Materials - Property of IBM
// 5900-AEO
//
// Copyright IBM Corp. 2020, 2021. All Rights Reserved.
//
// US Government Users Restricted Rights - Use, duplication, or
// disclosure restricted by GSA ADP Schedule Contract with IBM Corp.
// ------------------------------------------------------ {COPYRIGHT-END} ---
package config

import (
	"github.com/caarlos0/env"
)

// Config for Demo Cartridge
type Config struct {
	Function                 string `env:"FUNCTION"`
	BootstrapServers         string `env:"BOOTSTRAP_SERVERS"`
	KafkaCaCertPem           string `env:"KAFKA_CA_CERT_PEM"`
	KafkaUsername            string `env:"KAFKA_USERNAME"`
	KafkaPassword            string `env:"KAFKA_PASSWORD"`
	KafkaTopic               string `env:"KAFKA_TOPIC"`
	ServerImage              string `env:"SERVER_IMAGE"`
	EventProcessorImage      string `env:"EVENT_PROCESSOR_IMAGE"`
	EventProcessingTaskImage string `env:"EVENT_PROCESSING_TASK_IMAGE"`
	UseKnative               string `env:"USE_KNATIVE"`
	MessagesPerGroup         string `env:"MESSAGES_PER_GROUP"`
	SecondsToPause           string `env:"SECONDS_TO_PAUSE"`
	SequenceRepititions      string `env:"SEQUENCE_REPITITIONS"`
}

// Parse environment variable to config struct
func Parse() (*Config, error) {
	cfg := &Config{}
	if err := env.Parse(cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}
