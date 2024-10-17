package migration

import (
	"database/sql"
)

type Migration interface {
	Version() int
	Up(database *sql.DB) error
	Down(database *sql.DB) error
}
