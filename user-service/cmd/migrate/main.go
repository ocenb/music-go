package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/joho/godotenv"
	"github.com/ocenb/music-go/user-service/internal/utils"
)

func main() {
	var migrationsPath, migrationsTable, dbURL string
	var down int

	flag.StringVar(&migrationsPath, "migrations-path", "./migrations", "path to migrations")
	flag.StringVar(&migrationsTable, "migrations-table", "migrations", "name of migrations table")
	flag.StringVar(&dbURL, "db-url", "", "database URL")
	flag.IntVar(&down, "down", 0, "number of migrations to roll back (0 means no rollback)")
	flag.Parse()

	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	if dbURL == "" {
		dbURL = utils.GetDBUrl(os.Getenv("POSTGRES_HOST"), os.Getenv("POSTGRES_PORT"), os.Getenv("POSTGRES_USER"), os.Getenv("POSTGRES_PASSWORD"), os.Getenv("POSTGRES_DB"), os.Getenv("POSTGRES_SSL_MODE"))
		if dbURL == "" {
			log.Fatal("db-url or DATABASE_URL environment variable is required")
		}
	}

	fmt.Println("dbURL", dbURL)
	m, err := migrate.New(
		"file://"+migrationsPath,
		dbURL,
	)
	if err != nil {
		log.Fatalf("migrate.New error: %v", err)
	}

	m.Log = &Log{verbose: true}

	if down > 0 {
		if err := m.Steps(-down); err != nil {
			if errors.Is(err, migrate.ErrNoChange) {
				fmt.Println("no migrations to roll back")
				return
			}
			log.Fatalf("m.Down error: %v", err)
		}
		fmt.Printf("rolled back %d migrations\n", down)
		return
	}

	if err := m.Up(); err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			fmt.Println("no migrations to apply")
			return
		}
		log.Fatalf("m.Up error: %v", err)
	}

	fmt.Println("migrations applied")
}

type Log struct {
	verbose bool
}

func (l *Log) Printf(format string, v ...interface{}) {
	if l.verbose {
		fmt.Printf(format, v...)
	}
}

func (l *Log) Verbose() bool {
	return l.verbose
}
