package config

import (
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

var defaultEnvVariables = map[string]interface{}{
	"config.port":          8080,
	"config.logging.level": "Info",
	"psql.host":            "localhost",
	"psql.port":            5432,
	"psql.username":        "postgres",
	"psql.password":        "password",
	"psql.db.name":         "postgres",
	"psql.timezone":        "Europe/Warsaw",
}

func Load() {
	logrus.Info("Loading configuration")
	configureDefaults()
	useEnvironment()
}

func GetString(key string) string {
	return viper.GetString(key)
}

func GetInt(key string) int {
	return viper.GetInt(key)
}

func GetBool(key string) bool {
	return viper.GetBool(key)
}

func configureDefaults() {
	logrus.Info("Setting default environment variables")
	for key, value := range defaultEnvVariables {
		viper.SetDefault(key, value)
	}
}

func useEnvironment() {
	viper.AutomaticEnv()

	r := strings.NewReplacer(".", "_")
	viper.SetEnvKeyReplacer(r)
}
