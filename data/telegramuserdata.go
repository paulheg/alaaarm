package data

import (
	"database/sql"
	"log"
)

type telegramUserData struct {
	database *sql.DB
}

// NewTelegramUserData creates a new instance of a struct with the Userdata interface
func NewTelegramUserData(db *sql.DB) Userdata {
	// check db, maybe do initial setup

	return &telegramUserData{
		database: db,
	}
}

func (t *telegramUserData) UserExists(id int64) (bool, error) {
	err := t.database.QueryRow("select ID from USERDATA where USERDATA.ID = ? limit 1", id).Scan(&id)

	if err != nil {
		if err != sql.ErrNoRows {
			return false, err
		}

		return false, nil
	}

	return true, nil
}

func (t *telegramUserData) GetAllAuthorizedUsers() ([]int64, error) {
	var userIds []int64

	rows, err := t.database.Query("select ID from USERDATA where USERDATA.AUTHORIZED = 1")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	for rows.Next() {
		var ID int64
		err := rows.Scan(&ID)
		if err != nil {
			log.Fatal(err)
		}

		userIds = append(userIds, ID)
	}
	err = rows.Err()
	if err != nil {
		return userIds, err
	}

	return userIds, nil
}

func (t *telegramUserData) IsUserAuthorized(id int64) (bool, error) {
	row := t.database.QueryRow("select AUTHORIZED from USERDATA where USERDATA.ID = ? limit 1", id)

	var authorized bool = false

	err := row.Scan(&authorized)

	if err != nil && err != sql.ErrNoRows {
		return false, err
	}

	return authorized, nil
}

func (t *telegramUserData) DeleteUser(id int64) error {
	tx, err := t.database.Begin()
	if err != nil {
		return err
	}

	_, err = tx.Exec("delete from USERDATA where USERDATA.ID = ?", id)
	if err != nil {
		return err
	}

	tx.Commit()
	return nil
}

func (t *telegramUserData) AddUser(id int64, username string) error {

	b, err := t.UserExists(id)
	if err != nil {
		return err
	}
	if b {
		return ErrUserAlreadyExists
	}

	tx, err := t.database.Begin()
	if err != nil {
		return err
	}

	_, err = tx.Exec("insert into USERDATA(ID, USERNAME, AUTHORIZED) values(?, ?, ?)", id, username, false)
	if err != nil {
		return err
	}

	tx.Commit()
	return nil
}
