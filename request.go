package simba

import (
	"encoding/json"
	"io"
	"net/http"
)

// decodeBodyIfNeeded decodes the request body if it is not of NoBody type
func decodeBodyIfNeeded[T any](r *http.Request, req *T) error {
	if _, isNoBody := any(*req).(NoBody); isNoBody {
		return nil
	}

	if r.Header.Get("Content-Type") != "application/json" {
		return NewHttpError(http.StatusBadRequest, "invalid content type", nil)
	}

	return readJson(r.Body, req)
}

// readJson reads the JSON body and unmarshalls it into the model
func readJson(body io.ReadCloser, model any) error {
	decoder := json.NewDecoder(body)
	if options.RequestDisallowUnknownFields {
		decoder.DisallowUnknownFields()
	}
	err := decoder.Decode(&model)
	if err != nil {
		return NewHttpError(http.StatusBadRequest, "invalid request body", err)
	}
	return nil
}
