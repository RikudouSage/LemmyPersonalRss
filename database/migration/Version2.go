package migration

import (
	"database/sql"
)

type Version2 struct {
}

func (receiver *Version2) Version() int {
	return 2
}

func (receiver *Version2) Up(database *sql.DB) (err error) {
	_, err = database.Exec("alter table users add column image_url text null default null")
	return
}

func (receiver *Version2) Down(database *sql.DB) (err error) {
	_, err = database.Exec("alter table users drop column image_url")
	return
}

func init() {
	GetManager().migrations = append(GetManager().migrations, &Version2{})
}
