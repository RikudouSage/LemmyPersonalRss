package database

import (
	"LemmyPersonalRss/database/migration"
	"LemmyPersonalRss/dto"
	"database/sql"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
)

type SqliteDatabase struct {
	db *sql.DB
}

func NewSqliteDatabase(path string, migrationManager *migration.Manager) (*SqliteDatabase, error) {
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		return nil, err
	}

	err = migrationManager.Migrate(db)
	if err != nil {
		return nil, err
	}

	return &SqliteDatabase{db: db}, nil
}

func (receiver *SqliteDatabase) FindByUserId(userId int) *dto.AppUser {
	receiver.validate()

	rows, err := receiver.db.Query("SELECT * FROM users WHERE id = ?", userId)
	if err != nil {
		fmt.Println(err)
		return nil
	}
	defer rows.Close()

	if !rows.Next() {
		fmt.Println("No user found")
		return nil
	}

	user := &dto.AppUser{}
	err = rows.Scan(&user.Id, &user.Hash, &user.Jwt, &user.Username)
	if err != nil {
		fmt.Println(err)
		return nil
	}

	return user
}

func (receiver *SqliteDatabase) StoreUser(user *dto.AppUser) error {
	receiver.validate()

	_, err := receiver.db.Exec(
		"INSERT INTO users (id, hash, jwt, username) VALUES (?, ?, ?, ?)",
		user.Id,
		user.Hash,
		user.Jwt,
		user.Username,
	)

	if err != nil {
		fmt.Println(err)
	}

	return err
}

func (receiver *SqliteDatabase) FindByHash(userHash string) *dto.AppUser {
	receiver.validate()

	rows, err := receiver.db.Query("SELECT * FROM users WHERE hash = ?", userHash)
	if err != nil {
		fmt.Println(err)
		return nil
	}
	defer rows.Close()

	if !rows.Next() {
		return nil
	}

	user := &dto.AppUser{}
	err = rows.Scan(&user.Id, &user.Hash, &user.Jwt, &user.Username)
	if err != nil {
		fmt.Println(err)
		return nil
	}

	return user
}

func (receiver *SqliteDatabase) validate() {
	if receiver.db == nil {
		panic("Please use NewSqliteDatabase(path) to create an instance of SqliteDatabase")
	}
}
