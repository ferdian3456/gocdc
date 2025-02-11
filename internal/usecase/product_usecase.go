package usecase

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"github.com/IBM/sarama"
	"github.com/elastic/go-elasticsearch/v8"
	"github.com/go-playground/validator"
	"github.com/knadh/koanf/v2"
	"github.com/rs/zerolog"
	"gocdc/internal/helper"
	"gocdc/internal/model/domain"
	"gocdc/internal/model/web/product"
	"gocdc/internal/repository"
	"time"
)

type ProductUsecase struct {
	UserRepository    *repository.UserRepository
	ProductRepository *repository.ProductRepository
	KafkaWriter       sarama.SyncProducer
	DB                *sql.DB
	ElasticSearch     *elasticsearch.Client
	Validator         *validator.Validate
	Log               *zerolog.Logger
	Koanf             *koanf.Koanf
}

func NewProductUsecase(userRepository *repository.UserRepository, kafkaWriter sarama.SyncProducer, productRepository *repository.ProductRepository, db *sql.DB, elasticsearch *elasticsearch.Client, validator *validator.Validate, zerolog *zerolog.Logger, koanf *koanf.Koanf) *ProductUsecase {
	return &ProductUsecase{
		UserRepository:    userRepository,
		ProductRepository: productRepository,
		KafkaWriter:       kafkaWriter,
		DB:                db,
		ElasticSearch:     elasticsearch,
		Validator:         validator,
		Log:               zerolog,
		Koanf:             koanf,
	}
}

func (usecase *ProductUsecase) Create(ctx context.Context, request product.ProductCreateRequest, userUUID string) error {
	err := usecase.Validator.Struct(request)
	if err != nil {
		respErr := errors.New("invalid request body")
		usecase.Log.Warn().Err(respErr).Msg(err.Error())
		return respErr
	}

	tx, err := usecase.DB.Begin()
	if err != nil {
		respErr := errors.New("failed to start transaction")
		usecase.Log.Panic().Err(err).Msg(respErr.Error())
	}

	defer helper.CommitOrRollback(tx)

	err = usecase.UserRepository.CheckUserExistenceWithTx(ctx, tx, userUUID)
	if err != nil {
		usecase.Log.Warn().Msg(err.Error())
		return err
	}

	userEmail, err := usecase.UserRepository.FindUserEmailByUUID(ctx, tx, userUUID)
	if err != nil {
		usecase.Log.Warn().Msg(err.Error())
		return err
	}

	productEvent := product.ProductEvent{}

	now := time.Now()
	product := domain.Product{
		Seller_id:       userUUID,
		Name:            request.Name,
		Product_picture: request.Product_picture,
		Quantity:        request.Quantity,
		Price:           request.Price,
		Weight:          request.Weight,
		Size:            request.Size,
		Description:     request.Description,
		Created_at:      &now,
		Updated_at:      &now,
	}

	usecase.ProductRepository.CreateWithTx(ctx, tx, product)

	topic := "product.activity"

	productEvent.Id = userUUID
	productEvent.Email = *userEmail
	productEvent.Event = "Create"
	productEvent.Created_at = &now

	messageJSON, err := json.Marshal(productEvent)
	if err != nil {
		respErr := errors.New("failed to marshal a json")
		usecase.Log.Panic().Err(err).Msg(respErr.Error())
	}

	_, _, err = usecase.KafkaWriter.SendMessage(&sarama.ProducerMessage{
		Topic: topic,
		Value: sarama.ByteEncoder(messageJSON),
	})

	return nil
}

func (usecase *ProductUsecase) Update(ctx context.Context, request product.ProductUpdateRequest, userUUID string, productID int) error {
	err := usecase.Validator.Struct(request)
	if err != nil {
		respErr := errors.New("invalid request body")
		usecase.Log.Warn().Err(respErr).Msg(err.Error())
		return respErr
	}

	tx, err := usecase.DB.Begin()
	if err != nil {
		respErr := errors.New("failed to start transaction")
		usecase.Log.Panic().Err(err).Msg(respErr.Error())
	}

	defer helper.CommitOrRollback(tx)

	err = usecase.ProductRepository.CheckOwnershipWithTx(ctx, tx, userUUID, productID)
	if err != nil {
		usecase.Log.Warn().Msg(err.Error())
		return err
	}

	now := time.Now()

	product := domain.Product{
		Id:          productID,
		Name:        request.Name,
		Quantity:    request.Quantity,
		Price:       request.Price,
		Weight:      request.Weight,
		Size:        request.Size,
		Description: request.Description,
		Updated_at:  &now,
	}

	usecase.ProductRepository.UpdateWithTx(ctx, tx, product)

	return nil
}

func (usecase *ProductUsecase) Delete(ctx context.Context, userUUID string, productID int) error {
	err := usecase.ProductRepository.CheckOwnership(ctx, userUUID, productID)
	if err != nil {
		usecase.Log.Warn().Msg(err.Error())
		return err
	}

	usecase.ProductRepository.Delete(ctx, productID)

	return nil
}

func (usecase *ProductUsecase) FindProductInfo(ctx context.Context, productID int) (product.ProductResponse, error) {
	productResponse, err := usecase.ProductRepository.FindProductInfo(ctx, productID)
	if err != nil {
		usecase.Log.Warn().Msg(err.Error())
		return productResponse, err
	}

	return productResponse, nil
}

func (usecase *ProductUsecase) FindAllProduct(ctx context.Context) ([]product.ProductResponse, error) {
	productResponse, err := usecase.ProductRepository.FindAllProduct(ctx)
	if err != nil {
		usecase.Log.Warn().Msg(err.Error())
		return productResponse, err
	}

	return productResponse, nil
}

func (usecase *ProductUsecase) FindProductHomePage(ctx context.Context) ([]product.ProductHomePageResponse, error) {
	tx, err := usecase.DB.Begin()
	if err != nil {
		respErr := errors.New("failed to start transaction")
		usecase.Log.Panic().Err(err).Msg(respErr.Error())
	}

	defer helper.CommitOrRollback(tx)

	productResponse, err := usecase.ProductRepository.FindAllProductWithTx(ctx, tx)
	if err != nil {
		usecase.Log.Warn().Msg(err.Error())
		return []product.ProductHomePageResponse{}, err
	}

	productHomePageResponses := []product.ProductHomePageResponse{}

	for _, productArray := range productResponse {
		userResponse, err := usecase.UserRepository.FindUserInfoWithTx(ctx, tx, productArray.Seller_id)
		if err != nil {
			usecase.Log.Warn().Msg(err.Error())
			return []product.ProductHomePageResponse{}, err
		}

		productHomePageResponse := product.ProductHomePageResponse{
			Id:             productArray.Id,
			Seller_id:      productArray.Seller_id,
			Seller_name:    userResponse.Name,
			Seller_address: userResponse.Address,
			Name:           productArray.Name,
			Quantity:       productArray.Quantity,
			Price:          productArray.Price,
			Weight:         productArray.Weight,
			Size:           productArray.Size,
			Status:         productArray.Status,
			Description:    productArray.Description,
			Created_at:     productArray.Created_at,
			Updated_at:     productArray.Updated_at,
		}

		productHomePageResponses = append(productHomePageResponses, productHomePageResponse)
	}

	return productHomePageResponses, nil
}
