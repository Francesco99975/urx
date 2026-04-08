package api

import (
	"fmt"
	"net/http"

	"github.com/Francesco99975/urx/internal/apperrors"
	"github.com/Francesco99975/urx/internal/enums"
	"github.com/Francesco99975/urx/internal/models"
	"github.com/Francesco99975/urx/internal/shared"
	"github.com/labstack/echo/v4"
)

func Shorten() echo.HandlerFunc {
	return func(c echo.Context) error {
		url := c.Param("url")

		ctx := c.Request().Context()

		shortenedUrl, err := shared.Shorten(ctx, url)
		if err != nil {
			return apperrors.SendReturnedHTMLErrorMessage(c, apperrors.ErrorMessage{Error: apperrors.GenericError{Code: http.StatusInternalServerError, UserMessage: "failed to shorten url", Message: fmt.Errorf("failed to shorten url: %v", err).Error()}, Box: enums.Boxes.TOAST_TR, Persistance: "3000"}, nil)
		}

		return c.JSON(http.StatusOK, models.JSONShortenedResponse{ShortenedUrl: shortenedUrl})
	}
}
