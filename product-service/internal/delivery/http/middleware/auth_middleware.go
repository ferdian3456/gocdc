package middleware

import (
	"context"
	"github.com/golang-jwt/jwt/v5"
	"github.com/julienschmidt/httprouter"
	"github.com/knadh/koanf/v2"
	"github.com/rs/zerolog"
	"gocdc/internal/helper"
	"gocdc/internal/model/web"
	"gocdc/internal/usecase"
	"net/http"
	"strings"
)

const (
	userUUIDkey   = "user_uuid"
	userAuthToken = "user_auth_token"
)

type AuthMiddleware struct {
	Handler        http.Handler
	Log            *zerolog.Logger
	Config         *koanf.Koanf
	ProductUsecase *usecase.ProductUsecase
}

func NewAuthMiddleware(handler http.Handler, zerolog *zerolog.Logger, koanf *koanf.Koanf, productUsecase *usecase.ProductUsecase) *AuthMiddleware {
	return &AuthMiddleware{
		Handler:        handler,
		Log:            zerolog,
		Config:         koanf,
		ProductUsecase: productUsecase,
	}
}

func (middleware *AuthMiddleware) ServeHTTP(next httprouter.Handle) httprouter.Handle {
	return func(writer http.ResponseWriter, request *http.Request, p httprouter.Params) {
		headerToken := request.Header.Get("Authorization")

		if headerToken == "" {
			writer.Header().Set("Content-Type", "application/json")
			writer.WriteHeader(http.StatusUnauthorized)

			webResponse := web.WebResponse{
				Code:   http.StatusUnauthorized,
				Status: "Unauthorized",
				Data:   "No token provided",
			}

			middleware.Log.Warn().Msg("Unauthorized, no token provided")
			helper.WriteToResponseBody(writer, webResponse)
			return
		}

		splitToken := strings.Split(headerToken, "Bearer ")
		if len(splitToken) != 2 {
			writer.Header().Set("Content-Type", "application/json")
			writer.WriteHeader(http.StatusUnauthorized)

			webResponse := web.WebResponse{
				Code:   http.StatusUnauthorized,
				Status: "Unauthorized",
				Data:   "Token format is not match",
			}

			middleware.Log.Warn().Msg("Unauthorized, token format is not match")
			helper.WriteToResponseBody(writer, webResponse)
			return
		}

		secretKey := middleware.Config.String("SECRET_KEY_ACCESS_TOKEN")
		secretKeyByte := []byte(secretKey)

		token, err := jwt.Parse(splitToken[1], func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, http.ErrNotSupported
			}
			return secretKeyByte, nil
		})

		if err != nil {
			if err == jwt.ErrTokenMalformed {
				writer.Header().Set("Content-Type", "application/json")
				writer.WriteHeader(http.StatusUnauthorized)

				webResponse := web.WebResponse{
					Code:   http.StatusUnauthorized,
					Status: "Unauthorized",
					Data:   "Token is malformed",
				}

				middleware.Log.Warn().Msg("Unauthorized, invalid token")
				helper.WriteToResponseBody(writer, webResponse)
				return
			} else if err.Error() == "token has invalid claims: token is expired" {
				writer.Header().Set("Content-Type", "application/json")
				writer.WriteHeader(http.StatusUnauthorized)

				webResponse := web.WebResponse{
					Code:   http.StatusUnauthorized,
					Status: "Unauthorized",
					Data:   "Token is expired",
				}

				middleware.Log.Warn().Msg("Unauthorized, invalid token")
				helper.WriteToResponseBody(writer, webResponse)
				return
			} else {
				writer.Header().Set("Content-Type", "application/json")
				writer.WriteHeader(http.StatusUnauthorized)

				webResponse := web.WebResponse{
					Code:   http.StatusUnauthorized,
					Status: "Unauthorized",
					Data:   "Invalid token",
				}

				middleware.Log.Warn().Msg("Unauthorized, invalid token")
				helper.WriteToResponseBody(writer, webResponse)
				return
			}
		}

		var id string
		if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
			if val, exists := claims["id"]; exists {
				if strVal, ok := val.(string); ok {
					id = strVal
				}
			} else {
				writer.Header().Set("Content-Type", "application/json")
				writer.WriteHeader(http.StatusUnauthorized)

				webResponse := web.WebResponse{
					Code:   http.StatusUnauthorized,
					Status: "Unauthorized",
					Data:   "Invalid Token",
				}

				middleware.Log.Warn().Msg("Unauthorized, invalid token")
				helper.WriteToResponseBody(writer, webResponse)
				return
			}
		}

		_, err = middleware.ProductUsecase.FindUserExistenceAPI(request.Context(), id)
		if err != nil {
			writer.Header().Set("Content-Type", "application/json")
			writer.WriteHeader(http.StatusUnauthorized)

			webResponse := web.WebResponse{
				Code:   http.StatusUnauthorized,
				Status: "Unauthorized",
				Data:   "User not found, please register",
			}

			middleware.Log.Warn().Msg("Unauthorized, user not found")
			helper.WriteToResponseBody(writer, webResponse)
			return
		}

		middleware.Log.Debug().Msg("User:" + id)
		ctx := context.WithValue(request.Context(), userUUIDkey, id)
		next(writer, request.WithContext(ctx), p)
	}
}

func (middleware *AuthMiddleware) ServeExternalService(next httprouter.Handle) httprouter.Handle {
	return func(writer http.ResponseWriter, request *http.Request, p httprouter.Params) {
		headerToken := request.Header.Get("Authorization")

		if headerToken == "" {
			writer.Header().Set("Content-Type", "application/json")
			writer.WriteHeader(http.StatusUnauthorized)

			webResponse := web.WebResponse{
				Code:   http.StatusUnauthorized,
				Status: "Unauthorized",
				Data:   "No token provided",
			}

			middleware.Log.Warn().Msg("Unauthorized, no token provided")
			helper.WriteToResponseBody(writer, webResponse)
			return
		}

		splitToken := strings.Split(headerToken, "Bearer ")
		if len(splitToken) != 2 {
			writer.Header().Set("Content-Type", "application/json")
			writer.WriteHeader(http.StatusUnauthorized)

			webResponse := web.WebResponse{
				Code:   http.StatusUnauthorized,
				Status: "Unauthorized",
				Data:   "Token format is not match",
			}

			middleware.Log.Warn().Msg("Unauthorized, token format is not match")
			helper.WriteToResponseBody(writer, webResponse)
			return
		}

		secretKey := middleware.Config.String("SECRET_KEY_ACCESS_TOKEN")
		secretKeyByte := []byte(secretKey)

		token, err := jwt.Parse(splitToken[1], func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, http.ErrNotSupported
			}
			return secretKeyByte, nil
		})

		if err != nil {
			if err == jwt.ErrTokenMalformed {
				writer.Header().Set("Content-Type", "application/json")
				writer.WriteHeader(http.StatusUnauthorized)

				webResponse := web.WebResponse{
					Code:   http.StatusUnauthorized,
					Status: "Unauthorized",
					Data:   "Token is malformed",
				}

				middleware.Log.Warn().Msg("Unauthorized, invalid token")
				helper.WriteToResponseBody(writer, webResponse)
				return
			} else if err.Error() == "token has invalid claims: token is expired" {
				writer.Header().Set("Content-Type", "application/json")
				writer.WriteHeader(http.StatusUnauthorized)

				webResponse := web.WebResponse{
					Code:   http.StatusUnauthorized,
					Status: "Unauthorized",
					Data:   "Token is expired",
				}

				middleware.Log.Warn().Msg("Unauthorized, invalid token")
				helper.WriteToResponseBody(writer, webResponse)
				return
			} else {
				writer.Header().Set("Content-Type", "application/json")
				writer.WriteHeader(http.StatusUnauthorized)

				webResponse := web.WebResponse{
					Code:   http.StatusUnauthorized,
					Status: "Unauthorized",
					Data:   "Invalid token",
				}

				middleware.Log.Warn().Msg("Unauthorized, invalid token")
				helper.WriteToResponseBody(writer, webResponse)
				return
			}
		}

		var id string
		if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
			if val, exists := claims["id"]; exists {
				if strVal, ok := val.(string); ok {
					id = strVal
				}
			} else {
				writer.Header().Set("Content-Type", "application/json")
				writer.WriteHeader(http.StatusUnauthorized)

				webResponse := web.WebResponse{
					Code:   http.StatusUnauthorized,
					Status: "Unauthorized",
					Data:   "Invalid Token",
				}

				middleware.Log.Warn().Msg("Unauthorized, invalid token")
				helper.WriteToResponseBody(writer, webResponse)
				return
			}
		}

		_, err = middleware.ProductUsecase.FindUserExistenceAPI(request.Context(), splitToken[1])
		if err != nil {
			writer.Header().Set("Content-Type", "application/json")
			writer.WriteHeader(http.StatusUnauthorized)

			webResponse := web.WebResponse{
				Code:   http.StatusUnauthorized,
				Status: "Unauthorized",
				Data:   "User not found, please register",
			}

			middleware.Log.Warn().Msg("Unauthorized, user not found")
			helper.WriteToResponseBody(writer, webResponse)
			return
		}

		middleware.Log.Debug().Msg("User:" + id)
		ctx := context.WithValue(request.Context(), userUUIDkey, id)
		ctx = context.WithValue(request.Context(), userAuthToken, splitToken[1])
		next(writer, request.WithContext(ctx), p)
	}
}
