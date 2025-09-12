package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/vitistack/common/pkg/clients/k8sclient"
	"github.com/vitistack/common/pkg/loggers/vlog"
	"github.com/vitistack/vitistack-operator/internal/cache"
	"github.com/vitistack/vitistack-operator/internal/clients/dynamicclienthandler"
	"github.com/vitistack/vitistack-operator/internal/httpserver"
	"github.com/vitistack/vitistack-operator/internal/listeners/resourceloglistener"
	"github.com/vitistack/vitistack-operator/internal/listeners/resourcewriterlistener"
	"github.com/vitistack/vitistack-operator/internal/repositories"
	"github.com/vitistack/vitistack-operator/internal/services/dynamichandler"
	"github.com/vitistack/vitistack-operator/internal/services/initializeservice"
	"github.com/vitistack/vitistack-operator/internal/settings"
	"go.uber.org/automaxprocs/maxprocs"
)

// main is the entrypoint for the vitistack-operator binary.
func main() {
	_ = vlog.Setup(vlog.Options{Level: "info", ColorizeLine: true, AddCaller: true})
	defer func() {
		_ = vlog.Sync()
	}()

	_, _ = maxprocs.Set(maxprocs.Logger(vlog.Logr().Info))
	cancelChan := make(chan os.Signal, 1)

	stop := make(chan struct{})
	// catch SIGTERM or SIGINT.
	signal.Notify(cancelChan, syscall.SIGTERM, syscall.SIGINT)
	var err error

	settings.Init()
	k8sclient.Init()
	initializeservice.CheckPrerequisites()
	cache.Cache, err = cache.VitistackCache{}.NewVitistackCache()
	if err != nil {
		panic(err)
	}

	repositories.InitializeRepositories()
	resourceloglistener.RegisterListeners()
	resourcewriterlistener.RegisterWriters()

	go func() {
		httpserver.Start()
		sig := <-cancelChan
		_, _ = fmt.Println()
		_, _ = fmt.Println(sig)
		stop <- struct{}{}
	}()

	resourcehandler := dynamichandler.NewDynamicClientHandler()
	err = dynamicclienthandler.Start(k8sclient.DiscoveryClient, k8sclient.DynamicClient, resourcehandler, stop, cancelChan)
	if err != nil {
		vlog.Fatal("could not start dynamic client", err)
	}

	sig := <-cancelChan
	vlog.Info("Caught signal", "signal", sig)
}
