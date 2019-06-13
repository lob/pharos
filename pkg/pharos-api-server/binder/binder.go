package binder

import (
	"context"
	"fmt"
	"net/http"
	"reflect"
	"strings"

	"github.com/labstack/echo"
	"gopkg.in/go-playground/mold.v2"
	"gopkg.in/go-playground/mold.v2/modifiers"
	"gopkg.in/go-playground/validator.v9"
)

// Binder is a custom struct that implements the Echo Binder interface. It binds
// to a struct, uses mold to clean up the params, and validator to validate
// them.
type Binder struct {
	binder   *customBinder
	conform  *mold.Transformer
	validate *validator.Validate
}

// New initializes a new Binder instance with the appropriate validation
// functions registered.
func New() *Binder {
	binder := &customBinder{}
	conform := modifiers.New()
	validate := validator.New()

	// RegisterTagNameFunc allows us to specify a function to fetch the name of
	// the field by tag such that validator.FieldError.Field() returns the JSON
	// name as opposed to the Go struct field name.
	validate.RegisterTagNameFunc(func(fld reflect.StructField) string {
		return strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
	})

	return &Binder{binder, conform, validate}
}

// Bind binds, modifies, and validates payloads against the given struct.
func (b *Binder) Bind(i interface{}, c echo.Context) error {
	// Extract values from the request in the echo.Context into our interface i.
	if err := b.binder.bind(i, c); err != nil {
		return err
	}

	// Modify values based on the struct tags in our interface. Most likely this
	// is trimming whitespace from values.
	if err := b.conform.Struct(context.Background(), i); err != nil {
		return err
	}

	// Validate that the values on our struct are valid. If there are any errors,
	// format the first error and return it as an HTTP error.
	if err := b.validate.Struct(i); err != nil {
		errs := err.(validator.ValidationErrors)
		msg := format(errs[0])
		return echo.NewHTTPError(http.StatusUnprocessableEntity, msg)
	}

	return nil
}

func format(err validator.FieldError) string {
	if err.Tag() == "required" {
		return fmt.Sprintf("%s is required", err.Field())
	}

	if err.Tag() == "url" {
		return fmt.Sprintf("%s must be a valid URL", err.Field())
	}

	return fmt.Sprintf("%s is invalid", err.Field())
}
