package timezonelookup

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
)

// TimeZoneCollection ...
type TimeZoneCollection struct {
	Features []*Feature
}

type Feature struct {
	Geometry   Geometry
	Properties struct {
		Tzid string
	}
}

type Geometry struct {
	Coordinates []Coordinates
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
	var fc = &TimeZoneCollection{}
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

//UnmarshalJSON ...
func (g *Geometry) UnmarshalJSON(data []byte) (err error) {
	if err := json.Unmarshal(data, &polygonType); err != nil {
		return err
	}

	var minPoint = Point{Lon: 180, Lat: 90}
	var maxPoint = Point{Lon: -180, Lat: -90}

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
			updateMaxMinLatLon(&maxPoint, &minPoint, lat, lon)
		}
		g.Coordinates = append(g.Coordinates, Coordinates{
			Polygon:  pol,
			MaxPoint: maxPoint,
			MinPoint: minPoint,
		})
		return nil
	case "MultiPolygon":
		if err := json.Unmarshal(data, &multiPolygon); err != nil {
			return err
		}
		g.Coordinates = make([]Coordinates, len(multiPolygon.Coordinates))
		for j, poly := range multiPolygon.Coordinates {
			minPoint = Point{Lon: 180, Lat: 90}
			maxPoint = Point{Lon: -180, Lat: -90}
			pol := make([]Point, len(poly[0]))
			for i, v := range poly[0] {
				var lon = v[0]
				var lat = v[1]
				pol[i].Lon = lon
				pol[i].Lat = lat
				updateMaxMinLatLon(&maxPoint, &minPoint, lat, lon)
			}
			g.Coordinates[j] = Coordinates{
				Polygon:  pol,
				MaxPoint: maxPoint,
				MinPoint: minPoint,
			}
		}
		return nil
	default:
		return nil
	}
}

func updateMaxMinLatLon(maxPoint, minPoint *Point, lat, lon float64) {
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
	for _, f := range fc.Features {
		for _, coord := range f.Geometry.Coordinates {
			if coord.MinPoint.Lat < lat && coord.MinPoint.Lon < lon && coord.MaxPoint.Lat > lat && coord.MaxPoint.Lon > lon {
				if coord.contains(Point{lon, lat}) {
					return f.Properties.Tzid
				}
			}
		}
	}
	return
}

func (g *Coordinates) contains(point Point) (contains bool) {
	if len(g.Polygon) < 3 {
		return
	}
	var p = g.Polygon
	contains = intersectsWithRaycast(point, p[len(p)-1], p[0])
	for i := 1; i < len(p); i++ {
		if intersectsWithRaycast(point, p[i-1], p[i]) {
			contains = !contains
			if contains == true {
				return
			}
		}
	}
	return
}

func intersectsWithRaycast(point, start, end Point) bool {
	return (start.Lon > point.Lon) != (end.Lon > point.Lon) &&
		point.Lat < (end.Lat-start.Lat)*(point.Lon-start.Lon)/(end.Lon-start.Lon)+start.Lat
}
