package timezonelookup

type TimeZoneLookup interface {
	TimeZone(lat, lon float64) string
}
