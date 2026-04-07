package controllers

import (
	"context"
	"errors"
	"time"

	"net/http"
	"net/netip"

	"github.com/Francesco99975/urx/internal/cache"
	"github.com/Francesco99975/urx/internal/database"
	"github.com/Francesco99975/urx/internal/helpers"
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
			go func(c echo.Context, slug string) {
				ip, err := helpers.GetClientIP(c)
				if err == nil {
					incrementClickBySlug(slug, ip, new(c.Request().Header.Get("User-Agent")), new(c.Request().Header.Get("Referer")))
				} else {
					log.Errorf("Failed to get client IP: %v", err)
				}
			}(c, slug)
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

		go func(c echo.Context, id int64) {
			ip, err := helpers.GetClientIP(c)
			if err == nil {
				incrementClickByID(id, ip, new(c.Request().Header.Get("User-Agent")), new(c.Request().Header.Get("Referer")))
			} else {
				log.Errorf("Failed to get client IP: %v", err)
			}
		}(c, url.ID)

		return c.Redirect(http.StatusFound, url.LongUrl)
	}
}

func incrementClickBySlug(slug string, ip *netip.Addr, ua *string, rf *string) {
	log.Infof("Incrementing click for slug %s", slug)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	tx, err := database.Pool().BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		log.Errorf("failed to begin transaction: %v", err)
		return
	}
	defer database.HandleTransaction(ctx, tx, &err)

	repo := repository.New(tx)

	url, err := repo.GetAndIncrementURL(ctx, slug)
	if err != nil {
		log.Errorf("increment lookup failed: %v", err)
		return
	}

	_, err = repo.CreateClick(ctx, repository.CreateClickParams{
		UrlID:     url.ID,
		Ip:        ip,
		UserAgent: ua,
		Referer:   rf,
	})
	if err != nil {
		log.Errorf("failed to create click: %v", err)
	}
}

func incrementClickByID(id int64, ip *netip.Addr, ua *string, rf *string) {
	log.Infof("Incrementing click for id %d", id)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	tx, err := database.Pool().BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		log.Errorf("failed to begin transaction: %v", err)
		return
	}
	defer database.HandleTransaction(ctx, tx, &err)

	repo := repository.New(tx)

	_, err = repo.CreateClick(ctx, repository.CreateClickParams{
		UrlID:     id,
		Ip:        ip,
		UserAgent: ua,
		Referer:   rf,
	})
	if err != nil {
		log.Errorf("failed to create click: %v", err)
	}
}
