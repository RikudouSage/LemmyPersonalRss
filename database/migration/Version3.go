package migration

import (
	"database/sql"
)

type Version3 struct {
}

func (receiver *Version3) Version() int {
	return 3
}

func (receiver *Version3) Up(database *sql.DB) (err error) {
	_, err = database.Exec("alter table users add column instance text null default null")
	return
}

func (receiver *Version3) Down(database *sql.DB) (err error) {
	_, err = database.Exec("alter table users drop column instance")
	return
}

func init() {
	GetManager().migrations = append(GetManager().migrations, &Version3{})
}
