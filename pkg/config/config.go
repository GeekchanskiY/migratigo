// Copyright 2025 GeekchanskiY
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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
