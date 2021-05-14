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
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"time"
)

var currentDirectory = func() string {
	var _, currentFilePath, _, _ = runtime.Caller(0)
	return strings.ReplaceAll(currentFilePath, "/geojson.go", "")
}()

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

	var timeZoneFile = filepath.Join(currentDirectory, "timezones-with-oceans.geojson.zip")

	if len(geoJsonFile) == 0 {
		geoJsonFile = os.Getenv("GEO_JSON_FILE")
	}
	if len(geoJsonFile) == 0 {
		logger.Println("no geo time zone file specified, using default")
		geoJsonFile = timeZoneFile
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
	cache, err := os.Open(filepath.Join(currentDirectory, "tzdata.snappy"))
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
	cache, err := os.Create(filepath.Join(currentDirectory, "tzdata.snappy"))
	if err != nil {
		return err
	}
	defer cache.Close()
	snp := snappy.NewBufferedWriter(cache)
	enc := gob.NewEncoder(snp)
	if err := enc.Encode(fc); err != nil {
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
	var polygon = c.Polygon
	if windingNumber(point.Lat, point.Lon, polygon) == 0 {
		return false
	}
	return true
}

func windingNumber(lat, lon float64, polygon []Point) int {
	if len(polygon) < 3 {
		return 0
	}

	var wn = 0
	var edgeCount = len(polygon) - 5

	for i, j := 0, 5; i < edgeCount; i, j = i+5, j+5 {
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
