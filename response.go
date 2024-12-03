package simba

import (
	"encoding/json"
	"net/http"
)

// writeResponse writes the response to the client
func writeResponse(w http.ResponseWriter, r *http.Request, resp *Response, err error) {
	if err != nil {
		handleError(w, r, err)
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

	if resp.Body == nil {
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
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		return
	}

	err = writeJSON(w, http.StatusOK, resp.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

// writeJSON is a helper function for writing JSON responses
func writeJSON(w http.ResponseWriter, status int, v any) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	return json.NewEncoder(w).Encode(v)
}
