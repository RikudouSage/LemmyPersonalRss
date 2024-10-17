package migration

import (
	"database/sql"
)

type Version1 struct {
}

func (receiver *Version1) Version() int {
	return 1
}

func (receiver *Version1) Up(database *sql.DB) (err error) {
	_, err = database.Exec("create table users (id int primary key not null, hash text, jwt text, username text)")
	return
}

func (receiver *Version1) Down(database *sql.DB) (err error) {
	_, err = database.Exec("drop table users")
	return
}

func init() {
	GetManager().migrations = append(GetManager().migrations, &Version1{})
}
