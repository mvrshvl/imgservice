package di

import (
	"context"
	"go.uber.org/dig"
)

func BuildContainer(constructors ...interface{}) (*dig.Container, error) {
	container := dig.New()

	for _, c := range constructors {
		err := container.Provide(c)
		if err != nil {
			return nil, err
		}
	}

	return container, nil
}

type diKey struct{}

func WithContext(ctx context.Context, container *dig.Container) context.Context {
	return context.WithValue(ctx, diKey{}, container)
}

func FromContext(ctx context.Context) *dig.Container {
	return ctx.Value(diKey{}).(*dig.Container)
}
