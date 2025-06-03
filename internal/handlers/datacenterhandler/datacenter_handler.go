package datacenterhandler

import (
	"net/http"

	"github.com/vitistack/datacenter-operator/internal/helpers/httphelpers"
	nameservice "github.com/vitistack/datacenter-operator/internal/services/datacenternameservice"
)

func GetName(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	name, err := nameservice.GetName(ctx)
	if err != nil {
		http.Error(w, "Failed to get name from configmap", http.StatusInternalServerError)
		return
	}

	// Check if the name is valid
	if name == "" {
		http.Error(w, "Name is empty", http.StatusBadRequest)
		return
	}

	err = httphelpers.RespondWithJSON(w, http.StatusOK, map[string]string{"name": name})
	if err != nil {
		http.Error(w, "Failed to respond with JSON", http.StatusInternalServerError)
		return
	}
}
