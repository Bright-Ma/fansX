package config

import (
	"github.com/spf13/viper"
)

type Config struct {
	Websocket struct {
		WriteBufferSize int
		ReadBufferSize  int
		WriteTimeout    int64
		ReadTimeout     int64
		Port            string
	}
	Etcd struct {
		EndPoints   []string
		DialTimeout int64
		Name        string
		Addr        string
		TTL         int64
		WatchPrefix string
	}
	Redis struct {
		Addr string
		DB   int
		TTL  int64
	}
	NSQ struct {
		Topic string
		Addr  string
	}
	Model       int
	IP          string
	VirtualNums int
}

func ReadConfig(file string) Config {
	var c Config
	viper.SetConfigFile(file)
	if err := viper.ReadInConfig(); err != nil {
		panic(err.Error())
	}
	if err := viper.Unmarshal(&c); err != nil {
		panic(err.Error())
	}
	return c
}
