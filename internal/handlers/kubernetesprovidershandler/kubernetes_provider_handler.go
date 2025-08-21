package kubernetesprovidershandler

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/vitistack/vitistack-operator/internal/helpers/httphelpers"
	"github.com/vitistack/vitistack-operator/internal/helpers/uuidhelpers"
	"github.com/vitistack/vitistack-operator/internal/repositories"
)

func GetKubernetesProviderByUID(w http.ResponseWriter, r *http.Request) {
	// Extract URL parameters
	vars := mux.Vars(r)
	id := vars["uid"]

	if id == "" {
		httphelpers.RespondWithError(w, http.StatusBadRequest, "UUID is required")
		return
	}

	// Validate the UUID format
	if !uuidhelpers.IsValidUUID(id) {
		httphelpers.RespondWithError(w, http.StatusBadRequest, "Invalid UUID format")
		return
	}

	kp, err := repositories.KubernetesProviderRepository.GetByUID(r.Context(), id)
	if err != nil {
		httphelpers.RespondWithError(w, http.StatusInternalServerError, "Failed to retrieve Kubernetes provider")
		return
	}
	if kp.Name == "" {
		httphelpers.RespondWithError(w, http.StatusNotFound, "Kubernetes provider not found")
		return
	}
	if err := httphelpers.RespondWithJSON(w, http.StatusOK, kp); err != nil {
		httphelpers.RespondWithError(w, http.StatusInternalServerError, "Failed to serialize Kubernetes provider")
		return
	}
}

func GetKubernetesProviders(w http.ResponseWriter, r *http.Request) {
	kps, err := repositories.KubernetesProviderRepository.GetAll(r.Context())
	if err != nil {
		httphelpers.RespondWithError(w, http.StatusInternalServerError, "Failed to retrieve Kubernetes providers")
		return
	}

	if err := httphelpers.RespondWithJSON(w, http.StatusOK, kps); err != nil {
		httphelpers.RespondWithError(w, http.StatusInternalServerError, "Failed to serialize Kubernetes providers")
		return
	}
}
