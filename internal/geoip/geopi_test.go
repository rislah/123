package geoip_test

import (
	"testing"

	"github.com/rislah/fakes/internal/geoip"
	"github.com/stretchr/testify/assert"
)

func TestGeoIP(t *testing.T) {
	_, err := geoip.New(".")
	assert.Error(t, err)
}
