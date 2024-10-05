package utils

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/cisco-eti/sre-go-helloworld/pkg/models"
)

// OKResponse writes back data as json with a status of 200
func OKResponse(w http.ResponseWriter, data interface{}) error {
	resp := models.APIResponse{
		StatusCode: http.StatusOK,
		Data:       data,
	}
	return writeResponse(w, resp)
}

// CreatedResponse writes back data as json with a status of 201
func CreatedResponse(w http.ResponseWriter, data interface{}) error {
	resp := models.APIResponse{
		StatusCode: http.StatusCreated,
		Data:       data,
	}
	return writeResponse(w, resp)
}

// BadRequestResponse writes back errors as json with a status of 400
func BadRequestResponse(w http.ResponseWriter, errMsgs ...string) error {
	resp := models.APIResponse{
		StatusCode: http.StatusBadRequest,
	}
	for _, errMsg := range errMsgs {
		resp.Errors = append(resp.Errors, models.Error{Error: errMsg})
	}
	return writeResponse(w, resp)
}

// UnauthorizedResponse writes back errors as json with a status of 401
func UnauthorizedResponse(w http.ResponseWriter, errMsgs ...string) error {
	resp := models.APIResponse{
		StatusCode: http.StatusUnauthorized,
	}
	for _, errMsg := range errMsgs {
		resp.Errors = append(resp.Errors, models.Error{Error: errMsg})
	}
	return writeResponse(w, resp)
}

// NotFoundResponse writes back errors as json with a status of 404
func NotFoundResponse(w http.ResponseWriter, errMsgs ...string) error {
	resp := models.APIResponse{
		StatusCode: http.StatusNotFound,
	}
	for _, errMsg := range errMsgs {
		resp.Errors = append(resp.Errors, models.Error{Error: errMsg})
	}
	return writeResponse(w, resp)
}

// ServerErrorResponse writes back errors as json with a status of 500
func ServerErrorResponse(w http.ResponseWriter, errMsgs ...string) error {
	resp := models.APIResponse{
		StatusCode: http.StatusInternalServerError,
	}
	for _, errMsg := range errMsgs {
		resp.Errors = append(resp.Errors, models.Error{Error: errMsg})
	}
	return writeResponse(w, resp)
}

func writeResponse(w http.ResponseWriter, data models.APIResponse) error {
	if data.StatusCode == 0 {
		if len(data.Errors) == 0 {
			data.StatusCode = http.StatusOK
		} else {
			data.StatusCode = http.StatusInternalServerError
		}
	}
	if data.Description == "" {
		data.Description = http.StatusText(data.StatusCode)
	}

	j, err := json.Marshal(data)
	if err != nil {
		//log.Warning("marshaling response to json: %s", err)
		return writeRawResponse(w, http.StatusInternalServerError,
			[]byte(`{"message":"unexpected server error"}`))
	}

	return writeRawResponse(w, data.StatusCode, j)
}

func writeRawResponse(w http.ResponseWriter, status int, data []byte) error {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Content-Length", strconv.Itoa(len(data)))
	w.WriteHeader(status)
	_, err := w.Write(data)
	return err
}
