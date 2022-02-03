package server

import (
	"context"
	"github.com/gin-gonic/gin"
	"nir/linking"
	logging "nir/log"
)

type Controller struct {
}

func NewController() *Controller {
	return &Controller{}
}

func (c *Controller) GroupWithCtx(ctx context.Context) func(router *gin.Engine) {
	return func(router *gin.Engine) {
		group := router.Group("/")
		{
			group.GET("/analyze", func(c *gin.Context) {
				_, err := linking.Run(ctx, c.Query("address"))
				if err != nil {
					logging.Error(ctx, "can't linking account", err)
				}
			})
		}
	}
}
