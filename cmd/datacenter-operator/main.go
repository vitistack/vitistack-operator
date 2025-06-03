package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/NorskHelsenett/ror/pkg/rlog"
	"github.com/vitistack/datacenter-operator/internal/cache"
	"github.com/vitistack/datacenter-operator/internal/clients"
	"github.com/vitistack/datacenter-operator/internal/clients/dynamicclienthandler"
	"github.com/vitistack/datacenter-operator/internal/httpserver"
	"github.com/vitistack/datacenter-operator/internal/listeners/resourceloglistener"
	"github.com/vitistack/datacenter-operator/internal/listeners/resourcewriterlistener"
	"github.com/vitistack/datacenter-operator/internal/repositories"
	"github.com/vitistack/datacenter-operator/internal/services/dynamichandler"
	"github.com/vitistack/datacenter-operator/internal/services/initializeservice"
	"github.com/vitistack/datacenter-operator/internal/settings"
	"go.uber.org/automaxprocs/maxprocs"
)

func main() {
	_, _ = maxprocs.Set(maxprocs.Logger(rlog.Infof))
	cancelChan := make(chan os.Signal, 1)

	stop := make(chan struct{})
	// catch SIGETRM or SIGINTERRUPT.
	signal.Notify(cancelChan, syscall.SIGTERM, syscall.SIGINT)
	var err error

	settings.Init()
	clients.Init()
	initializeservice.CheckPrerequisites()
	cache.Cache, err = cache.DatacenterCache{}.NewDatacenterCache()
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
	err = dynamicclienthandler.Start(clients.DiscoveryClient, clients.DynamicClient, resourcehandler, stop, cancelChan)
	if err != nil {
		rlog.Fatal("could not start dynamic client", err)
	}

	sig := <-cancelChan
	rlog.Info("Caught signal", rlog.Any("signal", sig))
}
