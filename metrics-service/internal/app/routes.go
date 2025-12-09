package app

import (
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func (a *App) RegisterRoutes() {
	app := a.Router.Group("/metrics")
	app.GET("", gin.WrapH(promhttp.Handler()))

}
