package handlers

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/sillen102/simba"
)

func decodeBodyIfNeeded[T any](r *http.Request, req *T) (*T, error) {
	// Check if the type is NoBody
	if _, isNoBody := any(*req).(simba.NoBody); isNoBody {
		return nil, nil
	}

	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(req); err != nil {
		return nil, errors.New("invalid request body")
	}

	return req, nil
}

// readJson reads the JSON body and unmarshalls it into the model.
func readJson(body io.ReadCloser, model any) error {
	decoder := json.NewDecoder(body)
	decoder.DisallowUnknownFields()
	err := decoder.Decode(&model)
	if err != nil {
		return errors.New("invalid request body")
	}
	return nil
}
