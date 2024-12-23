package simba

import (
	"encoding/json"
	"net/http"

	"github.com/sillen102/simba/logging"
)

// TODO: Response testing
//  1. Error response formatting
//  2. Response headers and cookies
//  3. Response compression
//  4. Response specific test cases (such as 204 when body is nil and status is 0)

// writeResponse writes the response to the client
func writeResponse(w http.ResponseWriter, r *http.Request, resp *Response, err error) {
	if err != nil {
		HandleError(w, r, err)
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

	if resp.Body == nil && resp.Status == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	if _, ok := resp.Body.(NoBody); ok {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	if resp.Status != 0 {
		err = writeJSON(w, resp.Status, resp.Body)
		if err != nil {
			handleUnexpectedError(w)
			return
		}
		return
	}

	err = writeJSON(w, http.StatusOK, resp.Body)
	if err != nil {
		handleUnexpectedError(w)
		return
	}
}

// writeJSON is a helper function for writing JSON responses
func writeJSON(w http.ResponseWriter, status int, v any) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	return json.NewEncoder(w).Encode(v)
}
