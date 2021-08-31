package tz

type geoStorage interface {
	StoreFile(filename string, obj interface{}) error
	LoadFile(filename string, obj interface{}) ([]byte, error)
}
