package binder

import (
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"strings"

	"github.com/labstack/echo"
	"github.com/pkg/errors"
)

type customBinder struct{}

func (cb *customBinder) bind(i interface{}, c echo.Context) error {
	req := c.Request()

	if req.ContentLength == 0 {
		if req.Method == http.MethodGet || req.Method == http.MethodDelete {
			qp := c.QueryParams()
			for key := range qp {
				if err := setField(i, key, qp.Get(key)); err != nil {
					return echo.NewHTTPError(http.StatusBadRequest, err.Error())
				}
			}
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

func setField(obj interface{}, name string, value interface{}) error {
	structValue := reflect.ValueOf(obj).Elem()
	structType := reflect.TypeOf(obj).Elem()
	structFieldValue := structValue.FieldByName(name)

	for i := 0; i < structType.NumField(); i++ {
		field := structType.Field(i)
		tag := field.Tag.Get("query")

		if tag == name {
			structFieldValue = structValue.Field(i)
			break
		}
	}

	if !structFieldValue.IsValid() || !structFieldValue.CanSet() {
		return errors.New(fmt.Sprintf("%s is not allowed", name))
	}

	structFieldType := structFieldValue.Type()
	val := reflect.ValueOf(value)

	if structFieldType.Kind() == reflect.Bool {
		switch val.String() {
		case "false":
			structFieldValue.SetBool(false)
			return nil
		case "true":
			structFieldValue.SetBool(true)
			return nil
		default:
			return errors.New(fmt.Sprintf("%s must be a boolean", name))
		}
	} else {
		structFieldValue.Set(val)
		return nil
	}
}
