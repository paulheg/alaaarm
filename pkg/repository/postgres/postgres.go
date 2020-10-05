package postgres

import (
	"database/sql"
	"path"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3" // sqlite driver
	"github.com/paulheg/alaaarm/pkg/models"
	"github.com/paulheg/alaaarm/pkg/repository"
)

type sqlxdata struct {
	db     *sqlx.DB
	config *repository.Configuration
}

//New Repository
func New() repository.Repository {
	return &sqlxdata{}
}

func (r *sqlxdata) InitDatabase(config *repository.Configuration) error {

	r.config = config

	switch r.config.Database {
	case "sqlite":
		db, err := sqlx.Open("sqlite3", r.config.ConnectionString)
		if err != nil {
			return err
		}

		err = db.Ping()
		if err != nil {
			return err
		}

		r.db = db
	case "postgres":

	}

	return nil
}

func (r *sqlxdata) MigrateDatabase() error {

	version := struct {
		models.AutoIncrement
		UpdatedAt sql.NullTime `db:"updated_at"`
		Version   uint         `db:"version"`
	}{}

	err := r.db.Get(&version, `SELECT * FROM VERSION 
ORDER BY id DESC
LIMIT 1`)

	// run initial setup
	if err != nil && err != sql.ErrNoRows {
		err = r.runMigration(path.Join(r.config.MigrationDirectory, "setup.sql"))
		if err != nil {
			return err
		}
	}

	// run all migrations

	return nil
}

// Invitation Related
func (r *sqlxdata) CreateInvite(alert models.Alert) (models.Invite, error) {

	invite := models.NewInvite(&alert)

	result, err := r.db.NamedExec(`INSERT INTO INVITE (created_at, alert_id, token, one_time, expiration)
VALUES (:created_at, :alert_id, :token, :one_time, :expiration)`, &invite)

	if err != nil {
		return models.Invite{}, err
	}

	id, err := result.LastInsertId()

	if err != nil {
		return models.Invite{}, err
	}

	return r.GetInvite(uint(id))
}

func (r *sqlxdata) GetInviteByToken(token string) (models.Invite, error) {
	var invite models.Invite

	err := r.db.Get(&invite, `SELECT i.* FROM INVITE AS i
WHERE i.token = $1
AND i.deleted_at IS NULL;`, token)

	if err != nil {
		return invite, err
	}

	alert, err := r.GetAlert(invite.AlertID)
	if err != nil {
		return invite, err
	}

	invite.Alert = &alert

	return invite, err
}

func (r *sqlxdata) GetInvite(inviteID uint) (models.Invite, error) {
	var invite models.Invite

	err := r.db.Get(&invite, `SELECT i.* FROM INVITE AS i
WHERE i.id = $1
AND i.deleted_at IS NULL`, inviteID)

	if err != nil {
		return invite, err
	}

	alert, err := r.GetAlert(invite.AlertID)
	if err != nil {
		return invite, err
	}

	invite.Alert = &alert

	return invite, err
}

// Alert related
func (r *sqlxdata) CreateAlert(alert models.Alert) (models.Alert, error) {

	if alert.Owner.ID <= 0 {
		return models.Alert{}, repository.ErrMandatoryDataMissing
	}

	result, err := r.db.NamedExec(`INSERT INTO ALERT(created_at, name, description, owner_id, token)
		VALUES(:created_at, :name, :description, :owner_id, :token)`, &alert)

	if err != nil {
		return models.Alert{}, err
	}

	id, err := result.LastInsertId()

	if err != nil {
		return models.Alert{}, err
	}

	return r.GetAlert(uint(id))
}

func (r *sqlxdata) DeleteAlert(alert models.Alert) error {
	alert.DeletedAt.Scan(time.Now())

	result, err := r.db.Exec(`UPDATE ALERT
SET deleted_at = $2
WHERE ALERT.id = $1
AND ALERT.deleted_at IS NULL;`, alert.ID, alert.DeletedAt)

	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rows < 1 {
		return sql.ErrNoRows
	}

	return nil
}

func (r *sqlxdata) GetAlertReceiver(alert models.Alert) ([]models.User, error) {
	var receiver []models.User

	err := r.db.Select(&receiver, `SELECT u.* FROM USER AS u
INNER JOIN ALERT_RECEIVER AS ar ON ar.user_id = u.id
WHERE ar.alert_id = $1
AND u.deleted_at IS NULL;`, alert.ID)

	return receiver, err
}

func (r *sqlxdata) GetAlertByToken(token string) (models.Alert, error) {
	var alert models.Alert

	err := r.db.Get(&alert, `SELECT a.* FROM ALERT AS a
WHERE a.token = $1
AND a.deleted_at IS NULL;`, token)

	if err != nil {
		return alert, err
	}

	owner, err := r.GetUser(alert.OwnerID)
	if err != nil {
		return alert, err
	}

	alert.Owner = owner

	return alert, err
}

func (r *sqlxdata) GetAlert(id uint) (models.Alert, error) {
	var alert models.Alert

	err := r.db.Get(&alert, `SELECT a.* FROM ALERT AS a
WHERE a.id = $1
AND a.deleted_at IS NULL;`, id)

	if err != nil {
		return alert, err
	}

	owner, err := r.GetUser(alert.OwnerID)
	if err != nil {
		return alert, err
	}

	alert.Owner = owner

	return alert, err
}

func (r *sqlxdata) UpdateAlertToken(alert models.Alert) (models.Alert, error) {

	alert.ChangeToken()
	alert.UpdatedAt.Scan(time.Now())

	result, err := r.db.Exec(`UPDATE ALERT
SET token = $2, updated_at = $3
WHERE id = $1
AND deleted_at IS NULL;`, alert.ID, alert.Token, alert.UpdatedAt)

	if err != nil {
		return alert, err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return alert, err
	}

	if rows < 1 {
		return alert, sql.ErrNoRows
	}

	return alert, err
}

func (r *sqlxdata) GetUserAlerts(userID uint) ([]models.Alert, error) {
	var alerts []models.Alert

	err := r.db.Select(&alerts, `SELECT a.* FROM ALERT AS a
WHERE a.owner_id = $1
AND a.deleted_at IS NULL;`, userID)

	return alerts, err
}

func (r *sqlxdata) GetUserSubscribedAlerts(userID uint) ([]models.Alert, error) {
	var alerts []models.Alert

	err := r.db.Select(&alerts, `SELECT a.* FROM ALERT AS a
INNER JOIN ALERT_RECEIVER AS ar ON ar.alert_id = a.id
WHERE ar.user_id = $1
AND a.deleted_at IS NULL;`, userID)

	return alerts, err
}

func (r *sqlxdata) AddUserToAlert(alert models.Alert, user models.User) error {

	alertReceiver := models.NewAlertReceiver(&user, &alert)

	result, err := r.db.NamedExec(`INSERT INTO ALERT_RECEIVER(alert_id, user_id, created_at)
VALUES(:alert_id, :user_id, :created_at);`, &alertReceiver)

	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rows < 1 {
		return sql.ErrNoRows
	}

	return nil
}

func (r *sqlxdata) RemoveUserFromAlert(alert models.Alert, user models.User) error {
	// TODO: SQL update not working

	result, err := r.db.Exec(`UPDATE ALERT_RECEIVER
SET deleted_at = $3
WHERE ALERT_RECEIVER.user_id = $1 AND ALERT_RECEIVER.alert_id = $2
AND ALERT_RECEIVER.deleted_at IS NULL;`, user.ID, alert.ID, time.Now())

	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()

	if err != nil {
		return err
	}

	if rows < 1 {
		return sql.ErrNoRows
	}

	return nil
}

// User functions
func (r *sqlxdata) CreateUser(user models.User) (models.User, error) {

	user.CreatedAt.Time = time.Now()

	result, err := r.db.NamedExec(`INSERT INTO USER(created_at, username, telegram_id)
VALUES(:created_at, :username, :telegram_id);`, user)

	if err != nil {
		return models.User{}, err
	}

	userID, err := result.LastInsertId()

	if err != nil {
		return models.User{}, err
	}

	return r.GetUser(uint(userID))
}

func (r *sqlxdata) GetUser(userID uint) (models.User, error) {
	var user models.User

	err := r.db.Get(&user, `SELECT u.* FROM USER AS u
WHERE u.id = $1
AND u.deleted_at IS NULL`, userID)

	return user, err
}

func (r *sqlxdata) GetUserTelegram(telegramID int64) (models.User, error) {
	var user models.User

	err := r.db.Get(&user, `SELECT u.* FROM USER AS u
WHERE u.telegram_id = $1
AND u.deleted_at IS NULL`, telegramID)

	return user, err
}

func (r *sqlxdata) DeleteUser(userID uint) error {

	deletedAt := time.Now()

	result, err := r.db.Exec(`UPDATE USER
SET deleted_at = $2
WHERE u.id = $1
AND u.deleted_at IS NULL`, userID, deletedAt)

	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rows < 1 {
		return sql.ErrNoRows
	}

	return nil
}
