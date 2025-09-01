package httpserver

import (
	"fmt"
	"net/http"
	"time"

	"github.com/NorskHelsenett/ror/pkg/rlog"
	"github.com/gorilla/mux"
	"github.com/spf13/viper"
	"github.com/vitistack/vitistack-operator/internal/routes"
	"github.com/vitistack/vitistack-operator/pkg/consts"
)

func Start() {
	router := mux.NewRouter()
	routes.SetupRoutes(router)

	host := "localhost"
	if viper.GetBool(consts.DEVELOPMENT) {
		rlog.Info("Running in development mode")
	} else {
		rlog.Info("Running in production mode")
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
	rlog.Info(fmt.Sprintf("Starting server on port localhost:%s", port))
	rlog.Fatal("Http server stopped", server.ListenAndServe())
}
