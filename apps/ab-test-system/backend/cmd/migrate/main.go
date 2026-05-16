package main

import (
	"errors"
	"flag"
	"log"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func main() {
	var (
		dsn       = flag.String("dsn", "", "PostgreSQL DSN (required)")
		direction = flag.String("direction", "up", "Migration direction: up or down")
		steps     = flag.Int("steps", 0, "Number of steps (0 = all)")
	)
	flag.Parse()

	if *dsn == "" {
		log.Fatal("-dsn is required")
	}

	m, err := migrate.New("file://migrations", *dsn)
	if err != nil {
		log.Fatalf("creating migrator: %v", err)
	}
	defer func() {
		srcErr, dbErr := m.Close()
		if srcErr != nil {
			log.Printf("closing migrate source: %v", srcErr)
		}
		if dbErr != nil {
			log.Printf("closing migrate db: %v", dbErr)
		}
	}()

	switch *direction {
	case "up":
		if *steps > 0 {
			err = m.Steps(*steps)
		} else {
			err = m.Up()
		}
	case "down":
		if *steps > 0 {
			err = m.Steps(-(*steps))
		} else {
			err = m.Down()
		}
	default:
		log.Fatalf("unknown direction %q, use 'up' or 'down'", *direction)
	}

	if err != nil && !errors.Is(err, migrate.ErrNoChange) {
		log.Fatalf("migration failed: %v", err)
	}

	log.Println("migration complete")
}
