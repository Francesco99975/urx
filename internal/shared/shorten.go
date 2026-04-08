package shared

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"net"
	"net/url"
	"time"

	"github.com/Francesco99975/urx/cmd/boot"
	"github.com/Francesco99975/urx/internal/cache"
	"github.com/Francesco99975/urx/internal/database"
	"github.com/Francesco99975/urx/internal/helpers"
	"github.com/Francesco99975/urx/internal/repository"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/labstack/gommon/log"
)

func Shorten(ctx context.Context, url string) (string, error) {
	validatedUrl, err := validateURL(helpers.RemoveWhitespace(url))
	if err != nil {
		return "", fmt.Errorf("submitted url is invalid: %w", err)
	}

	hash := sha256.Sum256([]byte(validatedUrl))
	long_url_hash := hash[:]

	tx, err := database.Pool().BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return "", fmt.Errorf("failed to open database to create transaction: %w", err)
	}
	defer database.HandleTransaction(ctx, tx, &err)

	repo := repository.New(tx)

	var code string

	var existsErr error
	code, existsErr = repo.GetCodeByHash(ctx, long_url_hash)
	if existsErr != nil {
		slug, err := boot.Slugerr.New()
		if err != nil {
			return "", fmt.Errorf("failed to generate short code via Slugger: %w", err)
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
			return "", fmt.Errorf("failed to create database entry for short url: %w", err)
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

	return fmt.Sprintf("%s/%s", boot.Environment.URL, code), nil
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

	log.Debugf("Validating host: %s", host)

	if host == "localhost" {
		return "", errors.New("localhost not allowed")
	}

	ip := net.ParseIP(host)
	log.Debugf("Parsed IP: %s", ip)
	if ip != nil && helpers.IsPrivateIP(ip) {
		return "", errors.New("private IP not allowed")
	}

	u.Fragment = ""
	return u.String(), nil
}
