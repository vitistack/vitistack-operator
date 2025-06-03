package versionhandler

import (
	"net/http"

	"github.com/vitistack/datacenter-operator/internal/helpers/httphelpers"
	"github.com/vitistack/datacenter-operator/internal/settings"
)

func GetVersion(w http.ResponseWriter, r *http.Request) {
	err := httphelpers.RespondWithJSON(w, http.StatusOK, map[string]string{"version": settings.Version, "commit": settings.Commit})
	if err != nil {
		httphelpers.RespondWithError(w, http.StatusInternalServerError, "Something went wrong")
		return
	}
}
