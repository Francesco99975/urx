package controllers

import (
	"net/http"

	"github.com/Francesco99975/urx/internal/config"
	"github.com/Francesco99975/urx/internal/helpers"
	"github.com/Francesco99975/urx/views"
	"github.com/labstack/echo/v4"
)

func PrivacyPolicy() echo.HandlerFunc {
	return func(c echo.Context) error {
		data := config.GetDefaultSite(c.Request())
		html := helpers.MustRenderHTML(views.PrivacyPolicy(data))
		return c.Blob(http.StatusOK, "text/html; charset=utf-8", html)
	}
}

func TermsOfService() echo.HandlerFunc {
	return func(c echo.Context) error {
		data := config.GetDefaultSite(c.Request())
		html := helpers.MustRenderHTML(views.Terms(data))
		return c.Blob(http.StatusOK, "text/html; charset=utf-8", html)
	}
}
