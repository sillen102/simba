package simba

import (
	"net/http"

	"github.com/uptrace/bunrouter"
)

// HandleError is a helper function for handling errors in HTTP handlers
func HandleError(w http.ResponseWriter, r bunrouter.Request, err error) {
	http.Error(w, err.Error(), http.StatusInternalServerError)
}
