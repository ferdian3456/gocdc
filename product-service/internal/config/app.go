package config

import (
	"database/sql"
	"github.com/IBM/sarama"
	"github.com/elastic/go-elasticsearch/v8"
	"github.com/go-playground/validator"
	"github.com/julienschmidt/httprouter"
	"github.com/knadh/koanf/v2"
	"github.com/rs/zerolog"
	"gocdc/internal/delivery/http"
	"gocdc/internal/delivery/http/middleware"
	"gocdc/internal/delivery/http/route"
	"gocdc/internal/repository"
	"gocdc/internal/usecase"
)

type ServerConfig struct {
	UserServiceUrl string
	Router         *httprouter.Router
	DB             *sql.DB
	ElasticSearch  *elasticsearch.Client
	KafkaProducer  sarama.SyncProducer
	KafkaConsumer  sarama.Consumer
	Log            *zerolog.Logger
	Validate       *validator.Validate
	Config         *koanf.Koanf
}

func Server(config *ServerConfig) {
	productRepository := repository.NewProductRepository(config.Log, config.DB)
	productUsecase := usecase.NewProductUsecase(config.UserServiceUrl, productRepository, config.KafkaProducer, config.DB, config.ElasticSearch, config.Validate, config.Log, config.Config)
	productController := http.NewProductController(productUsecase, config.Log)

	authMiddleware := middleware.NewAuthMiddleware(config.Router, config.Log, config.Config, productUsecase)

	routeConfig := route.RouteConfig{
		Router:            config.Router,
		ProductController: productController,
		AuthMiddleware:    authMiddleware,
	}

	routeConfig.SetupRoute()
}
