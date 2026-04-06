package controllers

import (
	"context"
	"errors"
	"time"

	"net/http"

	"github.com/Francesco99975/urx/internal/cache"
	"github.com/Francesco99975/urx/internal/database"
	"github.com/Francesco99975/urx/internal/repository"
	"github.com/jackc/pgx/v5"
	"github.com/labstack/echo/v4"
	"github.com/labstack/gommon/log"
)

func Redirect() echo.HandlerFunc {
	return func(c echo.Context) error {
		slug := c.Param("slug")

		ctx := c.Request().Context()

		tx, err := database.Pool().BeginTx(ctx, pgx.TxOptions{})
		if err != nil {
			return echo.NewHTTPError(http.StatusBadGateway, "Server Error")
		}

		defer database.HandleTransaction(ctx, tx, &err)

		repo := repository.New(tx)

		long_url, err := cache.Client().Get(ctx, slug).Result()
		if err == nil && long_url != "" {
			log.Debugf("Cache hit for slug %s", slug)
			go incrementClickBySlug(slug)
			return c.Redirect(http.StatusFound, long_url)
		}

		url, err := repo.GetAndIncrementURL(ctx, slug)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return echo.ErrNotFound
			}
			return echo.NewHTTPError(http.StatusBadGateway, "Server Error")
		}

		go func(u *repository.GetAndIncrementURLRow) {
			log.Debugf("Cache miss for slug %s", slug)
			bgCtx := context.Background()
			err = cache.Client().Set(bgCtx, slug, u.LongUrl, 24*time.Hour).Err()
			if err != nil {
				log.Errorf("Failed to set cache: %v", err)
			}
		}(url)

		go incrementClickByID(url.ID)

		return c.Redirect(http.StatusFound, url.LongUrl)
	}
}

func incrementClickBySlug(slug string) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	repo := repository.New(database.Pool())

	url, err := repo.GetAndIncrementURL(ctx, slug)
	if err != nil {
		log.Errorf("increment lookup failed: %v", err)
		return
	}

	err = repo.IncrementClicks(ctx, url.ID)
	if err != nil {
		log.Errorf("increment failed: %v", err)
	}
}

func incrementClickByID(id int64) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	repo := repository.New(database.Pool())

	err := repo.IncrementClicks(ctx, id)
	if err != nil {
		log.Errorf("increment failed: %v", err)
	}
}
