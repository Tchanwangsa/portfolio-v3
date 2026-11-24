package portfolioapi

import (
	"errors"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/ringsaturn/tzf"
)

// The time zone finder holds roughly 50MB of boundary data, so it loads on the
// first webhook delivery that carries a route map instead of at startup.
var zoneFinder = sync.OnceValues(tzf.NewDefaultFinder)

// routeZone finds the time zone where a run started from the Google Static Maps
// URL that IFTTT supplies as the RouteMapImageUrl ingredient.
func routeZone(mapURL string) (*time.Location, error) {
	lat, lng, err := routeStart(mapURL)
	if err != nil {
		return nil, err
	}
	finder, err := zoneFinder()
	if err != nil {
		return nil, err
	}
	name := finder.GetTimezoneName(lng, lat)
	if name == "" {
		return nil, errors.New("no time zone at the route's start")
	}
	return time.LoadLocation(name)
}

// routeStart returns the first coordinate of a static map's path, falling back
// to its markers or centre for maps drawn without a path.
func routeStart(mapURL string) (lat, lng float64, err error) {
	parsed, err := url.Parse(strings.TrimSpace(mapURL))
	if err != nil || parsed.Host == "" {
		return 0, 0, errors.New("not a map URL")
	}
	query := parsed.Query()
	for _, key := range []string{"path", "markers", "center"} {
		for _, value := range query[key] {
			// An encoded polyline comes last and may itself contain "|".
			if _, encoded, ok := strings.Cut(value, "enc:"); ok {
				if lat, lng, err = firstPolylinePoint(encoded); err == nil {
					return lat, lng, nil
				}
				continue
			}
			for _, part := range strings.Split(value, "|") {
				if lat, lng, err = parseLatLng(part); err == nil {
					return lat, lng, nil
				}
			}
		}
	}
	return 0, 0, errors.New("map URL has no route coordinates")
}

func parseLatLng(value string) (float64, float64, error) {
	latText, lngText, ok := strings.Cut(value, ",")
	if !ok {
		return 0, 0, errors.New("not a coordinate")
	}
	lat, latErr := strconv.ParseFloat(strings.TrimSpace(latText), 64)
	lng, lngErr := strconv.ParseFloat(strings.TrimSpace(lngText), 64)
	if latErr != nil || lngErr != nil || !validCoordinate(lat, lng) {
		return 0, 0, errors.New("not a coordinate")
	}
	return lat, lng, nil
}

// firstPolylinePoint decodes the first point of a Google encoded polyline.
func firstPolylinePoint(encoded string) (float64, float64, error) {
	var values [2]float64
	pos := 0
	for i := range values {
		var result, shift int
		for {
			if pos >= len(encoded) || shift > 30 {
				return 0, 0, errors.New("truncated polyline")
			}
			b := int(encoded[pos]) - 63
			pos++
			if b < 0 || b > 63 {
				return 0, 0, errors.New("invalid polyline")
			}
			result |= (b & 0x1f) << shift
			shift += 5
			if b < 0x20 {
				break
			}
		}
		if result&1 != 0 {
			result = ^result
		}
		values[i] = float64(result>>1) / 1e5
	}
	if !validCoordinate(values[0], values[1]) {
		return 0, 0, errors.New("polyline starts outside valid coordinates")
	}
	return values[0], values[1], nil
}

func validCoordinate(lat, lng float64) bool {
	return lat >= -90 && lat <= 90 && lng >= -180 && lng <= 180 && (lat != 0 || lng != 0)
}
