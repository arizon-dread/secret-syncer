package main

import (
	"log"

	"github.com/arizon-dread/secret-syncer/internal/conf"
	"github.com/arizon-dread/secret-syncer/pkg/handlers"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lmicroseconds)
	log.Printf("performing simple config validation")

	_, err := conf.GetConfig()
	if err != nil {
		log.Fatalf("%v", err)
	}

	log.Printf("syncing secrets")
	err = handlers.SyncMonitoredSecrets()
	if err != nil {
		log.Printf("failed syncing secrets completely, %v", err)
		return
	}
	log.Printf("secrets synced successfully")
}
