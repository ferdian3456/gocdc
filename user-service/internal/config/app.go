package config

import (
	"database/sql"
	"github.com/IBM/sarama"
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
	Router        *httprouter.Router
	DB            *sql.DB
	KafkaProducer sarama.SyncProducer
	KafkaConsumer sarama.Consumer
	Log           *zerolog.Logger
	Validate      *validator.Validate
	Config        *koanf.Koanf
}

func Server(config *ServerConfig) {
	userRepository := repository.NewUserRepository(config.Log, config.DB)
	userUsecase := usecase.NewUserUsecase(userRepository, config.KafkaProducer, config.DB, config.Validate, config.Log, config.Config)
	userController := http.NewUserController(userUsecase, config.Log)

	authMiddleware := middleware.NewAuthMiddleware(config.Router, config.Log, config.Config, userUsecase)

	routeConfig := route.RouteConfig{
		Router:         config.Router,
		UserController: userController,
		AuthMiddleware: authMiddleware,
	}

	routeConfig.SetupRoute()
}
