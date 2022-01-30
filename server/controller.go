package server

import (
	"github.com/gin-gonic/gin"
	"nir/linking"
)

type Controller struct {
}

func NewController() *Controller {
	return &Controller{}
}

func (c *Controller) Group(router *gin.Engine) {
	group := router.Group("/")
	{
		group.GET("/analyze", aml)
	}
}

func aml(ctx *gin.Context) {
	linking.Run(ctx, ctx.Query("address"))
}
