package log

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"nir/config"
	"os"
	"path/filepath"
	"time"
)

func New(cfg *config.Config) (*logrus.Entry, error) {
	logPath := filepath.Join(".", "logs")

	err := os.MkdirAll(logPath, os.ModePerm)
	if err != nil {
		return nil, err
	}

	f, err := os.Create(filepath.Join(logPath, fmt.Sprintf("%s.log", time.Now().String())))
	if err != nil {
		return nil, err
	}

	logger := logrus.New()

	logger.SetOutput(f)
	logger.SetLevel(cfg.Log.Level)

	return logrus.NewEntry(logger), nil
}
