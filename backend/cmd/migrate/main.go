package main

import (
	"flag"
	"log"
	"os"

	"was/backend/internal/config"
	"was/backend/migrations"
)

func main() {
	direction := flag.String("direction", "up", "Migration direction: up or down")
	flag.Parse()

	cfg := config.Load()

	log.Printf("[migrate] Executing migration: %s on %s", *direction, cfg.DatabaseURL)

	var err error
	if *direction == "down" {
		err = migrations.RollbackMigration(cfg.DatabaseURL)
	} else {
		err = migrations.RunMigrations(cfg.DatabaseURL)
	}

	if err != nil {
		log.Fatalf("[migrate] Migration failed: %v", err)
	}

	log.Println("[migrate] Migration completed successfully.")
	os.Exit(0)
}

