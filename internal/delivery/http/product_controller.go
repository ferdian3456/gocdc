package http

import (
	"errors"
	"github.com/julienschmidt/httprouter"
	"github.com/rs/zerolog"
	"gocdc/internal/helper"
	"gocdc/internal/model/web"
	"gocdc/internal/model/web/product"
	"gocdc/internal/usecase"
	"net/http"
	"strconv"
)

type ProductController struct {
	ProductUsecase *usecase.ProductUsecase
	Log            *zerolog.Logger
}

func NewProductController(productUsecase *usecase.ProductUsecase, zerolog *zerolog.Logger) *ProductController {
	return &ProductController{
		ProductUsecase: productUsecase,
		Log:            zerolog,
	}
}

func (controller ProductController) Create(writer http.ResponseWriter, request *http.Request, _ httprouter.Params) {
	userUUID, _ := request.Context().Value("user_uuid").(string)

	productCreateRequest := product.ProductCreateRequest{}
	helper.ReadFromRequestBody(request, &productCreateRequest)

	err := controller.ProductUsecase.Create(request.Context(), productCreateRequest, userUUID)
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

func (controller ProductController) Update(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
	userUUID, _ := request.Context().Value("user_uuid").(string)

	productID := params.ByName("productID")
	fixProductID, err := strconv.Atoi(productID)
	if err != nil {
		respErr := errors.New("error converting string to int")
		controller.Log.Panic().Err(err).Msg(respErr.Error())
	}

	productCreateRequest := product.ProductUpdateRequest{}
	helper.ReadFromRequestBody(request, &productCreateRequest)

	// kalau productid sama userid gk sama atau bukan ownernya???

	err = controller.ProductUsecase.Update(request.Context(), productCreateRequest, userUUID, fixProductID)
	if err != nil {
		if err.Error() == "product not found" {
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
	}

	helper.WriteToResponseBody(writer, webResponse)
}

func (controller ProductController) Delete(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
	userUUID, _ := request.Context().Value("user_uuid").(string)

	productID := params.ByName("productID")
	fixProductID, err := strconv.Atoi(productID)
	if err != nil {
		respErr := errors.New("error converting string to int")
		controller.Log.Panic().Err(err).Msg(respErr.Error())
	}

	err = controller.ProductUsecase.Delete(request.Context(), userUUID, fixProductID)
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

func (controller ProductController) FindProductInfo(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
	productID := params.ByName("productID")
	fixProductID, err := strconv.Atoi(productID)
	if err != nil {
		respErr := errors.New("error converting string to int")
		controller.Log.Panic().Err(err).Msg(respErr.Error())
	}

	productResponse, err := controller.ProductUsecase.FindProductInfo(request.Context(), fixProductID)
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
		Data:   productResponse,
	}

	helper.WriteToResponseBody(writer, webResponse)

}
