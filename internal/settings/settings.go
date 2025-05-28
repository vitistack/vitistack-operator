package settings

import (
	"github.com/NorskHelsenett/oss-datacenter-operator/pkg/consts"
	"github.com/spf13/viper"
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
