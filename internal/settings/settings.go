package settings

import (
	"github.com/spf13/viper"
	"github.com/vitistack/datacenter-operator/pkg/consts"
)

var (
	Version = "0.0.0"
	Commit  = "localdev"
)

func Init() {
	// Initialize settings here
	viper.SetDefault(consts.DATACENTERCRDNAME, "datacenter")
	viper.SetDefault(consts.CONFIGMAPNAME, "datacenter-config")
	viper.SetDefault(consts.NAMESPACE, "default")
	viper.SetDefault(consts.DEVELOPMENT, false)

	viper.AutomaticEnv()
}
