package initialization

import (
	"fmt"
	"os"
	"strings"

	"github.com/pkg/errors"
	"github.com/spf13/viper"
	"golang.org/x/exp/slog"
)

func Viper() {

	viper.SetDefault("StaticDir", "./static")
	viper.SetDefault("TemplateDir", "./template")

	viper.SetConfigName("config")
	viper.SetConfigType("json")

	// possible locations for config file
	viper.AddConfigPath("./configs")

	viper.SetEnvPrefix("GOIABADA")
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			// Config file was found but another error was produced
			slog.Error(errors.Wrap(err, "unable to initialize configuration - make sure a config.json file exists and has content").Error())
			os.Exit(1)
		}
	}

	slog.Info("viper configuration initialized")
	if len(viper.ConfigFileUsed()) > 0 {
		slog.Info(fmt.Sprintf("viper config file used: %v", viper.ConfigFileUsed()))
	}
}
