package main

import (
	"context"
	"fmt"
	"github.com/sirupsen/logrus"
	"log"
	"nir/config"
	"nir/di"
	logging "nir/log"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	container, err := di.BuildContainer(
		config.New,
		logging.New,
	)
	if err != nil {
		log.Fatal(err)
	}

	ctx = di.WithContext(ctx, container)

	test(ctx)
}

func test(ctx context.Context) error {
	return di.FromContext(ctx).Invoke(func(l *logrus.Entry, c *config.Config) error {
		l.Error(fmt.Sprintf("config %v", c))

		return nil
	})
}
