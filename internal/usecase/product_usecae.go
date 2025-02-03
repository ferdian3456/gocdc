package usecase

import (
	"context"
	"database/sql"
	"errors"
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
	DB                *sql.DB
	ElasticSearch     *elasticsearch.Client
	Validator         *validator.Validate
	Log               *zerolog.Logger
	Koanf             *koanf.Koanf
}

func NewProductUsecase(userRepository *repository.UserRepository, productRepository *repository.ProductRepository, db *sql.DB, elasticsearch *elasticsearch.Client, validator *validator.Validate, zerolog *zerolog.Logger, koanf *koanf.Koanf) *ProductUsecase {
	return &ProductUsecase{
		UserRepository:    userRepository,
		ProductRepository: productRepository,
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

	now := time.Now()
	product := domain.Product{
		Seller_id:   userUUID,
		Name:        request.Name,
		Quantity:    request.Quantity,
		Price:       request.Price,
		Weight:      request.Weight,
		Size:        request.Size,
		Description: request.Description,
		Created_at:  &now,
		Updated_at:  &now,
	}

	usecase.ProductRepository.CreateWithTx(ctx, tx, product)

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
