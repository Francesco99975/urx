package geo

import (
	"log"

	"github.com/oschwald/geoip2-golang"
)

var GeoDB *geoip2.Reader

func InitGeo(path string) {
	var err error
	GeoDB, err = geoip2.Open(path)
	if err != nil {
		log.Fatalf("failed to open geoip db: %v", err)
	}
}
