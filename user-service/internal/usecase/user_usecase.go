package usecase

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/IBM/sarama"
	"github.com/go-playground/validator"
	"github.com/golang-jwt/jwt/v5"
	googleuuid "github.com/google/uuid"
	"github.com/knadh/koanf/v2"
	"github.com/rs/zerolog"
	"gocdc/internal/helper"
	"gocdc/internal/model/domain"
	"gocdc/internal/model/web"
	"gocdc/internal/model/web/user"
	"gocdc/internal/repository"
	"golang.org/x/crypto/bcrypt"
	"net/http"
	"time"
)

type UserUsecase struct {
	UserRepository *repository.UserRepository
	KafkaWriter    sarama.SyncProducer
	DB             *sql.DB
	Validator      *validator.Validate
	Log            *zerolog.Logger
	Config         *koanf.Koanf
}

func NewUserUsecase(userRepository *repository.UserRepository, kafkaWriter sarama.SyncProducer, db *sql.DB, validator *validator.Validate, zerolog *zerolog.Logger, koanf *koanf.Koanf) *UserUsecase {
	return &UserUsecase{
		UserRepository: userRepository,
		KafkaWriter:    kafkaWriter,
		DB:             db,
		Validator:      validator,
		Log:            zerolog,
		Config:         koanf,
	}
}

func (usecase *UserUsecase) Register(ctx context.Context, request user.UserRegisterRequest) (web.TokenResponse, error) {
	err := usecase.Validator.Struct(request)
	if err != nil {
		respErr := errors.New("invalid request body")
		usecase.Log.Warn().Err(respErr).Msg(err.Error())
		return web.TokenResponse{}, respErr
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

	userAuditEvent := user.AuditEvent{}
	userNotificationEvent := user.NotificationEvent{}
	userVerificationEvent := user.VerificationEvent{}

	user := domain.User{
		Id:              uuid.String(),
		Profile_picture: request.Profile_picture,
		Name:            request.Name,
		Email:           request.Email,
		Password:        string(hashedPassword),
		Address:         request.Address,
		PhoneNumber:     request.PhoneNumber,
		Created_at:      &now,
		Updated_at:      &now,
	}

	err = usecase.UserRepository.CheckCredentialUniqueWithTx(ctx, tx, user)
	if err != nil {
		usecase.Log.Warn().Msg(err.Error())
		return web.TokenResponse{}, err
	}

	usecase.UserRepository.RegisterWithTx(ctx, tx, user)

	secretKeyAccess := usecase.Config.String("SECRET_KEY_ACCESS_TOKEN")
	secretKeyAccessByte := []byte(secretKeyAccess)

	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"id":  user.Id,
		"exp": now.Add(5 * time.Minute).Unix(),
	})

	accessTokenString, err := accessToken.SignedString(secretKeyAccessByte)
	if err != nil {
		respErr := errors.New("failed to sign a token")
		usecase.Log.Panic().Err(err).Msg(respErr.Error())
	}

	secretKeyRefresh := usecase.Config.String("SECRET_KEY_REFRESH_TOKEN")
	secretKeyRefreshByte := []byte(secretKeyRefresh)

	addedTime := now.Add(30 * 24 * time.Hour)
	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"id":  user.Id,
		"exp": addedTime.Unix(),
	})

	refreshTokenString, err := refreshToken.SignedString(secretKeyRefreshByte)
	if err != nil {
		respErr := errors.New("failed to sign a token")
		usecase.Log.Panic().Err(err).Msg(respErr.Error())
	}

	refreshTokenHash := sha256.New()
	refreshTokenHash.Write([]byte(refreshTokenString))
	hashedRefreshToken := refreshTokenHash.Sum(nil)

	hashedRefreshTokenHex := hex.EncodeToString(hashedRefreshToken)

	refreshTokenToDB := domain.RefreshToken{
		User_id:              uuid.String(),
		Hashed_refresh_token: hashedRefreshTokenHex,
		Created_at:           &now,
		Expired_at:           &addedTime,
	}

	usecase.UserRepository.AddRefreshTokenWithTx(ctx, tx, refreshTokenToDB)

	tokenResponse := web.TokenResponse{
		Access_Token:  accessTokenString,
		Refresh_Token: refreshTokenString,
	}

	topic := "user.activity"

	userAuditEvent.Id = uuid.String()
	userAuditEvent.Event = "Create"
	userAuditEvent.Created_at = &now

	messageJSON, err := json.Marshal(userAuditEvent)
	if err != nil {
		respErr := errors.New("failed to marshal a json")
		usecase.Log.Panic().Err(err).Msg(respErr.Error())
	}

	_, _, err = usecase.KafkaWriter.SendMessage(&sarama.ProducerMessage{
		Topic: topic,
		Value: sarama.ByteEncoder(messageJSON),
	})

	if err != nil {
		respErr := errors.New("failed to produce an event to kafka broker")
		usecase.Log.Panic().Err(err).Msg(respErr.Error())
	}

	topic = "user.notification"

	userNotificationEvent.Id = uuid.String()
	userNotificationEvent.Email = request.Email
	userNotificationEvent.Event = "Create"
	userNotificationEvent.Created_at = &now

	messageJSON1, err := json.Marshal(userNotificationEvent)
	if err != nil {
		respErr := errors.New("failed to marshal a json")
		usecase.Log.Panic().Err(err).Msg(respErr.Error())
	}

	_, _, err = usecase.KafkaWriter.SendMessage(&sarama.ProducerMessage{
		Topic: topic,
		Value: sarama.ByteEncoder(messageJSON1),
	})

	if err != nil {
		respErr := errors.New("failed to produce an event to kafka broker")
		usecase.Log.Panic().Err(err).Msg(respErr.Error())
	}

	topic = "user.verification"

	userVerificationEvent.Id = uuid.String()
	userVerificationEvent.Profile_picture = request.Profile_picture

	messageJSON2, err := json.Marshal(userVerificationEvent)
	if err != nil {
		respErr := errors.New("failed to marshal a json")
		usecase.Log.Panic().Err(err).Msg(respErr.Error())
	}

	_, _, err = usecase.KafkaWriter.SendMessage(&sarama.ProducerMessage{
		Topic: topic,
		Value: sarama.ByteEncoder(messageJSON2),
	})

	if err != nil {
		respErr := errors.New("failed to produce an event to kafka broker")
		usecase.Log.Panic().Err(err).Msg(respErr.Error())
	}

	return tokenResponse, nil
}

func (usecase *UserUsecase) Login(ctx context.Context, request user.UserLoginRequest) (web.TokenResponse, error) {
	err := usecase.Validator.Struct(request)
	if err != nil {
		respErr := errors.New("invalid request body")
		usecase.Log.Warn().Err(respErr).Msg(err.Error())
		return web.TokenResponse{}, respErr
	}

	tx, err := usecase.DB.Begin()
	if err != nil {
		respErr := errors.New("failed to start transaction")
		usecase.Log.Panic().Err(err).Msg(respErr.Error())
	}

	defer helper.CommitOrRollback(tx)

	user, err := usecase.UserRepository.LoginWithTx(ctx, tx, request.Email)
	if err != nil {
		usecase.Log.Warn().Msg(err.Error())
		return web.TokenResponse{}, err
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(request.Password))
	if err != nil {
		respErr := errors.New("wrong password")
		usecase.Log.Warn().Err(respErr).Msg(err.Error())
		return web.TokenResponse{}, respErr
	}

	secretKeyAccess := usecase.Config.String("SECRET_KEY_ACCESS_TOKEN")
	secretKeyAccessByte := []byte(secretKeyAccess)

	now := time.Now()

	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"id":  user.Id,
		"exp": now.Add(5 * time.Minute).Unix(),
	})

	accessTokenString, err := accessToken.SignedString(secretKeyAccessByte)
	if err != nil {
		respErr := errors.New("failed to sign a token")
		usecase.Log.Panic().Err(err).Msg(respErr.Error())
	}

	secretKeyRefresh := usecase.Config.String("SECRET_KEY_REFRESH_TOKEN")
	secretKeyRefreshByte := []byte(secretKeyRefresh)

	addedTime := now.Add(30 * 24 * time.Hour)
	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"id":  user.Id,
		"exp": addedTime.Unix(),
	})

	refreshTokenString, err := refreshToken.SignedString(secretKeyRefreshByte)
	if err != nil {
		respErr := errors.New("failed to sign a token")
		usecase.Log.Panic().Err(err).Msg(respErr.Error())
	}

	refreshTokenHash := sha256.New()
	refreshTokenHash.Write([]byte(refreshTokenString))
	hashedRefreshToken := refreshTokenHash.Sum(nil)

	hashedRefreshTokenHex := hex.EncodeToString(hashedRefreshToken)

	refreshTokenToDB := domain.RefreshToken{
		User_id:              user.Id,
		Hashed_refresh_token: hashedRefreshTokenHex,
		Created_at:           &now,
		Expired_at:           &addedTime,
	}

	usecase.UserRepository.UpdateRefreshToken(ctx, tx, "Revoke", user.Id)
	usecase.UserRepository.AddRefreshTokenWithTx(ctx, tx, refreshTokenToDB)

	tokenResponse := web.TokenResponse{
		Access_Token:  accessTokenString,
		Refresh_Token: refreshTokenString,
	}

	return tokenResponse, nil
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

func (usecase *UserUsecase) TokenRenewal(ctx context.Context, request user.RenewalTokenRequest) (web.TokenResponse, error) {
	err := usecase.Validator.Struct(request)
	if err != nil {
		respErr := errors.New("invalid request body")
		usecase.Log.Warn().Err(respErr).Msg(err.Error())
		return web.TokenResponse{}, respErr
	}

	tx, err := usecase.DB.Begin()
	if err != nil {
		respErr := errors.New("failed to start transaction")
		usecase.Log.Panic().Err(err).Msg(respErr.Error())
	}

	defer helper.CommitOrRollback(tx)

	secretKeyRefresh := usecase.Config.String("SECRET_KEY_REFRESH_TOKEN")
	secretKeyRefreshByte := []byte(secretKeyRefresh)

	token, err := jwt.Parse(request.Refresh_token, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, http.ErrNotSupported
		}
		return secretKeyRefreshByte, nil
	})

	if err != nil {
		if err == jwt.ErrTokenMalformed {
			respErr := errors.New("Token is malformed")
			usecase.Log.Warn().Err(respErr).Msg(err.Error())
			return web.TokenResponse{}, respErr
		} else if err.Error() == "token has invalid claims: token is expired" {

		} else {
			respErr := errors.New("Invalid token")
			usecase.Log.Warn().Err(respErr).Msg(err.Error())
			return web.TokenResponse{}, respErr
		}
	}

	var id string
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		if val, exists := claims["id"]; exists {
			if strVal, ok := val.(string); ok {
				id = strVal
			}
		} else {
			respErr := errors.New("Invalid token")
			usecase.Log.Warn().Err(respErr).Msg(err.Error())
			return web.TokenResponse{}, respErr
		}
	}

	err = usecase.UserRepository.CheckUserExistenceWithTx(ctx, tx, id)
	if err != nil {
		respErr := errors.New("User not found")
		usecase.Log.Warn().Err(respErr).Msg(err.Error())
		return web.TokenResponse{}, respErr
	}

	requestRefreshTokenHash := sha256.New()
	requestRefreshTokenHash.Write([]byte(request.Refresh_token))
	hashedRequestRefreshToken := requestRefreshTokenHash.Sum(nil)

	hashedRequestRefreshTokenHex := hex.EncodeToString(hashedRequestRefreshToken)

	hashedDBRefreshTokenHex, err := usecase.UserRepository.FindLatestRefreshToken(ctx, tx)
	if err != nil {
		usecase.Log.Warn().Msg(err.Error())
		return web.TokenResponse{}, err
	}

	if hashedRequestRefreshTokenHex != hashedDBRefreshTokenHex {
		fmt.Println("hashed from request: ", hashedRequestRefreshTokenHex)
		fmt.Println("hashed from db :", hashedDBRefreshTokenHex)
		usecase.UserRepository.UpdateRefreshToken(ctx, tx, "Revoke", id)
		respErr := errors.New("Refresh token reuse detected. For security reasons, you have been logged out. Please sign in again.")
		usecase.Log.Warn().Msg(respErr.Error())
		return web.TokenResponse{}, respErr
	}

	secretKeyAccess := usecase.Config.String("SECRET_KEY_ACCESS_TOKEN")
	secretKeyAccessByte := []byte(secretKeyAccess)

	now := time.Now()

	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"id":  id,
		"exp": now.Add(5 * time.Minute).Unix(),
	})

	accessTokenString, err := accessToken.SignedString(secretKeyAccessByte)
	if err != nil {
		respErr := errors.New("failed to sign a token")
		usecase.Log.Panic().Err(err).Msg(respErr.Error())
	}

	addedTime := now.Add(30 * 24 * time.Hour)
	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"id":  id,
		"exp": addedTime.Unix(),
	})

	refreshTokenString, err := refreshToken.SignedString(secretKeyRefreshByte)
	if err != nil {
		respErr := errors.New("failed to sign a token")
		usecase.Log.Panic().Err(err).Msg(respErr.Error())
	}

	refreshTokenHash := sha256.New()
	refreshTokenHash.Write([]byte(refreshTokenString))
	hashedRefreshToken := refreshTokenHash.Sum(nil)

	hashedRefreshTokenHex := hex.EncodeToString(hashedRefreshToken)

	refreshTokenToDB := domain.RefreshToken{
		User_id:              id,
		Hashed_refresh_token: hashedRefreshTokenHex,
		Created_at:           &now,
		Expired_at:           &addedTime,
	}

	usecase.UserRepository.UpdateRefreshToken(ctx, tx, "Revoke", id)
	usecase.UserRepository.AddRefreshTokenWithTx(ctx, tx, refreshTokenToDB)

	tokenResponse := web.TokenResponse{
		Access_Token:  accessTokenString,
		Refresh_Token: refreshTokenString,
	}

	return tokenResponse, nil
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
