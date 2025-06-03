package healthhandler

import (
	"net/http"

	"github.com/vitistack/datacenter-operator/internal/helpers/httphelpers"
)

func HealthCheck(w http.ResponseWriter, r *http.Request) {
	err := httphelpers.RespondWithJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	if err != nil {
		httphelpers.RespondWithError(w, http.StatusInternalServerError, "Something went wrong")
		return
	}
}
