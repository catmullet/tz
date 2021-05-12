![](https://raw.githubusercontent.com/catmullet/tz/master/assets/tz.png)
# Time Zone Lookup by Lat/Lon
### Description
This Repo for timezone lookup by latitude and longitude. It is a mash up of the two popular Go libraries
* [go-tz](https://github.com/ugjka/go-tz)
* [timezoneLookup](https://github.com/evanoberholster/timezoneLookup)
### Goals of this project?
1. Speed
2. Accuracy.
3. Easy.
4. Concurrent safe.
5. Easy updating of zip file locally.
### Simple
Grab the tz library.
```zsh
    go get github.com/catmullet/tz
```
Update the geo json data from [evansiroky/timezone-boundary-builder](https://github.com/evansiroky/timezone-boundary-builder). (Likely not necessary.)
```shell
    make updategeo
```
Initialize the GeoJson Time Zone Lookup.
```go 
    lookup, err := tz.NewGeoJsonTimeZoneLookup("timezones-with-oceans.geojson.zip")
```
Parse by Lat Lon and return string value of time zone.
```go 
    timeZone := lookup.TimeZone(5.261417, -3.925778)
```
Parse for ```time.Location```.
```go
    var ti = time.Now()
    
    var loc = lookup.Location(5.261417, -3.925778)
    
    // Set time to location from lookup, easy!
    ti.In(loc).Zone()
```