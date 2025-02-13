package config

import (
	"github.com/elastic/go-elasticsearch/v8"
	"github.com/knadh/koanf/v2"
	"github.com/rs/zerolog"
)

func NewElasticClient(config *koanf.Koanf, log *zerolog.Logger) *elasticsearch.Client {
	cfg := elasticsearch.Config{
		Addresses: []string{
			config.String("ELASTICSEARCH_URI"),
		},
		Username: config.String("ELASTICSEARCH_USERNAME"),
		Password: config.String("ELASTICSEARCH_PASSWORD"),
	}

	esClient, err := elasticsearch.NewClient(cfg)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create elastic search client")
	}

	return esClient
}
