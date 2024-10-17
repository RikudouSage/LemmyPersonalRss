package migration

import (
	"database/sql"
	"slices"
	"strings"
)

var globalManager *Manager

type Manager struct {
	migrations []Migration
}

func GetManager() *Manager {
	if globalManager == nil {
		globalManager = &Manager{
			migrations: make([]Migration, 0),
		}
	}
	return globalManager
}

func (receiver *Manager) RegisterMigration(migration Migration, db *sql.DB) (err error) {
	_, err = db.Exec("insert into migrations (version) values (?)", migration.Version())
	return
}

func (receiver *Manager) Migrate(db *sql.DB) error {
	migrations := receiver.getMigrations()
	var currentVersion int

	rows, err := db.Query("select max(version) from migrations")

	if err != nil {
		if !strings.Contains(err.Error(), "no such table") {
			return err
		}
		currentVersion = -1
	} else {
		if !rows.Next() {
			currentVersion = -1
		} else {
			err = rows.Scan(&currentVersion)
			if err != nil {
				return err
			}
		}
		err = rows.Close()
		if err != nil {
			return err
		}
	}

	for _, migration := range migrations {
		if migration.Version() <= currentVersion {
			continue
		}

		err = migration.Up(db)
		if err != nil {
			return err
		}
		err = receiver.RegisterMigration(migration, db)
		if err != nil {
			return err
		}
	}

	return nil
}

func (receiver *Manager) getMigrations() []Migration {
	migrations := slices.Clone(receiver.migrations)
	slices.SortFunc(migrations, func(a, b Migration) int {
		if a.Version() == b.Version() {
			return 0
		}
		if a.Version() > b.Version() {
			return 1
		}
		return -1
	})

	return migrations
}
