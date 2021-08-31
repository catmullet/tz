package tz

import (
	"embed"
	"encoding/json"
	"github.com/golang/snappy"
	"log"
	"strings"
)

type LocalGeoStorage struct {
	efs embed.FS
}

func newLocalGeoStorage(fs embed.FS) geoStorage {
	return &LocalGeoStorage{
		efs: fs,
	}
}

func (lgs LocalGeoStorage) StoreFile(_ string, _ interface{}) error {
	log.Fatalln("local geo storage: store file not used")
	return nil
}

func (lgs LocalGeoStorage) LoadFile(filename string, obj interface{}) ([]byte, error) {
	var decodedJson, err = lgs.efs.ReadFile(filename)
	if strings.Contains(filename, "snappy") {
		decodedJson, err = snappy.Decode(nil, decodedJson)
		if err != nil {
			return decodedJson, err
		}
	}
	if err != nil {
		return decodedJson, err
	}
	return decodedJson, json.Unmarshal(decodedJson, &obj)
}
