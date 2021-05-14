![](https://raw.githubusercontent.com/catmullet/tz/master/assets/tz.png)
# Time Zone Lookup by Lat/Lon
### Description
This is a library for timezone lookup by latitude and longitude. It uses a bounding box to check the most likely polygons with a more accurate winding number algorithm. We need both speed and highly accurate results and this is the resulting project.
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

### Benchmarks

_Tests performed with cpu: Intel(R) Core(TM) i7-9750H CPU @ 2.60GHz_
|Test                                                    |Operations Ran      |ns per operation|
|:------------------------------------------------------ | ------------------:| --------------:|
|BenchmarkTimeZone/coordinates_5.261417_-3.925778-12     |    	  109633 op	  |    9837 ns/op|
|BenchmarkTimeZone/coordinates_-15.678889_34.973889-12   |    	   41721 op	  |   27305 ns/op|
|BenchmarkTimeZone/coordinates_-12.65945_18.25674-12     |    	   74956 op	  |   17318 ns/op|
|BenchmarkTimeZone/coordinates_41.8976_-87.6205-12       |    	   59955 op	  |   19741 ns/op|
|BenchmarkTimeZone/coordinates_47.6897_-122.4023-12      |    	   65709 op	  |   17656 ns/op|
|BenchmarkTimeZone/coordinates_42.7235_-73.6931-12       |    	   98896 op	  |   12355 ns/op|
|BenchmarkTimeZone/coordinates_42.5807_-83.0223-12       |    	   98629 op	  |   10762 ns/op|
|BenchmarkTimeZone/coordinates_36.8381_-84.85-12         |    	   37291 op	  |   31924 ns/op|
|BenchmarkTimeZone/coordinates_40.1674_-85.3583-12       |    	   57832 op	  |   19866 ns/op|
|BenchmarkTimeZone/coordinates_37.9643_-86.7453-12       |    	   61338 op	  |   19274 ns/op|
|BenchmarkTimeZone/coordinates_38.6043_-90.2417-12       |    	   61954 op	  |   20196 ns/op|
|BenchmarkTimeZone/coordinates_41.1591_-104.8261-12      |    	   86883 op	  |   14256 ns/op|
|BenchmarkTimeZone/coordinates_35.1991_-111.6348-12      |    	   61996 op	  |   18842 ns/op|
|BenchmarkTimeZone/coordinates_43.1432_-115.675-12       |    	   50565 op	  |   24131 ns/op|
|BenchmarkTimeZone/coordinates_47.5886_-122.3382-12      |    	   68821 op	  |   18421 ns/op|
|BenchmarkTimeZone/coordinates_58.3168_-134.4397-12      |    	   93099 op	  |   13280 ns/op|
|BenchmarkTimeZone/coordinates_21.4381_-158.0493-12      |    	12163994 op	  |   96.95 ns/op|
|BenchmarkTimeZone/coordinates_42.7_-80-12               |    	  135962 op	  |    8506 ns/op|
|BenchmarkTimeZone/coordinates_51.0036_-114.0161-12      |    	  323289 op	  |    3779 ns/op|
|BenchmarkTimeZone/coordinates_-16.4965_-68.1702-12      |    	   64844 op	  |   18756 ns/op|
|BenchmarkTimeZone/coordinates_-31.9369_115.8453-12      |    	  275347 op	  |    4256 ns/op|
|BenchmarkTimeZone/coordinates_42_-87.5-12               |    	   63606 op	  |   19925 ns/op|
|BenchmarkTimeZone/coordinates_41.8976_-87.6205#01-12    |    	   63562 op	  |   19546 ns/op|
|BenchmarkTimeZone/coordinates_47.6897_-122.4023#01-12   |    	   70030 op	  |   17950 ns/op|
|BenchmarkTimeZone/coordinates_42.7235_-73.6931#01-12    |    	   96730 op	  |   12206 ns/op|
|BenchmarkTimeZone/coordinates_42.5807_-83.0223#01-12    |    	  112562 op	  |   10770 ns/op|
|BenchmarkTimeZone/coordinates_36.8381_-84.85#01-12      |    	   38658 op	  |   33090 ns/op|
|BenchmarkTimeZone/coordinates_40.1674_-85.3583#01-12    |    	   58844 op	  |   20873 ns/op|
|BenchmarkTimeZone/coordinates_37.9643_-86.7453#01-12    |    	   61153 op	  |   20413 ns/op|
|BenchmarkTimeZone/coordinates_38.6043_-90.2417#01-12    |    	   50996 op	  |   21369 ns/op|
|BenchmarkTimeZone/coordinates_41.1591_-104.8261#01-12   |    	   76140 op	  |   15558 ns/op|
|BenchmarkTimeZone/coordinates_35.1991_-111.6348#01-12   |    	   60016 op	  |   19801 ns/op|
|BenchmarkTimeZone/coordinates_43.1432_-115.675#01-12    |    	   46855 op	  |   25990 ns/op|
|BenchmarkTimeZone/coordinates_47.5886_-122.3382#01-12   |    	   53744 op	  |   23905 ns/op|
|BenchmarkTimeZone/coordinates_58.3168_-134.4397#01-12   |    	   66007 op	  |   16711 ns/op|
|BenchmarkTimeZone/coordinates_21.4381_-158.0493#01-12   |    	 8266512 op	  |   123.8 ns/op|
|BenchmarkTimeZone/coordinates_42.7_-80#01-12            |    	  112492 op	  |   10378 ns/op|
|BenchmarkTimeZone/coordinates_51.0036_-114.0161#01-12   |    	  294398 op	  |    4424 ns/op|
|BenchmarkTimeZone/coordinates_-16.4965_-68.1702#01-12   |    	   60153 op	  |   19283 ns/op|
|BenchmarkTimeZone/coordinates_-31.9369_115.8453#01-12   |    	  263713 op	  |    4694 ns/op|
|BenchmarkTimeZone/coordinates_42_-87.5#01-12            |    	   61891 op	  |   20164 ns/op|
|BenchmarkTimeZone/coordinates_41.8976_-87.6205#02-12    |    	   55926 op	  |   20487 ns/op|
|BenchmarkTimeZone/coordinates_47.6897_-122.4023#02-12   |    	   67516 op	  |   19023 ns/op|
|BenchmarkTimeZone/coordinates_42.7235_-73.6931#02-12    |    	   92011 op	  |   13674 ns/op|
|BenchmarkTimeZone/coordinates_42.5807_-83.0223#02-12    |    	   97562 op	  |   12193 ns/op|
|BenchmarkTimeZone/coordinates_36.8381_-84.85#02-12      |    	   36657 op	  |   32004 ns/op|
|BenchmarkTimeZone/coordinates_40.1674_-85.3583#02-12    |    	   59169 op	  |   20827 ns/op|
|BenchmarkTimeZone/coordinates_37.9643_-86.7453#02-12    |    	   59452 op	  |   20075 ns/op|
|BenchmarkTimeZone/coordinates_38.6043_-90.2417#02-12    |    	   56371 op	  |   19961 ns/op|
|BenchmarkTimeZone/coordinates_41.1591_-104.8261#02-12   |    	   79848 op	  |   14434 ns/op|
|BenchmarkTimeZone/coordinates_35.1991_-111.6348#02-12   |    	   59436 op	  |   18598 ns/op|
|BenchmarkTimeZone/coordinates_43.1432_-115.675#02-12    |    	   49734 op	  |   24137 ns/op|
|BenchmarkTimeZone/coordinates_47.5886_-122.3382#02-12   |    	   66818 op	  |   18188 ns/op|
|BenchmarkTimeZone/coordinates_58.3168_-134.4397#02-12   |    	   94588 op	  |   12781 ns/op|
|BenchmarkTimeZone/coordinates_21.4381_-158.0493#02-12   |    	12312768 op	  |   99.03 ns/op|
|BenchmarkTimeZone/coordinates_42.7_-80#02-12            |    	  129206 op	  |    8987 ns/op|
|BenchmarkTimeZone/coordinates_51.0036_-114.0161#02-12   |    	  310020 op	  |    4098 ns/op|
|BenchmarkTimeZone/coordinates_-16.4965_-68.1702#02-12   |    	   61988 op	  |   21531 ns/op|
|BenchmarkTimeZone/coordinates_-31.9369_115.8453#02-12   |    	  167610 op	  |    7179 ns/op|
|BenchmarkTimeZone/coordinates_42_-87.5#02-12            |    	   51849 op	  |   20606 ns/op|
