package tz

import (
	"archive/zip"
	"bytes"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"github.com/golang/snappy"
	"io"
	"log"
	"math"
	"os"
	"path/filepath"
	"sort"
	"time"
)

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

func NewGeoJsonTimeZoneLookup(geoJsonFile string, logOutput ...io.Writer) (TimeZoneLookup, error) {

	logger := log.New(io.MultiWriter(logOutput...), "tz", log.Lshortfile)
	logger.Println("initializing...")

	const timzonesFile = "timezones-with-oceans.geojson.zip"

	if len(geoJsonFile) == 0 {
		geoJsonFile = os.Getenv("GEO_JSON_FILE")
	}
	if len(geoJsonFile) == 0 {
		logger.Println("no geo time zone file specified, using default")
		geoJsonFile = timzonesFile
	}

	var fc = &TimeZoneCollection{
		Features: make([]*Feature, 500),
	}

	if err := findCachedModel(fc); err == nil {
		logger.Println("cached model found")
		return fc, nil
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

	for _, feat := range fc.Features {
		f := feat
		sort.SliceStable(f.Geometry.Coordinates, func(i, j int) bool {
			return f.Geometry.Coordinates[i].MinPoint.Lon <= f.Geometry.Coordinates[j].MinPoint.Lon
		})
	}

	sort.SliceStable(fc.Features, func(i, j int) bool {
		return fc.Features[i].Geometry.MinPoint.Lon <= fc.Features[j].Geometry.MinPoint.Lon
	})

	logger.Println("finished")
	return fc, createCachedModel(fc)
}

func findCachedModel(fc *TimeZoneCollection) error {
	cache, err := os.Open(filepath.Join("cache", "tzdata.snappy"))
	if err != nil {
		return err
	}
	defer cache.Close()
	snp := snappy.NewReader(cache)
	dec := gob.NewDecoder(snp)
	if err := dec.Decode(fc); err != nil && err != io.EOF {
		return err
	}
	return nil
}

func createCachedModel(fc *TimeZoneCollection) error {
	cache, err := os.Create(filepath.Join("cache", "tzdata.snappy"))
	if err != nil {
		return err
	}
	defer cache.Close()
	snp := snappy.NewBufferedWriter(cache)
	enc := gob.NewEncoder(snp)
	if err := enc.Encode(fc); err != nil {
		cache.Close()
		return err
	}

	return nil
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

func (fc TimeZoneCollection) TimeZone(lat, lon float64) (tz string) {
	for _, feat := range fc.Features {
		f := feat
		tzString := f.Properties.Tzid
		if f.Geometry.MinPoint.Lat < lat &&
			f.Geometry.MinPoint.Lon < lon &&
			f.Geometry.MaxPoint.Lat > lat &&
			f.Geometry.MaxPoint.Lon > lon {
			for _, c := range f.Geometry.Coordinates {
				coord := c
				if coord.MinPoint.Lat < lat &&
					coord.MinPoint.Lon < lon &&
					coord.MaxPoint.Lat > lat &&
					coord.MaxPoint.Lon > lon {
					if coord.contains(Point{lon, lat}) {
						return tzString
					}
				}
			}
		}
	}
	return
}

func (fc TimeZoneCollection) Location(lat, lon float64) (loc *time.Location, err error) {
	return time.LoadLocation(fc.TimeZone(lat, lon))
}

func (c Coordinates) contains(point Point) bool {
	const tolerance = 5
	var polyLen = len(c.Polygon)
	if polyLen < 3 {
		return false
	}

	start := polyLen - tolerance
	end := 0

	contains := rayCast(point, c.Polygon[start], c.Polygon[end])
	for i, j := tolerance, tolerance; i < polyLen; i, j = i+tolerance, i {
		if rayCast(point, c.Polygon[j], c.Polygon[i]) {
			contains = !contains
		}
	}
	return contains
}

func rayCast(point, start, end Point) bool {
	var pLat, pLon, startLat, startLon, endLat, endLon = point.Lat, point.Lon, start.Lat, start.Lon, end.Lat, end.Lon

	if startLat > endLat {
		startLat, startLon, endLat, endLon = endLat, endLon, startLat, startLon
	}
	for pLat == startLat || pLat == endLat {
		pLat = math.Nextafter(pLat, math.Inf(1))
	}
	if pLat < startLat || pLat > endLat {
		return false
	}
	if startLon > endLon {
		if pLon > startLon {
			return false
		}
		if pLon < endLon {
			return true
		}
	} else {
		if pLon > endLon {
			return false
		}
		if pLon < startLon {
			return true
		}
	}
	return (pLat-startLat)/(pLon-startLon) >= (endLat-startLat)/(endLon-startLon)
}
