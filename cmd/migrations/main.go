package main

import (
	"os"

	logger "github.com/lob/logger-go"
	"github.com/lob/pharos/pkg/pharos-api-server/config"
	"github.com/lob/pharos/pkg/pharos-api-server/database"
	migrations "github.com/robinjoseph08/go-pg-migrations"
)

const directory = "./cmd/migrations"

func main() {
	log := logger.New()

	cfg := config.New()
	db, err := database.New(cfg)
	if err != nil {
		log.Err(err).Fatal("database error")
	}

	err = migrations.Run(db, directory, os.Args)
	if err != nil {
		log.Err(err).Fatal("migration error")
	}
}
