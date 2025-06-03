package routes

import (
	"github.com/NorskHelsenett/oss-datacenter-operator/internal/handlers/datacenterhandler"
	"github.com/NorskHelsenett/oss-datacenter-operator/internal/handlers/healthhandler"
	"github.com/NorskHelsenett/oss-datacenter-operator/internal/handlers/kubernetesprovidershandler"
	"github.com/NorskHelsenett/oss-datacenter-operator/internal/handlers/machineprovidershandler"
	"github.com/NorskHelsenett/oss-datacenter-operator/internal/handlers/versionhandler"
	"github.com/NorskHelsenett/oss-datacenter-operator/internal/middlewares"
	"github.com/gorilla/mux"
)

func SetupRoutes(r *mux.Router) {
	r.Use(middlewares.LoggingMiddleware)
	r.Use(middlewares.ContentTypeMiddleware) // Add ContentTypeMiddleware for all routes

	r.HandleFunc("/health", healthhandler.HealthCheck).Methods("GET")
	r.HandleFunc("/v1/info/version", versionhandler.GetVersion).Methods("GET")

	v1route := r.NewRoute().Subrouter().PathPrefix("/v1").Subrouter()
	v1route.Use(middlewares.AuthMiddleware)
	v1route.HandleFunc("/datacenter/name", datacenterhandler.GetName).Methods("GET")

	v1route.HandleFunc("/machineproviders", machineprovidershandler.GetMachineProviders).Methods("GET")
	v1route.HandleFunc("/machineproviders/{uid}", machineprovidershandler.GetMachineProviderByUID).Methods("GET")

	v1route.HandleFunc("/kubernetesproviders", kubernetesprovidershandler.GetKubernetesProviders).Methods("GET")
	v1route.HandleFunc("/kubernetesproviders/{uid}", kubernetesprovidershandler.GetKubernetesProviderByUID).Methods("GET")
}
