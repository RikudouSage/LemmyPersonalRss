package migration

import (
	"database/sql"
)

type Version0 struct {
}

func (receiver *Version0) Version() int {
	return 0
}

func (receiver *Version0) Up(database *sql.DB) (err error) {
	_, err = database.Exec("create table migrations (version string not null, created_at text not null default current_timestamp)")
	return
}

func (receiver *Version0) Down(database *sql.DB) (err error) {
	_, err = database.Exec("drop table migrations")
	return
}

func init() {
	GetManager().migrations = append(globalManager.migrations, &Version0{})
}
