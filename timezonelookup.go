package timezonelookup

import "time"

type TimeZoneLookup interface {
	TimeZone(lat, lon float64) string
	Location(lat, lon float64) (*time.Location, error)
}
