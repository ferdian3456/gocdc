package helper

import (
	"github.com/bytedance/sonic"
	"io"
	"net/http"
)

func ReadFromRequestBody(request *http.Request, result interface{}) {
	body, err := io.ReadAll(request.Body)
	if err != nil {
		PanicIfError(err)
	}

	err = sonic.Unmarshal(body, result)
	PanicIfError(err)
}

func WriteToResponseBody(writer http.ResponseWriter, response interface{}) {
	writer.Header().Set("Content-Type", "application/json")

	jsonData, err := sonic.Marshal(response)
	PanicIfError(err)

	_, err = writer.Write(jsonData)
	PanicIfError(err)
}
