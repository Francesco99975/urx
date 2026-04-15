package main

import (
	"context"
	"fmt"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/Francesco99975/urx/cmd/boot"
	"github.com/Francesco99975/urx/internal/api"
	"github.com/Francesco99975/urx/internal/config"
	"github.com/Francesco99975/urx/internal/enums"
	"github.com/Francesco99975/urx/internal/helpers"

	"github.com/Francesco99975/urx/internal/controllers"
	"github.com/Francesco99975/urx/internal/middlewares"
	"github.com/Francesco99975/urx/views"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"
)

func createRouter(ctx context.Context) *echo.Echo {
	e := echo.New()
	e.Use(middleware.RequestLogger())
	e.Use(middleware.RemoveTrailingSlash())
	e.Use(middlewares.RateLimiter())
	// Apply Gzip middleware, but skip it for /metrics
	e.Use(middleware.GzipWithConfig(middleware.GzipConfig{
		Level: 5,
		Skipper: func(c echo.Context) bool {
			return c.Path() == "/metrics" // Skip compression for /metrics
		},
	}))
	e.Use(middlewares.MonitoringMiddleware())
	e.GET("/metrics", echo.WrapHandler(promhttp.Handler()), middlewares.MetricsAccessMiddleware())
	e.GET("/healthcheck", func(c echo.Context) error {
		time.Sleep(5 * time.Second)
		return c.JSON(http.StatusOK, "OK")
	})
	e.POST("/csp-violation-report", func(c echo.Context) error {
		log.Warnf("CSP Violation Report: %s", c.Request().RequestURI)
		return c.NoContent(http.StatusOK)
	})

	e.GET("/sw.js", func(c echo.Context) error {
		c.Response().Header().Set("Content-Type", "application/javascript")
		c.Response().Header().Set("Cache-Control", "no-cache")
		return c.File("./static/sw.js")
	})

	e.Static("/assets", "./static")
	e.GET("/assets/dist/*", func(c echo.Context) error {
		c.Response().Header().Set("Cache-Control", "public, max-age=31536000, immutable")
		return c.File(filepath.Join("./static/dist", c.Param("*")))
	})

	e.GET("/sitemap.xml", func(c echo.Context) error {
		sitemap := config.GetDefaultSite(c.Request()).Sitemap

		if sitemap == nil {
			log.Warnf("Sitemap not found")
			return c.NoContent(404)
		}

		return c.Blob(200, "application/xml", sitemap)
	})

	e.GET("/robots.txt", func(c echo.Context) error {
		var content string

		baseURL := boot.Environment.URL

		if boot.Environment.GoEnv == enums.Environments.PRODUCTION {
			content = fmt.Sprintf(`User-agent: *
Allow: /

Sitemap: %s/sitemap.xml
`, baseURL)
		} else {
			content = fmt.Sprintf(`User-agent: *
Disallow: /

Sitemap: %s/sitemap.xml
`, baseURL)
		}

		return c.Blob(200, "text/plain", []byte(content))
	})

	web := e.Group("")

	web.Use(middlewares.SecurityHeaders())

	if boot.Environment.GoEnv == enums.Environments.DEVELOPMENT {
		e.Logger.SetLevel(log.DEBUG)
		log.SetLevel(log.DEBUG)

	}

	if boot.Environment.GoEnv == enums.Environments.PRODUCTION {
		e.Logger.SetLevel(log.INFO)
		log.SetLevel(log.INFO)

	}

	web.Use(middleware.CSRFWithConfig(middleware.CSRFConfig{
		TokenLookup:    "form:_csrf,header:X-CSRF-Token",
		CookieName:     "csrf_token",
		CookiePath:     "/",
		CookieHTTPOnly: true,
		CookieSecure:   boot.Environment.GoEnv == enums.Environments.PRODUCTION,
		CookieSameSite: http.SameSiteLaxMode,
		Skipper: func(c echo.Context) bool {
			// Skip CSRF for the /webhook route
			return c.Path() == "/webhook"

		},
	}))

	web.GET("/", controllers.Index())
	web.GET("/stats", controllers.Stats())
	web.GET("/privacy", controllers.PrivacyPolicy())
	web.GET("/terms", controllers.TermsOfService())
	web.POST("/shorten", controllers.Shorten())
	web.GET("/:slug", controllers.Redirect())

	apigrp := e.Group("/api")

	apiv1 := apigrp.Group("/v1")
	apiv1.POST("/shorten/:url", api.Shorten())

	e.HTTPErrorHandler = serverErrorHandler

	return e
}

func serverErrorHandler(err error, c echo.Context) {

	if c.Response().Committed {
		return
	}
	// Default to internal server error (500)
	code := http.StatusInternalServerError
	var message any = "Internal Server Error"

	// Check if it's an echo.HTTPError
	if he, ok := err.(*echo.HTTPError); ok {
		code = he.Code
		message = he.Message
	}

	// Check the Accept header to decide the response format
	if strings.Contains(c.Request().Header.Get("Accept"), "application/json") {
		// Respond with JSON if the client prefers JSON
		_ = c.JSON(code, map[string]any{
			"error":   true,
			"message": message,
			"status":  code,
		})
	} else {
		// Prepare data for rendering the error page (HTML)
		data := config.GetDefaultSite(c.Request())

		html := helpers.MustRenderHTML(views.Error(data, fmt.Sprintf("%d", code), message.(string)))

		// Respond with HTML (default) if the client prefers HTML
		_ = c.Blob(code, "text/html; charset=utf-8", html)
	}
}
