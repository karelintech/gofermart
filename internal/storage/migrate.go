package storage

import (
	"fmt"
	"path/filepath"

	"github.com/golang-migrate/migrate/v4"
	// Для выплнения init()
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

// RunMigrations - запуск миграций при запуске приложения
func RunMigrations(dbURL, migrationsPath string) error {
	if dbURL == "" {
		return fmt.Errorf("database URL is empty")
	}

	abcPath, _ := filepath.Abs(migrationsPath)
	pathToMigrate := "file://" + abcPath + "/migrations"
	m, err := migrate.New(pathToMigrate, dbURL)
	if err != nil {
		return err
	}
	defer m.Close()

	if err = m.Up(); err != nil && err != migrate.ErrNoChange {
		return err
	}

	return nil
}
