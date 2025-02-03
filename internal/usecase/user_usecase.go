package usecase

import (
	"context"
	"database/sql"
	"errors"
	"github.com/go-playground/validator"
	"github.com/golang-jwt/jwt/v5"
	googleuuid "github.com/google/uuid"
	"github.com/knadh/koanf/v2"
	"github.com/rs/zerolog"
	"gocdc/internal/helper"
	"gocdc/internal/model/domain"
	"gocdc/internal/model/web/user"
	"gocdc/internal/repository"
	"golang.org/x/crypto/bcrypt"
	"time"
)

type UserUsecase struct {
	UserRepository *repository.UserRepository
	DB             *sql.DB
	Validator      *validator.Validate
	Log            *zerolog.Logger
	Config         *koanf.Koanf
}

func NewUserUsecase(userRepository *repository.UserRepository, db *sql.DB, validator *validator.Validate, zerolog *zerolog.Logger, koanf *koanf.Koanf) *UserUsecase {
	return &UserUsecase{
		UserRepository: userRepository,
		DB:             db,
		Validator:      validator,
		Log:            zerolog,
		Config:         koanf,
	}
}

func (usecase *UserUsecase) Register(ctx context.Context, request user.UserRegisterRequest) (string, error) {
	err := usecase.Validator.Struct(request)
	if err != nil {
		respErr := errors.New("invalid request body")
		usecase.Log.Warn().Err(respErr).Msg(err.Error())
		return "", respErr
	}

	// start transaction to ensure that credential is unique before create/insert data
	tx, err := usecase.DB.Begin()
	if err != nil {
		respErr := errors.New("failed to start transaction")
		usecase.Log.Panic().Err(err).Msg(respErr.Error())
	}

	defer helper.CommitOrRollback(tx)

	now := time.Now()
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(request.Password), bcrypt.DefaultCost)
	if err != nil {
		respErr := errors.New("error generating password hash")
		usecase.Log.Panic().Err(err).Msg(respErr.Error())
	}

	uuid := googleuuid.New()

	user := domain.User{
		Id:          uuid.String(),
		Name:        request.Name,
		Email:       request.Email,
		Password:    string(hashedPassword),
		Address:     request.Address,
		PhoneNumber: request.PhoneNumber,
		Created_at:  &now,
		Updated_at:  &now,
	}

	err = usecase.UserRepository.CheckCredentialUniqueWithTx(ctx, tx, user)
	if err != nil {
		usecase.Log.Warn().Msg(err.Error())
		return "", err
	}

	usecase.UserRepository.RegisterWithTx(ctx, tx, user)

	secretKey := usecase.Config.String("SECRET_KEY")
	secretKeyByte := []byte(secretKey)

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"id":      user.Id,
		"expired": time.Date(2030, 10, 10, 12, 0, 0, 0, time.UTC).Unix(),
	})

	tokenString, err := token.SignedString(secretKeyByte)
	if err != nil {
		respErr := errors.New("failed to sign a token")
		usecase.Log.Panic().Err(err).Msg(respErr.Error())
	}

	return tokenString, nil
}

func (usecase *UserUsecase) Login(ctx context.Context, request user.UserLoginRequest) (string, error) {
	err := usecase.Validator.Struct(request)
	if err != nil {
		respErr := errors.New("invalid request body")
		usecase.Log.Warn().Err(respErr).Msg(err.Error())
		return "", respErr
	}

	user, err := usecase.UserRepository.Login(ctx, request.Email)
	if err != nil {
		usecase.Log.Warn().Msg(err.Error())
		return "", err
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(request.Password))
	if err != nil {
		respErr := errors.New("wrong password")
		usecase.Log.Warn().Err(respErr).Msg(err.Error())
		return "", respErr
	}

	secretKey := usecase.Config.String("SECRET_KEY")
	secretKeyByte := []byte(secretKey)

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"id":      user.Id,
		"expired": time.Date(2030, 10, 10, 12, 0, 0, 0, time.UTC).Unix(),
	})

	tokenString, err := token.SignedString(secretKeyByte)
	if err != nil {
		respErr := errors.New("failed to sign a token")
		usecase.Log.Panic().Err(err).Msg(respErr.Error())
	}

	return tokenString, nil
}

func (usecase *UserUsecase) Update(ctx context.Context, request user.UserUpdateRequest, userUUID string) error {
	err := usecase.Validator.Struct(request)
	if err != nil {
		respErr := errors.New("invalid request body")
		usecase.Log.Warn().Err(respErr).Msg(err.Error())
		return respErr
	}

	now := time.Now()

	user := domain.User{
		Id:          userUUID,
		Name:        request.Name,
		Email:       request.Email,
		Password:    request.Password,
		Address:     request.Address,
		PhoneNumber: request.PhoneNumber,
		Updated_at:  &now,
	}

	usecase.UserRepository.Update(ctx, user)

	return nil
}

func (usecase *UserUsecase) Delete(ctx context.Context, userUUID string) error {
	err := usecase.UserRepository.CheckUserExistence(ctx, userUUID)
	if err != nil {
		usecase.Log.Warn().Msg(err.Error())
		return err
	}

	usecase.UserRepository.Delete(ctx, userUUID)

	return nil
}

func (usecase *UserUsecase) FindUserInfo(ctx context.Context, userUUID string) (user.UserResponse, error) {
	user, err := usecase.UserRepository.FindUserInfo(ctx, userUUID)
	if err != nil {
		usecase.Log.Warn().Msg(err.Error())
		return user, err
	}

	return user, nil
}

func (usecase *UserUsecase) CheckUserExistance(ctx context.Context, userUUID string) error {
	err := usecase.UserRepository.CheckUserExistence(ctx, userUUID)
	if err != nil {
		usecase.Log.Warn().Msg(err.Error())
		return err
	}

	return nil
}
