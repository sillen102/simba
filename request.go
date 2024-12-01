package simba

import (
	"encoding/json"
	"errors"
	"io"

	"github.com/uptrace/bunrouter"
)

// decodeBodyIfNeeded decodes the request body if it is not of NoBody type
func decodeBodyIfNeeded[T any](r bunrouter.Request, req *T) error {
	if _, isNoBody := any(*req).(NoBody); isNoBody {
		return nil
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
		return errors.New("invalid request body")
	}
	return nil
}
