package usecase

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/IBM/sarama"
	"github.com/elastic/go-elasticsearch/v8"
	"github.com/go-playground/validator"
	"github.com/knadh/koanf/v2"
	"github.com/rs/zerolog"
	"gocdc/internal/helper"
	"gocdc/internal/model/domain"
	"gocdc/internal/model/web"
	"gocdc/internal/model/web/product"
	"gocdc/internal/repository"
	"io/ioutil"
	"net/http"
	"time"
)

type ProductUsecase struct {
	UserServiceUrl    string
	ProductRepository *repository.ProductRepository
	KafkaWriter       sarama.SyncProducer
	DB                *sql.DB
	ElasticSearch     *elasticsearch.Client
	Validator         *validator.Validate
	Log               *zerolog.Logger
	Koanf             *koanf.Koanf
}

func NewProductUsecase(userServiceUrl string, productRepository *repository.ProductRepository, kafkaWriter sarama.SyncProducer, db *sql.DB, elasticsearch *elasticsearch.Client, validator *validator.Validate, zerolog *zerolog.Logger, koanf *koanf.Koanf) *ProductUsecase {
	return &ProductUsecase{
		UserServiceUrl:    userServiceUrl,
		ProductRepository: productRepository,
		KafkaWriter:       kafkaWriter,
		DB:                db,
		ElasticSearch:     elasticsearch,
		Validator:         validator,
		Log:               zerolog,
		Koanf:             koanf,
	}
}

func (usecase *ProductUsecase) Create(ctx context.Context, request product.ProductCreateRequest, userUUID string, userAuthToken string) error {
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

	_, err = usecase.FindUserExistenceAPI(ctx, userAuthToken)
	if err != nil {
		usecase.Log.Warn().Msg(err.Error())
		return err
	}

	userEmail, err := usecase.FindUserEmailByUUIDAPI(ctx, userAuthToken)
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
	productEvent.Email = userEmail
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

func (usecase *ProductUsecase) FindUserExistenceAPI(ctx context.Context, userAuthToken string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("%sexistence", usecase.UserServiceUrl), nil)
	if err != nil {
		return "", err
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", userAuthToken))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var apiResp web.ExistenceApiResponse
	err = json.Unmarshal(body, &apiResp)
	if err != nil {
		return "", err
	}

	return apiResp.Data.Status, nil
}

func (usecase *ProductUsecase) FindUserEmailByUUIDAPI(ctx context.Context, userAuthToken string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("%semail", usecase.UserServiceUrl), nil)
	if err != nil {
		return "", err
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", userAuthToken))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var apiResp web.EmailApiResponse
	err = json.Unmarshal(body, &apiResp)
	if err != nil {
		return "", err
	}

	return apiResp.Data.Email, nil
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
		userResponse, err := usecase.UserRepository.FindUserInfoWith(ctx, tx, productArray.Seller_id)
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

func (usecase *ProductUsecase) FindUserInfoAPI(ctx context.Context, )
