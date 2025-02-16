package http

import (
	"github.com/julienschmidt/httprouter"
	"github.com/rs/zerolog"
	"gocdc/internal/helper"
	"gocdc/internal/model/web"
	"gocdc/internal/model/web/user"
	"gocdc/internal/usecase"
	"net/http"
)

type UserController struct {
	UserUsecase *usecase.UserUsecase
	Log         *zerolog.Logger
}

func NewUserController(userUsecase *usecase.UserUsecase, zerolog *zerolog.Logger) *UserController {
	return &UserController{
		UserUsecase: userUsecase,
		Log:         zerolog,
	}
}

func (controller UserController) Register(writer http.ResponseWriter, request *http.Request, _ httprouter.Params) {
	userRegisterRequest := user.UserRegisterRequest{}
	helper.ReadFromRequestBody(request, &userRegisterRequest)

	tokenResponse, err := controller.UserUsecase.Register(request.Context(), userRegisterRequest)
	if err != nil {
		writer.Header().Set("Content-Type", "application/json")
		writer.WriteHeader(http.StatusBadRequest)
		webResponse := web.WebResponse{
			Code:   http.StatusBadRequest,
			Status: "Bad Request",
			Data:   err.Error(),
		}
		helper.WriteToResponseBody(writer, webResponse)
		return
	}

	webResponse := web.WebResponse{
		Code:   200,
		Status: "OK",
		Data:   tokenResponse,
	}

	helper.WriteToResponseBody(writer, webResponse)
}

func (controller UserController) Login(writer http.ResponseWriter, request *http.Request, _ httprouter.Params) {
	userLoginRequest := user.UserLoginRequest{}
	helper.ReadFromRequestBody(request, &userLoginRequest)

	tokenResponse, err := controller.UserUsecase.Login(request.Context(), userLoginRequest)
	if err != nil {
		if err.Error() == "wrong email or password" {
			writer.Header().Set("Content-Type", "application/json")
			writer.WriteHeader(http.StatusNotFound)
			webResponse := web.WebResponse{
				Code:   http.StatusNotFound,
				Status: "Not Found",
				Data:   err.Error(),
			}

			helper.WriteToResponseBody(writer, webResponse)
			return
		}
		writer.Header().Set("Content-Type", "application/json")
		writer.WriteHeader(http.StatusBadRequest)
		webResponse := web.WebResponse{
			Code:   http.StatusBadRequest,
			Status: "Bad Request",
			Data:   err.Error(),
		}

		helper.WriteToResponseBody(writer, webResponse)
		return
	}

	webResponse := web.WebResponse{
		Code:   200,
		Status: "OK",
		Data:   tokenResponse,
	}

	helper.WriteToResponseBody(writer, webResponse)
}

func (controller UserController) Update(writer http.ResponseWriter, request *http.Request, _ httprouter.Params) {
	userUUID, _ := request.Context().Value("user_uuid").(string)

	userUpdateRequest := user.UserUpdateRequest{}
	helper.ReadFromRequestBody(request, &userUpdateRequest)

	err := controller.UserUsecase.Update(request.Context(), userUpdateRequest, userUUID)
	if err != nil {
		writer.Header().Set("Content-Type", "application/json")
		writer.WriteHeader(http.StatusBadRequest)
		webResponse := web.WebResponse{
			Code:   http.StatusBadRequest,
			Status: "Bad Request",
			Data:   err.Error(),
		}

		helper.WriteToResponseBody(writer, webResponse)
		return
	}

	webResponse := web.WebResponse{
		Code:   200,
		Status: "OK",
	}

	helper.WriteToResponseBody(writer, webResponse)
}

func (controller UserController) Delete(writer http.ResponseWriter, request *http.Request, _ httprouter.Params) {
	userUUID, _ := request.Context().Value("user_uuid").(string)

	err := controller.UserUsecase.Delete(request.Context(), userUUID)
	if err != nil {
		writer.Header().Set("Content-Type", "application/json")
		writer.WriteHeader(http.StatusNotFound)

		webResponse := web.WebResponse{
			Code:   http.StatusNotFound,
			Status: "Not Found",
			Data:   err.Error(),
		}

		helper.WriteToResponseBody(writer, webResponse)
		return
	}

	webResponse := web.WebResponse{
		Code:   200,
		Status: "OK",
	}

	helper.WriteToResponseBody(writer, webResponse)
}

func (controller UserController) FindUserInfo(writer http.ResponseWriter, request *http.Request, _ httprouter.Params) {
	userUUID, _ := request.Context().Value("user_uuid").(string)

	userResponse, err := controller.UserUsecase.FindUserInfo(request.Context(), userUUID)
	if err != nil {
		writer.Header().Set("Content-Type", "application/json")
		writer.WriteHeader(http.StatusNotFound)

		webResponse := web.WebResponse{
			Code:   http.StatusNotFound,
			Status: "Not Found",
			Data:   err.Error(),
		}

		helper.WriteToResponseBody(writer, webResponse)
		return
	}

	webResponse := web.WebResponse{
		Code:   200,
		Status: "OK",
		Data:   userResponse,
	}

	helper.WriteToResponseBody(writer, webResponse)
}

func (controller UserController) CheckUserExistence(writer http.ResponseWriter, request *http.Request, _ httprouter.Params) {
	userUUID, _ := request.Context().Value("user_uuid").(string)

	userStatusResponse, err := controller.UserUsecase.CheckUserExistence(request.Context(), userUUID)
	if err != nil {
		writer.Header().Set("Content-Type", "application/json")
		writer.WriteHeader(http.StatusNotFound)

		webResponse := web.WebResponse{
			Code:   http.StatusNotFound,
			Status: "Not Found",
			Data:   err.Error(),
		}

		helper.WriteToResponseBody(writer, webResponse)
		return
	}

	webResponse := web.WebResponse{
		Code:   200,
		Status: "OK",
		Data:   userStatusResponse,
	}

	helper.WriteToResponseBody(writer, webResponse)
}

func (controller UserController) FindUserNameAddress(writer http.ResponseWriter, request *http.Request, _ httprouter.Params) {
	userUUID, _ := request.Context().Value("user_uuid").(string)

	userResponse, err := controller.UserUsecase.FindUserNameAddress(request.Context(), userUUID)
	if err != nil {
		writer.Header().Set("Content-Type", "application/json")
		writer.WriteHeader(http.StatusNotFound)

		webResponse := web.WebResponse{
			Code:   http.StatusNotFound,
			Status: "Not Found",
			Data:   err.Error(),
		}

		helper.WriteToResponseBody(writer, webResponse)
		return
	}

	webResponse := web.WebResponse{
		Code:   200,
		Status: "OK",
		Data:   userResponse,
	}

	helper.WriteToResponseBody(writer, webResponse)
}

func (controller UserController) FindUserEmail(writer http.ResponseWriter, request *http.Request, _ httprouter.Params) {
	userUUID, _ := request.Context().Value("user_uuid").(string)

	userResponse, err := controller.UserUsecase.FindUserEmail(request.Context(), userUUID)
	if err != nil {
		writer.Header().Set("Content-Type", "application/json")
		writer.WriteHeader(http.StatusNotFound)

		webResponse := web.WebResponse{
			Code:   http.StatusNotFound,
			Status: "Not Found",
			Data:   err.Error(),
		}

		helper.WriteToResponseBody(writer, webResponse)
		return
	}

	userEmailRespone := user.UserEmailResponse{
		Email: userResponse,
	}

	webResponse := web.WebResponse{
		Code:   200,
		Status: "OK",
		Data:   userEmailRespone,
	}

	helper.WriteToResponseBody(writer, webResponse)
}

func (controller UserController) TokenRenewal(writer http.ResponseWriter, request *http.Request, _ httprouter.Params) {
	userRenewalTokenRequest := user.RenewalTokenRequest{}
	helper.ReadFromRequestBody(request, &userRenewalTokenRequest)

	tokenResponse, err := controller.UserUsecase.TokenRenewal(request.Context(), userRenewalTokenRequest)
	if err != nil {
		if err.Error() == "User not found" {
			writer.Header().Set("Content-Type", "application/json")
			writer.WriteHeader(http.StatusNotFound)

			webResponse := web.WebResponse{
				Code:   http.StatusNotFound,
				Status: "Not Found",
				Data:   err.Error(),
			}

			helper.WriteToResponseBody(writer, webResponse)
			return
		}
		if err.Error() == "invalid request body" || err.Error() == "Token is malformed" || err.Error() == "Invalid token" {
			writer.Header().Set("Content-Type", "application/json")
			writer.WriteHeader(http.StatusBadRequest)

			webResponse := web.WebResponse{
				Code:   http.StatusBadRequest,
				Status: "Bad Request",
				Data:   err.Error(),
			}

			helper.WriteToResponseBody(writer, webResponse)
			return
		}
		if err.Error() == "Refresh token reuse detected. For security reasons, you have been logged out. Please sign in again." {
			writer.Header().Set("Content-Type", "application/json")
			writer.WriteHeader(http.StatusForbidden)

			webResponse := web.WebResponse{
				Code:   http.StatusForbidden,
				Status: "Forbidden",
				Data:   err.Error(),
			}

			helper.WriteToResponseBody(writer, webResponse)
			return
		}
	}

	webResponse := web.WebResponse{
		Code:   200,
		Status: "OK",
		Data:   tokenResponse,
	}

	helper.WriteToResponseBody(writer, webResponse)
}
