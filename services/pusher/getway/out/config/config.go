package config

import (
	"github.com/spf13/viper"
)

type Config struct {
	Etcd struct {
		Endpoints   []string
		DialTimeout int64
		WatchPrefix string
	}
	Port        string
	Model       int
	VirtualNums int
}

func ReadConfig(file string) Config {
	c := Config{}
	viper.SetConfigFile(file)
	if err := viper.ReadInConfig(); err != nil {
		panic(err.Error())
	}
	if err := viper.Unmarshal(&c); err != nil {
		panic(err.Error())
	}
	return c
}
