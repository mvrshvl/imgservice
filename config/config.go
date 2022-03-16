package config

import (
	"fmt"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
	"github.com/sirupsen/logrus"
)

const (
	ConfigPath = "./config.yml"
)

type Config struct {
	Server Server `yaml:"server"`
}

type Server struct {
	Address  string        `yaml:"address" env:"SERVER_ADDRESS" env-default:":6060"`
	LogLevel logrus.Level  `yaml:"level" env:"SERVER_LOG_LEVEL" env-default:"4"`
	Timeout  time.Duration `yaml:"timeout" env:"SERVER_TIMEOUT" env-default:"1m"`
	LogPath  string        `yaml:"log_path" env:"SERVER_LOG_PATH" env-default:"./logs"`
}

func New(configPath string) (*Config, error) {
	cfg := new(Config)

	err := cleanenv.ReadConfig(configPath, cfg)
	if err != nil {
		return nil, fmt.Errorf("can't read config file: %w", err)
	}

	return cfg, nil
}

func (c Config) GetServer() Server {
	return c.Server
}
