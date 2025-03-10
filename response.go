package simba

import (
	"encoding/json"
	"net/http"

	"github.com/sillen102/simba/logging"
	"github.com/sillen102/simba/simbaErrors"
	"github.com/sillen102/simba/simbaModels"
)

// TODO: Response testing
//  1. Error response formatting
//  2. Response headers and cookies
//  3. Response compression
//  4. Response specific test cases (such as 204 when body is nil and status is 0)

// writeResponse writes the response to the client
func writeResponse[ResponseBody any](w http.ResponseWriter, r *http.Request, resp *simbaModels.Response[ResponseBody], err error) {
	if err != nil {
		simbaErrors.WriteError(w, r, err)
		return
	}

	// Check if resp is nil
	if resp == nil {
		// Log this unexpected condition
		logging.From(r.Context()).Error("unexpected nil response")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if resp.Headers != nil {
		for key, value := range resp.Headers {
			for _, v := range value {
				w.Header().Add(key, v)
			}
		}
	}

	if resp.Cookies != nil {
		for _, cookie := range resp.Cookies {
			http.SetCookie(w, cookie)
		}
	}

	var status int
	if resp.Status != 0 {
		status = resp.Status
	} else if any(resp.Body) == (simbaModels.NoBody{}) {
		status = http.StatusNoContent
	} else {
		status = http.StatusOK
	}

	err = writeJSON(w, status, resp.Body)
	if err != nil {
		simbaErrors.HandleUnexpectedError(w)
		return
	}
}

// writeJSON is a helper function for writing JSON responses
func writeJSON(w http.ResponseWriter, status int, v any) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	return json.NewEncoder(w).Encode(v)
}
