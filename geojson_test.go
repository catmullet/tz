package timezonelookup

import (
	"fmt"
	"os"
	"testing"
)

var tzl TimeZoneLookup

type coords struct {
	Lat float64
	Lon float64
}

var querys = []coords{
	{Lat: 5.261417, Lon: -3.925778},   // Abijan Airport
	{Lat: -15.678889, Lon: 34.973889}, // Blantyre Airport
	{Lat: -12.65945, Lon: 18.25674},
	{Lat: 41.8976, Lon: -87.6205},
	{Lat: 47.6897, Lon: -122.4023},
	{Lat: 42.7235, Lon: -73.6931},
	{Lat: 42.5807, Lon: -83.0223},
	{Lat: 36.8381, Lon: -84.8500},
	{Lat: 40.1674, Lon: -85.3583},
	{Lat: 37.9643, Lon: -86.7453},
	{Lat: 38.6043, Lon: -90.2417},
	{Lat: 41.1591, Lon: -104.8261},
	{Lat: 35.1991, Lon: -111.6348},
	{Lat: 43.1432, Lon: -115.6750},
	{Lat: 47.5886, Lon: -122.3382},
	{Lat: 58.3168, Lon: -134.4397},
	{Lat: 21.4381, Lon: -158.0493},
	{Lat: 42.7000, Lon: -80.0000},
	{Lat: 51.0036, Lon: -114.0161},
	{Lat: -16.4965, Lon: -68.1702},
	{Lat: -31.9369, Lon: 115.8453},
	{Lat: 42.0000, Lon: -87.5000},
	{Lat: 41.8976, Lon: -87.6205},
	{Lat: 47.6897, Lon: -122.4023},
	{Lat: 42.7235, Lon: -73.6931},
	{Lat: 42.5807, Lon: -83.0223},
	{Lat: 36.8381, Lon: -84.8500},
	{Lat: 40.1674, Lon: -85.3583},
	{Lat: 37.9643, Lon: -86.7453},
	{Lat: 38.6043, Lon: -90.2417},
	{Lat: 41.1591, Lon: -104.8261},
	{Lat: 35.1991, Lon: -111.6348},
	{Lat: 43.1432, Lon: -115.6750},
	{Lat: 47.5886, Lon: -122.3382},
	{Lat: 58.3168, Lon: -134.4397},
	{Lat: 21.4381, Lon: -158.0493},
	{Lat: 42.7000, Lon: -80.0000},
	{Lat: 51.0036, Lon: -114.0161},
	{Lat: -16.4965, Lon: -68.1702},
	{Lat: -31.9369, Lon: 115.8453},
	{Lat: 42.0000, Lon: -87.5000},
	{Lat: 41.8976, Lon: -87.6205},
	{Lat: 47.6897, Lon: -122.4023},
	{Lat: 42.7235, Lon: -73.6931},
	{Lat: 42.5807, Lon: -83.0223},
	{Lat: 36.8381, Lon: -84.8500},
	{Lat: 40.1674, Lon: -85.3583},
	{Lat: 37.9643, Lon: -86.7453},
	{Lat: 38.6043, Lon: -90.2417},
	{Lat: 41.1591, Lon: -104.8261},
	{Lat: 35.1991, Lon: -111.6348},
	{Lat: 43.1432, Lon: -115.6750},
	{Lat: 47.5886, Lon: -122.3382},
	{Lat: 58.3168, Lon: -134.4397},
	{Lat: 21.4381, Lon: -158.0493},
	{Lat: 42.7000, Lon: -80.0000},
	{Lat: 51.0036, Lon: -114.0161},
	{Lat: -16.4965, Lon: -68.1702},
	{Lat: -31.9369, Lon: 115.8453},
	{Lat: 42.0000, Lon: -87.5000},
}

func TestMain(m *testing.M) {
	var err error
	tzl, err = NewGeoJsonTimeZoneLookup("timezones-with-oceans.geojson.zip")
	if err != nil {
		os.Exit(1)
	}
	code := m.Run()
	os.Exit(code)
}

func Test(t *testing.T) {
	for _, q := range querys {
		t.Run(fmt.Sprintf("coordinates_%v_%v", q.Lat, q.Lon), func(t *testing.T) {
			tzid := tzl.TimeZone(q.Lat, q.Lon)
			t.Log(tzid)
			if len(tzid) == 0 {
				t.Fail()
			}
		})
	}
}

func Benchmark(b *testing.B) {
	for _, bm := range querys {
		b.Run(fmt.Sprintf("coordinates_%v_%v", bm.Lat, bm.Lon), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				tzl.TimeZone(bm.Lat, bm.Lon)
			}
		})
	}
}
