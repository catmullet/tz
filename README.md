![](https://raw.githubusercontent.com/catmullet/tz/master/assets/tz.png)
# Time Zone Lookup by Lat/Lon
### Description
This is a library for timezone lookup by latitude and longitude. It uses a bounding box to check the most likely polygons with a more accurate raycast algorithm. Seeing on avg ~31000 queries/second and worst case ~10000 queries/second. We need both speed and highly accurate results and this is the resulting project.
### Goals of this project?
1. Speed
2. Accuracy.
3. Easy.
4. Concurrent safe.
5. Easy updating of zip file locally.
## Simple
### 1. Grab the tz library.
```zsh
    go get github.com/catmullet/tz
```
### 2. Update the geo json data from [evansiroky/timezone-boundary-builder](https://github.com/evansiroky/timezone-boundary-builder)*. (Likely not necessary.)
```shell
    make updategeo
```
_*The script will delete the current_ ```tzdata.snappy``` _compiled TimeZone file and Download a new geojson zip file.  This new geojson file will be loaded and compiled on the next run or by running unit tests. A new ```tzdata.snappy``` file will be created._

### 3. Initialize the GeoJson Time Zone Lookup.
```go 
    lookup, err := tz.NewGeoJsonTimeZoneLookup("timezones-with-oceans.geojson.zip")
    // OR
    // export GEO_JSON_FILE="path/timezones-with-oceans.geojson.zip" or will pull from root directory
    lookup, err := tz.NewGeoJsonTimeZoneLookup("")
    // OR Get info from logging
    lookup, err := tz.NewGeoJsonTimeZoneLookup("", os.Stderr)
```
### 4. Parse
Parse by Lat Lon and return string value of time zone.
```go 
    timeZone := lookup.TimeZone(5.261417, -3.925778)
```
Parse for ```time.Location```.
```go
    var ti = time.Now()
    
    var loc, _ = lookup.Location(5.261417, -3.925778)
    
    // Set time to location from lookup, easy!
    ti.In(loc).Zone()
```
