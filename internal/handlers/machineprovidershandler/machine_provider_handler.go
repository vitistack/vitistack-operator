package machineprovidershandler

import (
	"net/http"

	"github.com/NorskHelsenett/oss-datacenter-operator/internal/helpers/httphelpers"
	"github.com/NorskHelsenett/oss-datacenter-operator/internal/helpers/uuidhelpers"
	"github.com/NorskHelsenett/oss-datacenter-operator/internal/repositories"
	"github.com/gorilla/mux"
)

func GetMachineProviders(w http.ResponseWriter, r *http.Request) {
	machineProviders, err := repositories.MachineProviderRepository.GetAll(r.Context())
	if err != nil {
		httphelpers.RespondWithError(w, http.StatusInternalServerError, "Failed to retrieve machine providers")
		return
	}

	if len(machineProviders) == 0 {
		httphelpers.RespondWithError(w, http.StatusNotFound, "No machine providers found")
		return
	}

	if err := httphelpers.RespondWithJSON(w, http.StatusOK, machineProviders); err != nil {
		httphelpers.RespondWithError(w, http.StatusInternalServerError, "Failed to serialize machine providers")
		return
	}
}

func GetMachineProviderByUID(w http.ResponseWriter, r *http.Request) {
	// Extract URL parameters
	vars := mux.Vars(r)
	id := vars["uid"]

	// Validate the UUID format
	if !uuidhelpers.IsValidUUID(id) {
		httphelpers.RespondWithError(w, http.StatusBadRequest, "Invalid UUID format")
		return
	}

	machineProvider, err := repositories.MachineProviderRepository.GetByUID(r.Context(), id)
	if err != nil {
		httphelpers.RespondWithError(w, http.StatusInternalServerError, "Failed to retrieve machine provider")
		return
	}

	if machineProvider.Name == "" {
		httphelpers.RespondWithError(w, http.StatusNotFound, "Machine provider not found")
		return
	}

	if err := httphelpers.RespondWithJSON(w, http.StatusOK, machineProvider); err != nil {
		httphelpers.RespondWithError(w, http.StatusInternalServerError, "Failed to serialize machine provider")
		return
	}
}
