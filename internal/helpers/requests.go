package helpers

import (
	"net/netip"

	"github.com/labstack/echo/v4"
	"github.com/labstack/gommon/log"
)

func GetClientIP(c echo.Context) (*netip.Addr, error) {
	ipStr := c.RealIP()

	addr, err := netip.ParseAddr(ipStr)
	if err != nil {
		log.Errorf("Failed to parse client IP: %s", ipStr)
		return nil, err
	}

	return &addr, nil
}
