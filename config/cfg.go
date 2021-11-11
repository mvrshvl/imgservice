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

const ConfigPath = "./config"

type Config struct {
	Blockchain
	Clustering
	Log
	Output
}

type Blockchain struct {
	BlocksTable       string `default:"blockchain_data/blocks.csv"`
	TransactionsTable string `default:"blockchain_data/transactions.csv"`
	ExchangesTable    string `default:"blockchain_data/exchanges.csv"`
}

type Clustering struct {
	MaxBlockDiff uint64  `default:"10000"`
	MaxETHDiff   float64 `default:"0.01"`
}

type Log struct {
	Level logrus.Level
}

type Output struct {
	GraphDepositsReuse string `default:"output/deposit_reuse.html"`
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
