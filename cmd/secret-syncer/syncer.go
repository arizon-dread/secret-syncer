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
	results := handlers.SyncMonitoredSecrets()
	success := 0
	fail := 0
	for _, res := range results {
		if res.Success == true {
			success++
		} else {
			fail++
		}
	}
	if success == 0 {
		log.Printf("failed syncing secrets completely, %v", err)
	}
	if fail == 0 {
		log.Printf("secrets synced successfully")
	}
	log.Printf("Synced %d secret(s)  successfully and failed to sync %d. Total: %d", success, fail, len(results))
}
