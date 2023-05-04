package configs

import (
	"github.com/spf13/viper"
)

type Configs struct {
	URLs              []string
	Interval          int
	DropboxKoboFolder string
}

func Load(filename, filepath string, interval int) (Configs, error) {
	var configs Configs

	viper.SetConfigName(filename)
	viper.AddConfigPath(filepath)

	if err := viper.ReadInConfig(); err != nil {
		return configs, err
	}

	viper.SetDefault("urls", []string{})
	viper.SetDefault("interval", interval)
	viper.SetDefault("dropboxKoboFolder", "/")

	err := viper.Unmarshal(&configs)
	if err != nil {
		return configs, err
	}

	return configs, nil
}
