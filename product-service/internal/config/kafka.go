package config

import (
	"github.com/IBM/sarama"
	"github.com/knadh/koanf/v2"
	"github.com/rs/zerolog"
	"strings"
)

func NewKafkaProducer(config *koanf.Koanf, log *zerolog.Logger) sarama.SyncProducer {
	kafka_port := config.String("KAFKA_BROKER_PORT")
	kafka_array_port := strings.Split(kafka_port, ";")

	producer, err := sarama.NewSyncProducer(kafka_array_port, nil)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create kafka producer")
	}

	return producer
}

func NewKafkaConsumer(config *koanf.Koanf, log *zerolog.Logger) sarama.Consumer {
	kafka_port := config.String("KAFKA_BROKER_PORT")
	kafka_array_port := strings.Split(kafka_port, ";")

	consumer, err := sarama.NewConsumer(kafka_array_port, nil)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create kafka consumer")
	}

	return consumer
}
