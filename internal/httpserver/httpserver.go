package httpserver

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/spf13/viper"
	"github.com/vitistack/common/pkg/loggers/vlog"
	"github.com/vitistack/vitistack-operator/internal/routes"
	"github.com/vitistack/vitistack-operator/pkg/consts"
)

func Start() {
	router := mux.NewRouter()
	routes.SetupRoutes(router)

	host := "localhost"
	if viper.GetBool(consts.DEVELOPMENT) {
		vlog.Info("Running in development mode")
	} else {
		vlog.Info("Running in production mode")
		host = ""
	}

	port := "9991"
	url := fmt.Sprintf("%s:%s", host, port)
	server := &http.Server{
		Handler:      router,
		Addr:         url,
		ReadTimeout:  20 * time.Millisecond,
		WriteTimeout: 20 * time.Millisecond,
	}
	vlog.Info(fmt.Sprintf("Starting server on port localhost:%s", port))
	vlog.Fatal("Http server stopped", server.ListenAndServe())
}
