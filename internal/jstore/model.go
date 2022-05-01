package jstore

import "github.com/petuhovskiy/gobus/internal/cyprusbus"

type FileData struct {
	CityRoutesPage *cyprusbus.CityRoutesPage
	Routes         map[string]*cyprusbus.RouteFull
	Stops          map[int]cyprusbus.StopInfo
}
