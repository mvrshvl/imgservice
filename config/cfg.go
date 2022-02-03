package config

import (
	"fmt"
	"github.com/mcuadros/go-defaults"
	"github.com/mitchellh/mapstructure"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"path"
	"reflect"
	"time"
)

const ConfigPath = "./test-config"

type Config struct {
	Clustering
	Log
	Output
	Database
	Ethereum
}

type Clustering struct {
	BatchBlocksSize uint64 `default:"false"`
}

type Log struct {
	Level logrus.Level
}

type Output struct {
	ShowSingleAccount  bool   `default:"false"`
	GraphDepositsReuse string `default:"output/cluster.html"`
}

type Ethereum struct {
	Address string `default:"false"`
}

func New() (*Config, error) {
	cfg := new(Config)
	defaults.SetDefaults(cfg)

	cfgViper := viper.New()

	cpath, cname := path.Split(ConfigPath)

	cfgViper.SetConfigName(cname)
	cfgViper.AddConfigPath(cpath)

	if err := cfgViper.ReadInConfig(); err != nil {
		return nil, err
	}

	err := cfgViper.Unmarshal(&cfg, viper.DecodeHook(
		mapstructure.ComposeDecodeHookFunc(
			duration,
			logrusLevel,
		),
	))
	if err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	return cfg, nil
}

func logrusLevel(_ reflect.Type, t reflect.Type, data interface{}) (interface{}, error) {
	if t.String() != "logrus.Level" {
		return data, nil
	}

	return logrus.ParseLevel(data.(string))
}

func duration(_ reflect.Type, t reflect.Type, data interface{}) (interface{}, error) {
	if t.String() != "time.Duration" {
		return data, nil
	}

	return time.ParseDuration(data.(string))
}

type Database struct {
	Address  string `default:"127.0.0.1:3307"`
	User     string `default:"admin"`
	Password string `default:"admin"`
	Name     string `default:"test"`
	Driver   string `default:"mysql"`
	Clean    bool   `default:"true"`
}
