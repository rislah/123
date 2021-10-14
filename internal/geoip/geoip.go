package geoip

import (
	"net"

	"github.com/oschwald/geoip2-golang"
)

type GeoIP struct {
	db *geoip2.Reader
}

func New(path string) (GeoIP, error) {
	db, err := geoip2.Open(path)
	if err != nil {
		return GeoIP{}, err
	}

	return GeoIP{db}, nil
}

func (g GeoIP) LookupCountry(ip net.IP) (string, error) {
	resp, err := g.db.Country(ip)
	return resp.Country.IsoCode, err
}
