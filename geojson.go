package tz

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"os"
	"sync"
	"time"
)

type TimeZoneCollection struct {
	Features []*Feature
	*sync.RWMutex
}

type Feature struct {
	Geometry   Geometry
	Properties struct {
		Tzid string
	}
}

type Geometry struct {
	Coordinates []Coordinates
	MaxPoint    Point
	MinPoint    Point
}

type Coordinates struct {
	Polygon  []Point
	MaxPoint Point
	MinPoint Point
}

type Point struct {
	Lon float64
	Lat float64
}

var polygonType struct {
	Type       string
	Geometries []*Geometry
}

var polygon struct {
	Coordinates [][][]float64
}

var multiPolygon struct {
	Coordinates [][][][]float64
}

func NewGeoJsonTimeZoneLookup(geoJsonFile string) (TimeZoneLookup, error) {
	var fc = &TimeZoneCollection{
		Features: make([]*Feature, 500),
		RWMutex:  new(sync.RWMutex),
	}
	if len(geoJsonFile) == 0 {
		geoJsonFile = os.Getenv("GEO_JSON_FILE")
	}
	if len(geoJsonFile) == 0 {
		geoJsonFile = "timezones-with-oceans.geojson.zip"
	}

	g, err := zip.OpenReader(geoJsonFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read geojson file: %w", err)
	}

	if len(g.File) == 0 {
		return nil, fmt.Errorf("zip file is empty")
	}
	var file = g.File[0]
	var buf = bytes.NewBuffer([]byte{})
	src, err := file.Open()

	if err != nil {
		return nil, fmt.Errorf("failed to unzip file: %w", err)
	}

	if _, err := io.Copy(buf, src); err != nil {
		return nil, fmt.Errorf("failed to unzip file: %w", err)
	}

	if err := src.Close(); err != nil {
		return nil, fmt.Errorf("failed to read geojson file: %w", err)
	}
	if err := g.Close(); err != nil {
		return nil, fmt.Errorf("failed to read geojson file: %w", err)
	}

	if err := json.NewDecoder(buf).Decode(fc); err != nil {
		return nil, fmt.Errorf("failed to read geojson file: %w", err)
	}

	return fc, nil
}

func (g *Geometry) UnmarshalJSON(data []byte) (err error) {
	if err := json.Unmarshal(data, &polygonType); err != nil {
		return err
	}

	if g.MaxPoint.Lat == 0 && g.MaxPoint.Lon == 0 {
		g.MaxPoint = Point{Lon: -180, Lat: -90}
	}
	if g.MinPoint.Lat == 0 && g.MinPoint.Lon == 0 {
		g.MinPoint = Point{Lon: 180, Lat: 90}
	}

	switch polygonType.Type {
	case "Polygon":
		if err := json.Unmarshal(data, &polygon); err != nil {
			return err
		}
		pol := make([]Point, len(polygon.Coordinates[0]))
		for i, v := range polygon.Coordinates[0] {
			lon := v[0]
			lat := v[1]
			pol[i].Lon = lon
			pol[i].Lat = lat
			updateMaxMin(&g.MaxPoint, &g.MinPoint, lat, lon)
		}
		g.Coordinates = append(g.Coordinates, Coordinates{
			Polygon: pol,
		})
		return nil
	case "MultiPolygon":
		if err := json.Unmarshal(data, &multiPolygon); err != nil {
			return err
		}
		g.Coordinates = make([]Coordinates, len(multiPolygon.Coordinates))
		for j, poly := range multiPolygon.Coordinates {
			pol := make([]Point, len(poly[0]))
			for i, v := range poly[0] {
				var lon = v[0]
				var lat = v[1]
				pol[i].Lon = lon
				pol[i].Lat = lat
				updateMaxMin(&g.MaxPoint, &g.MinPoint, lat, lon)
			}
			g.Coordinates[j] = Coordinates{
				Polygon: pol,
			}
		}
		return nil
	default:
		return nil
	}
}

func updateMaxMin(maxPoint, minPoint *Point, lat, lon float64) {
	if maxPoint.Lat < lat {
		maxPoint.Lat = lat
	}
	if maxPoint.Lon < lon {
		maxPoint.Lon = lon
	}
	if minPoint.Lat > lat {
		minPoint.Lat = lat
	}
	if minPoint.Lon > lon {
		minPoint.Lon = lon
	}
}

func (fc *TimeZoneCollection) TimeZone(lat, lon float64) (tz string) {
	fc.Lock()
	defer fc.Unlock()
	for i := range fc.Features {
		if fc.Features[i].Geometry.MinPoint.Lat < lat &&
			fc.Features[i].Geometry.MinPoint.Lon < lon &&
			fc.Features[i].Geometry.MaxPoint.Lat > lat &&
			fc.Features[i].Geometry.MaxPoint.Lon > lon {
			for j := range fc.Features[i].Geometry.Coordinates {
				if fc.Features[i].Geometry.Coordinates[j].contains(Point{lon, lat}) {
					return fc.Features[i].Properties.Tzid
				}
			}
		}
	}
	return
}

func (fc *TimeZoneCollection) Location(lat, lon float64) (loc *time.Location, err error) {
	return time.LoadLocation(fc.TimeZone(lat, lon))
}

func (c *Coordinates) contains(point Point) (contains bool) {
	if len(c.Polygon) < 3 {
		return
	}
	contains = rayCast(point, c.Polygon[len(c.Polygon)-1], c.Polygon[0])
	for i := 1; i < len(c.Polygon); i++ {
		if rayCast(point, c.Polygon[i-1], c.Polygon[i]) {
			contains = !contains
			if contains == true {
				return
			}
		}
	}
	return
}

func rayCast(point, start, end Point) bool {
	if start.Lat > end.Lat {
		start, end = end, start
	}
	for point.Lat == start.Lat || point.Lat == end.Lat {
		point.Lat = math.Nextafter(point.Lat, math.Inf(1))
	}
	if point.Lat < start.Lat || point.Lat > end.Lat {
		return false
	}
	if start.Lon > end.Lon {
		if point.Lon > start.Lon {
			return false
		}
		if point.Lon < end.Lon {
			return true
		}
	} else {
		if point.Lon > end.Lon {
			return false
		}
		if point.Lon < start.Lon {
			return true
		}
	}
	return (point.Lat-start.Lat)/(point.Lon-start.Lon) >= (end.Lat-start.Lat)/(end.Lon-start.Lon)
}
