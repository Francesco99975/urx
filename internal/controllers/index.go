package controllers

import (
	"fmt"

	"net/http"

	"github.com/Francesco99975/urx/internal/apperrors"

	"github.com/Francesco99975/urx/internal/config"

	"github.com/Francesco99975/urx/internal/enums"
	"github.com/Francesco99975/urx/internal/helpers"

	"github.com/Francesco99975/urx/internal/shared"
	"github.com/Francesco99975/urx/views"
	"github.com/Francesco99975/urx/views/components"
	"github.com/labstack/echo/v4"
	"github.com/labstack/gommon/log"
)

func Index() echo.HandlerFunc {
	return func(c echo.Context) error {
		data := config.GetDefaultSite(c.Request())

		data.CSRF = c.Get("csrf").(string)
		log.Debugf("CSRF: %s", data.CSRF)
		data.Nonce = c.Get("nonce").(string)
		log.Debugf("Nonce: %s", data.Nonce)

		html := helpers.MustRenderHTML(views.Index(data))

		return c.Blob(http.StatusOK, "text/html; charset=utf-8", html)
	}
}

func Shorten() echo.HandlerFunc {
	return func(c echo.Context) error {
		url := c.FormValue("url")

		ctx := c.Request().Context()

		shortenedUrl, err := shared.Shorten(ctx, url)
		if err != nil {
			return apperrors.SendReturnedHTMLErrorMessage(c, apperrors.ErrorMessage{Error: apperrors.GenericError{Code: http.StatusInternalServerError, UserMessage: "failed to shorten url", Message: fmt.Errorf("failed to shorten url: %v", err).Error()}, Box: enums.Boxes.TOAST_TR, Persistance: "3000"}, nil)
		}

		qrCode, err := helpers.GenerateQRCodeBase64(shortenedUrl)
		if err != nil {
			return apperrors.SendReturnedHTMLErrorMessage(c, apperrors.ErrorMessage{Error: apperrors.GenericError{Code: http.StatusInternalServerError, UserMessage: "failed to shorten url", Message: fmt.Errorf("failed to shorten url: %v", err).Error()}, Box: enums.Boxes.TOAST_TR, Persistance: "3000"}, nil)
		}

		html := helpers.MustRenderHTML(components.ShortenedResult(shortenedUrl, qrCode))

		return c.Blob(http.StatusOK, "text/html; charset=utf-8", html)
	}
}
