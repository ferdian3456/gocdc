package main

import (
	"context"
	"gocdc/internal/config"
	"gocdc/internal/exception"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func CORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		writer.Header().Set("Access-Control-Allow-Origin", "*")
		writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		writer.Header().Set("Access-Control-Allow-Headers", "X-Requested-With, Content-Type, Authorization")
		writer.Header().Set("Access-Control-Allow-Credentials", "True")
		if request.Method == http.MethodOptions {
			writer.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(writer, request)
	})
}

func main() {
	router := config.NewRouter()
	koanf := config.NewKoanf()
	zerolog := config.NewZeroLog()
	elasticsearch := config.NewElasticClient(koanf, &zerolog)
	kafkaProducer := config.NewKafkaProducer(koanf, &zerolog)
	kafkaConsumer := config.NewKafkaConsumer(koanf, &zerolog)
	db := config.NewDB(koanf, &zerolog)
	validator := config.NewValidator()

	config.Server(&config.ServerConfig{
		Router:        router,
		DB:            db,
		ElasticSearch: elasticsearch,
		KafkaProducer: kafkaProducer,
		KafkaConsumer: kafkaConsumer,
		Config:        koanf,
		Validate:      validator,
		Log:           &zerolog,
	})

	router.PanicHandler = exception.ErrorHandler

	GO_SERVER_PORT := koanf.String("GO_SERVER")

	server := http.Server{
		Addr:    GO_SERVER_PORT,
		Handler: CORS(router),
	}

	zerolog.Info().Msg("Server is running on " + GO_SERVER_PORT)

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			zerolog.Fatal().Err(err).Msg("Error Starting Server")
		}
	}()

	<-stop
	zerolog.Info().Msg("Got one of stop signals")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		zerolog.Fatal().Err(err).Msg("Timeout, forced kill!")
	}

	zerolog.Info().Msg("Server has shut down gracefully")
}
