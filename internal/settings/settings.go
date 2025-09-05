package settings

import (
	"github.com/spf13/viper"
	"github.com/vitistack/vitistack-operator/pkg/consts"
)

var (
	Version = "0.0.0"
	Commit  = "localdev"
)

func Init() {
	// Initialize settings here
	viper.SetDefault(consts.VITISTACKCRDNAME, "vitistack")
	viper.SetDefault(consts.CONFIGMAPNAME, "vitistack-config")
	viper.SetDefault(consts.NAMESPACE, "default")
	viper.SetDefault(consts.DEVELOPMENT, false)
	viper.SetDefault(consts.REGION, "Norway")
	viper.SetDefault(consts.LOCATION, "Trøndelag")
	viper.SetDefault(consts.JSON_LOGGING, true)

	viper.AutomaticEnv()
}
