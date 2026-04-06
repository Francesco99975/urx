package controllers

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"time"

	"github.com/Francesco99975/urx/cmd/boot"
	"github.com/Francesco99975/urx/internal/apperrors"
	"github.com/Francesco99975/urx/internal/cache"
	"github.com/Francesco99975/urx/internal/config"
	"github.com/Francesco99975/urx/internal/database"
	"github.com/Francesco99975/urx/internal/enums"
	"github.com/Francesco99975/urx/internal/helpers"
	"github.com/Francesco99975/urx/internal/repository"
	"github.com/Francesco99975/urx/views"
	"github.com/Francesco99975/urx/views/components"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
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

		validatedUrl, err := validateURL(helpers.RemoveWhitespace(url))
		if err != nil {
			return apperrors.SendReturnedHTMLErrorMessage(c, apperrors.ErrorMessage{Error: apperrors.GenericError{Code: http.StatusBadRequest, UserMessage: "cannot shorten invalid url", Message: fmt.Errorf("submitted url is invalid: %v", err).Error()}, Box: enums.Boxes.TOAST_TR, Persistance: "3000"}, nil)
		}

		hash := sha256.Sum256([]byte(validatedUrl))
		var long_url_hash []byte = hash[:]

		ctx := c.Request().Context()

		tx, err := database.Pool().BeginTx(ctx, pgx.TxOptions{})
		if err != nil {
			return apperrors.SendReturnedHTMLErrorMessage(c, apperrors.ErrorMessage{Error: apperrors.GenericError{Code: http.StatusInternalServerError, UserMessage: "failed to open database on signup", Message: fmt.Errorf("failed to open database on signup: %v", err).Error()}, Box: enums.Boxes.TOAST_TR, Persistance: "3000"}, nil)
		}
		defer database.HandleTransaction(ctx, tx, &err)

		repo := repository.New(tx)

		var code string

		var existsErr error
		code, existsErr = repo.GetCodeByHash(ctx, long_url_hash)
		if existsErr != nil {
			slug, err := boot.Slugerr.New()
			if err != nil {
				return apperrors.SendReturnedHTMLErrorMessage(c, apperrors.ErrorMessage{Error: apperrors.GenericError{Code: http.StatusInternalServerError, UserMessage: "failed to generate short url", Message: fmt.Errorf("failed to generate short code via Slugger: %v", err).Error()}, Box: enums.Boxes.TOAST_TR, Persistance: "3000"}, nil)
			}

			u, err := repo.CreateURL(ctx, repository.CreateURLParams{
				LongUrl:     validatedUrl,
				LongUrlHash: long_url_hash,
				Code:        slug.String(),
				UserID:      pgtype.UUID{Valid: false},
				ExpiresAt: pgtype.Timestamptz{
					Time:  time.Now().Add(time.Hour * 24 * 30 * 12),
					Valid: true,
				},
			})
			if err != nil {
				return apperrors.SendReturnedHTMLErrorMessage(c, apperrors.ErrorMessage{Error: apperrors.GenericError{Code: http.StatusInternalServerError, UserMessage: "failed to generate short url", Message: fmt.Errorf("failed to create database entry for short url: %v", err).Error()}, Box: enums.Boxes.TOAST_TR, Persistance: "3000"}, nil)
			}

			code = u.Code
		}

		go func() {
			log.Debugf("Setting warm cache for slug %s", code)
			err = cache.Client().Set(ctx, code, validatedUrl, time.Hour*24).Err()
			if err != nil {
				log.Errorf("Failed to set cache: %v", err)
			}
		}()

		html := helpers.MustRenderHTML(components.ShortenedResult(fmt.Sprintf("%s/%s", boot.Environment.URL, code)))

		return c.Blob(http.StatusOK, "text/html; charset=utf-8", html)
	}
}

func validateURL(input string) (string, error) {
	if input == "" || len(input) > 2048 {
		return "", errors.New("invalid length")
	}

	u, err := url.ParseRequestURI(input)
	if err != nil {
		return "", errors.New("invalid format")
	}

	if u.Scheme != "http" && u.Scheme != "https" {
		return "", errors.New("invalid scheme")
	}

	host := u.Hostname()

	ip := net.ParseIP(host)
	if ip != nil && helpers.IsPrivateIP(ip) {
		return "", errors.New("private IP not allowed")
	}

	u.Fragment = ""
	return u.String(), nil
}
