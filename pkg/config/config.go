package config

import "github.com/spf13/viper"

type Config struct {
	DbURL string `yaml:"db_url"`
}

func New() (*Config, error) {
	viper.SetConfigName("config")
	err := viper.ReadInConfig()
	if err != nil {
		return nil, err
	}

	return nil, nil
}
