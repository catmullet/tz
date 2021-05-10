// +build ignore
/*
 Sample Output:
	building geojson data in memory...
	finished building
	parsing...
	[Africa/Abidjan Africa/Blantyre Africa/Luanda America/Chicago America/Los_Angeles America/New_York America/Detroit America/Kentucky/Monticello America/Indiana/Indianapolis America/Chicago America/Chicago America/Chicago America/Denver America/Barbados Asia/Brunei America/Barbados Africa/Libreville America/Chicago America/Denver America/Denver America/Denver Africa/Ouagadougou America/Paramaribo]
	86926 ns/parse 1 ms/23
	finished
*/
package main

import (
	"fmt"
	"github.com/catmullet/tz"
	"log"
	"time"
)

type timeZone struct {
	Lat float64
	Lon float64
}

var timezones = []timeZone{
	{Lat: 5.261417, Lon: -3.925778},
	{Lat: -15.678889, Lon: 34.973889},
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
	{Lat: 13.11257, Lon: -59.61126},
	{Lat: 4.89192, Lon: 114.91151},
	{Lat: 13.27927, Lon: -59.64895},
	{Lat: 0.38483, Lon: 9.47567},
	{Lat: 37.12188, Lon: -92.26608},
	{Lat: 33.3219409, Lon: -111.6862607},
	{Lat: 40.215389, Lon: -111.660211},
	{Lat: 41.74568, Lon: -111.8191},
	{Lat: 11.21169, Lon: -4.30724},
	{Lat: 5.78378, Lon: -55.18182},
}

func main() {
	var parsedTimeZoneList = make([]string, len(timezones))

	log.Println("building geojson data in memory...")
	tzLookup, err := tz.NewGeoJsonTimeZoneLookup("timezones-with-oceans.geojson.zip")
	if err != nil {
		log.Fatalln("Failed to initialize tz", err)
	}
	log.Println("finished building")

	log.Println("parsing...")
	start := time.Now()
	for i := range timezones {
		parsedTimeZoneList[i] = tzLookup.TimeZone(timezones[i].Lat, timezones[i].Lon)
	}
	totalTime := time.Since(start)

	log.Println(parsedTimeZoneList)
	log.Println(totalTime.Nanoseconds()/int64(len(timezones)), "ns/parse", totalTime.Milliseconds(),
		fmt.Sprintf("ms/%d", len(timezones)))
	log.Println("finished")
}
