package main

import (
	"github.com/petuhovskiy/gobus/internal/conf"
	"github.com/petuhovskiy/gobus/internal/cyprusbus"
	"github.com/petuhovskiy/gobus/internal/jstore"
	log "github.com/sirupsen/logrus"
)

func main() {
	cfg, err := conf.ParseEnv()
	if err != nil {
		log.WithError(err).Fatal("failed to parse config from env")
	}

	store, err := jstore.NewStore(cfg.DataStore)
	if err != nil {
		log.WithError(err).Fatal("failed to create store")
	}

	cli := cyprusbus.NewClient()
	_ = cli

	data, err := store.LoadLatest()
	if err != nil {
		log.WithError(err).Fatal("failed to load latest data")
	}

	// TODO: something here

	err = store.Save(data)
	if err != nil {
		log.WithError(err).Fatal("failed to save data")
	}
}
