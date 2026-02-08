package httpapi

import (
	"github.com/gin-gonic/gin"
)

func NewEngine() *gin.Engine {
	r := gin.New()
	r.Use(gin.Recovery())
	r.GET("/healthz", func(c *gin.Context) { c.Status(200) })

	return r
}
