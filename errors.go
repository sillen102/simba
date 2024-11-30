package simba

import "net/http"

// HandleError is a helper function for handling errors in HTTP handlers
func HandleError(w http.ResponseWriter, r *http.Request, err error) {
	http.Error(w, err.Error(), http.StatusInternalServerError)
}
