package health

import (
	"github.com/labstack/echo/v4"
)

func (c *Controller) RegisterRoutes(router *echo.Group) {
	router.POST("/now", c.TestNow)
}
