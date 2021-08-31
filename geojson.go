package tz

import (
	"encoding/json"
	"fmt"
	"github.com/catmullet/tz/geodb"
	"math"
	"sort"
	"time"
)

const (
	timeZonesFilename = "combined-with-oceans.json.snappy"
)

type GeoJsonLookup interface {
	TimeZone(lat, lon float64) string
	Location(lat, lon float64) (*time.Location, error)
}

type Collection struct {
	Features []*Feature
}

type Feature struct {
	Geometry   Geometry
	Properties map[string]string
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

func NewTZ() (GeoJsonLookup, error) {
	var (
		fc = &Collection{Features: make([]*Feature, 500)}
	)

	if b, err := newLocalGeoStorage(geodb.GeoDbEmbedDirectory).LoadFile(timeZonesFilename, &fc); err != nil || len(b) == 0 {
		return nil, fmt.Errorf("failed to load file, %w", err)
	}

	for i := range fc.Features {
		f := fc.Features[i]
		sort.SliceStable(f.Geometry.Coordinates, func(i, j int) bool {
			return f.Geometry.Coordinates[i].MinPoint.Lon <= f.Geometry.Coordinates[j].MinPoint.Lon
		})
	}

	sort.SliceStable(fc.Features, func(i, j int) bool {
		return fc.Features[i].Geometry.MinPoint.Lon <= fc.Features[j].Geometry.MinPoint.Lon
	})

	return fc, nil
}

func (g *Geometry) UnmarshalJSON(data []byte) (err error) {
	if err := json.Unmarshal(data, &polygonType); err != nil {
		return err
	}

	if g.MaxPoint.Lat == 0.0 && g.MaxPoint.Lon == 0.0 {
		g.MaxPoint = Point{Lon: -180.0, Lat: -90.0}
	}
	if g.MinPoint.Lat == 0.0 && g.MinPoint.Lon == 0.0 {
		g.MinPoint = Point{Lon: 180.0, Lat: 90.0}
	}

	switch polygonType.Type {
	case "Polygon":
		if err := json.Unmarshal(data, &polygon); err != nil {
			return err
		}
		coord := Coordinates{Polygon: make([]Point, len(polygon.Coordinates[0])), MaxPoint: Point{Lon: -180.0,
			Lat: -90.0}, MinPoint: Point{Lon: 180.0, Lat: 90.0}}
		for i, v := range polygon.Coordinates[0] {
			lon := v[0]
			lat := v[1]
			coord.Polygon[i].Lon = lon
			coord.Polygon[i].Lat = lat
			updateMaxMin(&coord.MaxPoint, &coord.MinPoint, lat, lon)
			updateMaxMin(&g.MaxPoint, &g.MinPoint, lat, lon)
		}
		g.Coordinates = append(g.Coordinates, coord)
		return nil
	case "MultiPolygon":
		if err := json.Unmarshal(data, &multiPolygon); err != nil {
			return err
		}
		g.Coordinates = make([]Coordinates, len(multiPolygon.Coordinates))
		for j, poly := range multiPolygon.Coordinates {
			coord := Coordinates{Polygon: make([]Point, len(poly[0])), MaxPoint: Point{Lon: -180.0,
				Lat: -90.0}, MinPoint: Point{Lon: 180.0, Lat: 90.0}}
			for i, v := range poly[0] {
				lon := v[0]
				lat := v[1]
				coord.Polygon[i].Lon = lon
				coord.Polygon[i].Lat = lat
				updateMaxMin(&coord.MaxPoint, &coord.MinPoint, lat, lon)
				updateMaxMin(&g.MaxPoint, &g.MinPoint, lat, lon)
			}
			g.Coordinates[j] = coord
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

func (fc Collection) Location(lat, lon float64) (*time.Location, error) {
	if tz := fc.TimeZone(lat, lon); tz != "" {
		return time.LoadLocation(tz)
	}
	return nil, fmt.Errorf("failed to find time zone")
}

// TimeZone Recurse over lower find function for lat lon.
// First shrinking the polygon for search and if we find it return it. If we didn't find it search on full polygon.
func (fc Collection) TimeZone(lat, lon float64) string {
	var start, end = 0.001, 0.0001
	if result := fc.find(lat, lon, start); result != "" {
		return result
	}
	return fc.find(lat, lon, end)
}

func (fc Collection) find(lat, lon, percentage float64) string {
	for _, feat := range fc.Features {
		f := feat
		properties := f.Properties
		if f.Geometry.MinPoint.Lat <= lat &&
			f.Geometry.MinPoint.Lon <= lon &&
			f.Geometry.MaxPoint.Lat >= lat &&
			f.Geometry.MaxPoint.Lon >= lon {
			for _, c := range f.Geometry.Coordinates {
				coord := c
				if coord.MinPoint.Lat <= lat &&
					coord.MinPoint.Lon <= lon &&
					coord.MaxPoint.Lat >= lat &&
					coord.MaxPoint.Lon >= lon {
					if coord.contains(Point{lon, lat},
						// get a percentage of the polygon, either shrinking it or leaving it alone.
						int(math.Max(float64(len(coord.Polygon))*percentage, 1))) {
						return properties["tzid"]
					}
				}
			}
		}
	}
	return ""
}

func (c Coordinates) contains(point Point, indexjump int) bool {
	var polygon = c.Polygon
	if windingNumber(point.Lat, point.Lon, polygon, indexjump) == 0 {
		return false
	}
	return true
}

func windingNumber(lat, lon float64, polygon []Point, indexjump int) int {
	if len(polygon) < 3 {
		return 0
	}

	var wn = 0
	var edgeCount = len(polygon) - indexjump

	for i, j := 0, indexjump; i < edgeCount; i, j = i+indexjump, j+indexjump {
		var apLat, apLon, bLat, bLon = polygon[i].Lat, polygon[i].Lon, polygon[j].Lat, polygon[j].Lon
		if apLat <= lat {
			if bLat > lat {
				if isLeft(lat, lon, apLat, apLon, bLat, bLon) > 0 {
					wn++
				}
			}
		} else {
			if polygon[j].Lat <= lat {
				if isLeft(lat, lon, apLat, apLon, bLat, bLon) < 0 {
					wn--
				}
			}
		}
	}
	return wn
}

func isLeft(lat, lon, latA, lonA, latB, lonB float64) float64 {
	return (lonB-lonA)*(lat-latA) - (lon-lonA)*(latB-latA)
}
