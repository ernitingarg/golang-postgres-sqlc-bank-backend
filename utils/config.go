package utils

import (
	"github.com/spf13/viper"
)

// Config stores all configuration of the application
// The values are read using viper from .env file or environment variables
type Config struct {
	DbHost        string `mapstructure:"DATABASE_HOST"`
	DbName        string `mapstructure:"DATABASE_NAME"`
	DbUser        string `mapstructure:"DATABASE_USER"`
	DbPassword    string `mapstructure:"DATABASE_PASSWORD"`
	DbPort        int32  `mapstructure:"DATABASE_PORT"`
	DbUrl         string `mapstructure:"DATABASE_URL"`
	ServerAddress string `mapstructure:"SERVER_ADDRESS"`
}

// LoadConfig loads configuration from .env file and environment variables
// and return Config object
func LoadConfig(path string) (config Config, err error) {
	viper.SetConfigName("app")
	viper.SetConfigType("env")
	viper.AddConfigPath(path)
	viper.AutomaticEnv()
	err = viper.ReadInConfig()
	if err != nil {
		return
	}

	err = viper.Unmarshal(&config)
	return
}
