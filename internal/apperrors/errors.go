package apperrors

import (
	"fmt"

	"github.com/Francesco99975/urx/internal/config"
	"github.com/Francesco99975/urx/internal/enums"
	"github.com/Francesco99975/urx/internal/helpers"
	"github.com/Francesco99975/urx/internal/models"
	"github.com/Francesco99975/urx/internal/monitoring"
	"github.com/Francesco99975/urx/views"
	"github.com/Francesco99975/urx/views/components"
	"github.com/labstack/echo/v4"
	"github.com/labstack/gommon/log"
)

type GenericError struct {
	Code        int      `json:"code"`
	Message     string   `json:"message"`
	UserMessage string   `json:"userMessage"`
	Errors      []string `json:"errors"`
}

type ErrorMessage struct {
	Error       GenericError `json:"error"`
	Box         enums.Box    `json:"box"`
	Persistance string       `json:"persistence"`
}

func (ge *GenericError) Stringify() string {
	return fmt.Sprintf("[%d] %s <-- %v", ge.Code, ge.Message, ge.Errors)
}

func SendReturnedGenericJSONError(c echo.Context, err GenericError, r *helpers.Reporter) error {
	monitoring.RecordError(fmt.Sprintf("%d", err.Code))
	log.Error(err.Stringify())

	if r != nil {
		_ = r.Report(helpers.SeverityLevels.ERROR, err.Stringify())
	}

	return c.JSON(err.Code, models.JSONErrorResponse{Code: err.Code, Message: err.UserMessage, Errors: err.Errors})
}

func SendReturnedGenericHTMLError(c echo.Context, err GenericError, r *helpers.Reporter) error {
	monitoring.RecordError(fmt.Sprintf("%d", err.Code))
	log.Error(err.Stringify())

	if r != nil {
		_ = r.Report(helpers.SeverityLevels.ERROR, err.Stringify())
	}

	html := helpers.MustRenderHTML(views.Error(config.GetDefaultSite(c.Request()), fmt.Sprintf("%d", err.Code), err.UserMessage))

	return c.Blob(err.Code, "text/html", html)
}

func SendReturnedHTMLErrorMessage(c echo.Context, err ErrorMessage, r *helpers.Reporter) error {
	monitoring.RecordError(fmt.Sprintf("%d", err.Error.Code))
	log.Error(err.Error.Stringify())

	if r != nil {
		_ = r.Report(helpers.SeverityLevels.ERROR, err.Error.Stringify())
	}

	html := helpers.MustRenderHTML(components.ErrorMsg(err.Error.UserMessage, err.Box, err.Persistance))

	return c.Blob(err.Error.Code, "text/html", html)
}
