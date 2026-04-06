package helpers

import (
	"net/netip"

	"github.com/labstack/echo/v4"
)

func GetClientIP(c echo.Context) (*netip.Addr, error) {
	ipStr := c.RealIP()

	addr, err := netip.ParseAddr(ipStr)
	if err != nil {
		return nil, err
	}

	return &addr, nil
}
