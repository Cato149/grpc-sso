package main

import (
	"flag"
	"fmt"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/sqlite3"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

// go run ./cmd/migrations --storage-path=./storage/sso.db --migrations-path=./migrations
func main() {
	var storagePath, migrationsPath, targetTable string

	flag.StringVar(&storagePath, "storage-path", "", "path to migrations storage")
	flag.StringVar(&migrationsPath, "migrations-path", "", "path to migrations storage")
	flag.StringVar(&targetTable, "target-table", "migrations", "target table name")
	flag.Parse()

	if storagePath == "" || migrationsPath == "" {
		panic("storagePath, migrationsPath, and target table must be specified")
	}

	m, err := migrate.New(
		"file://"+migrationsPath,
		fmt.Sprintf("sqlite3://%s?x-migration-table=%s", storagePath, targetTable),
	)
	if err != nil {
		panic(err)
	}

	if err := m.Up(); err != nil {
		if err != migrate.ErrNoChange {
			fmt.Println("Nothing to migrate")

			return
		}

		panic(err)
	}

	fmt.Println("Successfully migrated!")
}
