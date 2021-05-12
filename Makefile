profile:
	go test -cpuprofile=cpu.out -benchmem -memprofile=mem.out -bench=. -v
update-geo:
	rm -f 'cache/tzdata.snappy' && curl -s https://api.github.com/repos/evansiroky/timezone-boundary-builder/releases/latest | grep -E 'browser_download_url' | grep 'timezones-with-oceans.geojson.zip' | cut -d '"' -f 4 | wget -qi -