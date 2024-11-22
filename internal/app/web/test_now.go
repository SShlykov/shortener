package health

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/sshlykov/shortener/internal/app/web/dto"
)

func (c *Controller) TestNow(ectx echo.Context) error {
	secret, err := dto.EjectSecret(ectx)
	if err != nil {
		return ectx.JSON(http.StatusBadRequest, echo.Map{"error": "should be secret in request body"})
	}

	if err = secret.Validate(); err != nil {
		return ectx.JSON(http.StatusBadRequest, echo.Map{"error": "secret is empty"})
	}

	res, err := c.svc.SelectNow(ectx.Request().Context())
	if err != nil {
		return ectx.JSON(http.StatusInternalServerError, echo.Map{"error": "unable to get now"})
	}

	return ectx.JSON(http.StatusOK, echo.Map{"now": res})
}
