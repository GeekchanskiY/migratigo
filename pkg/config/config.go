package config

import (
	"github.com/spf13/viper"
)

type Config struct {
	DbURL string `mapstructure:"db_url"`
}

func New() (*Config, error) {
	viper.AddConfigPath("./migratigo")
	viper.SetConfigName("config")
	err := viper.ReadInConfig()
	if err != nil {
		return nil, err
	}

	cfg := new(Config)
	err = viper.UnmarshalKey("migratigo", cfg)

	return cfg, err
}
