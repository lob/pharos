package binder

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/labstack/echo"
)

type customBinder struct{}

func (cb *customBinder) bind(i interface{}, c echo.Context) error {
	req := c.Request()

	if req.ContentLength == 0 {
		if req.Method == http.MethodGet || req.Method == http.MethodDelete {
			return nil
		}
		return echo.NewHTTPError(http.StatusBadRequest, "request body can't be empty")
	}

	contentType := req.Header.Get(echo.HeaderContentType)

	if strings.HasPrefix(contentType, echo.MIMEApplicationJSON) {
		d := json.NewDecoder(req.Body)
		d.DisallowUnknownFields()

		if err := d.Decode(i); err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, err.Error())
		}

		return nil
	}

	return echo.ErrUnsupportedMediaType
}
