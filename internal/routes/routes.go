package routes

import (
	"github.com/gorilla/mux"
	"github.com/vitistack/vitistack-operator/internal/handlers/healthhandler"
	"github.com/vitistack/vitistack-operator/internal/handlers/kubernetesprovidershandler"
	"github.com/vitistack/vitistack-operator/internal/handlers/machineprovidershandler"
	"github.com/vitistack/vitistack-operator/internal/handlers/versionhandler"
	"github.com/vitistack/vitistack-operator/internal/handlers/vitistackhandler"
	"github.com/vitistack/vitistack-operator/internal/middlewares"
)

func SetupRoutes(r *mux.Router) {
	r.Use(middlewares.LoggingMiddleware)
	r.Use(middlewares.ContentTypeMiddleware) // Add ContentTypeMiddleware for all routes

	r.HandleFunc("/health", healthhandler.HealthCheck).Methods("GET")
	r.HandleFunc("/v1/info/version", versionhandler.GetVersion).Methods("GET")

	v1route := r.NewRoute().Subrouter().PathPrefix("/v1").Subrouter()
	v1route.Use(middlewares.AuthMiddleware)
	v1route.HandleFunc("/vitistack/name", vitistackhandler.GetName).Methods("GET")

	v1route.HandleFunc("/machineproviders", machineprovidershandler.GetMachineProviders).Methods("GET")
	v1route.HandleFunc("/machineproviders/{uid}", machineprovidershandler.GetMachineProviderByUID).Methods("GET")

	v1route.HandleFunc("/kubernetesproviders", kubernetesprovidershandler.GetKubernetesProviders).Methods("GET")
	v1route.HandleFunc("/kubernetesproviders/{uid}", kubernetesprovidershandler.GetKubernetesProviderByUID).Methods("GET")
}
