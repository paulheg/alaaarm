package repository

import (
	"database/sql"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"           // postgres driver
	_ "github.com/mattn/go-sqlite3" // sqlite driver
	"github.com/paulheg/alaaarm/pkg/migration"
	"github.com/paulheg/alaaarm/pkg/models"
	"github.com/sirupsen/logrus"
)

type postgres struct {
	db     *sqlx.DB
	config *Configuration
	log    *logrus.Logger
}

// New Repository
func NewPostgres(log *logrus.Logger) Repository {
	return &postgres{
		log: log,
	}
}

func (p *postgres) MigrateDatabase() error {
	migrator := migration.New(p.db.DB, p, p.config.MigrationDirectory, p.log)

	return migrator.Migrate()
}

func (p *postgres) InitDatabase(config *Configuration) error {

	p.config = config

	db, err := sqlx.Open("postgres", p.config.ConnectionString)
	if err != nil {
		return err
	}

	err = db.Ping()
	if err != nil {
		return err
	}

	p.db = db

	return nil
}

// Invitation Related
func (p *postgres) CreateInvite(alert models.Alert) (models.Invite, error) {
	var id uint

	invite := models.NewInvite(&alert)

	err := p.db.Get(&id, `INSERT INTO "INVITE" (created_at, alert_id, token, one_time, expiration)
VALUES ($1, $2, $3, $4, $5)
RETURNING id`, invite.CreatedAt, invite.AlertID, invite.Token, invite.OneTime, invite.Expiration)

	if err != nil {
		return models.Invite{}, err
	}

	return p.GetInvite(id)
}

func (p *postgres) GetInviteByToken(token string) (models.Invite, error) {
	var invite models.Invite

	err := p.db.Get(&invite, `SELECT i.* FROM "INVITE" AS i
WHERE i.token = $1
AND i.deleted_at IS NULL;`, token)

	if err != nil {
		return invite, err
	}

	alert, err := p.GetAlert(invite.AlertID)
	if err != nil {
		return invite, err
	}

	invite.Alert = &alert

	return invite, err
}

func (p *postgres) GetInvite(inviteID uint) (models.Invite, error) {
	var invite models.Invite

	err := p.db.Get(&invite, `SELECT i.* FROM "INVITE" AS i
WHERE i.id = $1
AND i.deleted_at IS NULL`, inviteID)

	if err != nil {
		return invite, err
	}

	alert, err := p.GetAlert(invite.AlertID)
	if err != nil {
		return invite, err
	}

	invite.Alert = &alert

	return invite, err
}

func (p *postgres) GetInviteByAlertID(alertID uint) (models.Invite, error) {
	var invite models.Invite

	err := p.db.Get(&invite, `SELECT i.* FROM "INVITE" AS i
WHERE i.alert_id = $1
AND i.deleted_at IS NULL`, alertID)

	return invite, err
}

func (r *postgres) DeleteInvite(inviteID uint) error {

	result, err := r.db.Exec(`UPDATE "INVITE" AS i
SET deleted_at = $1
WHERE i.id = $2
AND i.deleted_at IS NULL`, time.Now(), inviteID)

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

// Alert related
func (p *postgres) CreateAlert(alert models.Alert) (models.Alert, error) {

	if alert.Owner.ID <= 0 {
		return models.Alert{}, ErrMandatoryDataMissing
	}

	var id uint

	err := p.db.Get(&id, `INSERT INTO "ALERT" (created_at, name, description, owner_id, token)
VALUES($1, $2, $3, $4, $5)
RETURNING id`, alert.CreatedAt, alert.Name, alert.Description, alert.OwnerID, alert.Token)

	if err != nil {
		return models.Alert{}, err
	}

	return p.GetAlert(id)
}

func (p *postgres) DeleteAlert(alert models.Alert) error {

	result, err := p.db.Exec(`UPDATE "ALERT" AS a
SET deleted_at = $1
WHERE a.id = $2
AND a.deleted_at IS NULL;`, time.Now(), alert.ID)

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

func (p *postgres) GetAlertReceiver(alert models.Alert) ([]models.User, error) {
	var receiver []models.User

	err := p.db.Select(&receiver, `SELECT u.* FROM "USER" AS u
INNER JOIN "ALERT_RECEIVER" AS ar ON ar.user_id = u.id
WHERE ar.alert_id = $1
AND u.deleted_at IS NULL AND ar.deleted_at IS NULL;`, alert.ID)

	return receiver, err
}

func (p *postgres) GetSubscriberCount(alert models.Alert) (int64, error) {

	var count int64

	err := p.db.Get(&count, `SELECT COUNT(u.*) FROM "USER" AS u
INNER JOIN "ALERT_RECEIVER" AS ar ON ar.user_id = u.id
WHERE ar.alert_id = $1
AND u.deleted_at IS NULL AND ar.deleted_at IS NULL;`, alert.ID)

	return count, err
}

func (p *postgres) GetAlertByToken(token string) (models.Alert, error) {
	var alert models.Alert

	err := p.db.Get(&alert, `SELECT a.* FROM "ALERT" AS a
WHERE a.token = $1
AND a.deleted_at IS NULL;`, token)

	if err != nil {
		return alert, err
	}

	owner, err := p.GetUser(alert.OwnerID)
	if err != nil {
		return alert, err
	}

	alert.Owner = owner

	return alert, err
}

func (p *postgres) GetAlert(id uint) (models.Alert, error) {
	var alert models.Alert

	err := p.db.Get(&alert, `SELECT a.* FROM "ALERT" AS a
WHERE a.id = $1
AND a.deleted_at IS NULL;`, id)

	if err != nil {
		return alert, err
	}

	owner, err := p.GetUser(alert.OwnerID)
	if err != nil {
		return alert, err
	}

	alert.Owner = owner

	return alert, err
}

func (p *postgres) UpdateAlertToken(alert models.Alert) (models.Alert, error) {

	alert.ChangeToken()
	alert.UpdatedAt.Scan(time.Now())

	result, err := p.db.Exec(`UPDATE "ALERT" AS a
SET token = $1, updated_at = $2
WHERE a.id = $3
AND a.deleted_at IS NULL;`, alert.Token, alert.UpdatedAt, alert.ID)

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

func (p *postgres) GetUserAlerts(userID uint) ([]models.Alert, error) {
	var alerts []models.Alert

	err := p.db.Select(&alerts, `SELECT a.* FROM "ALERT" AS a
WHERE a.owner_id = $1
AND a.deleted_at IS NULL;`, userID)

	return alerts, err
}

func (p *postgres) GetUserSubscribedAlerts(userID uint) ([]models.Alert, error) {
	var alerts []models.Alert

	err := p.db.Select(&alerts, `SELECT a.* FROM "ALERT" AS a
INNER JOIN "ALERT_RECEIVER" AS ar ON ar.alert_id = a.id
WHERE ar.user_id = $1
AND a.deleted_at IS NULL AND ar.deleted_at IS NULL;`, userID)

	return alerts, err
}

func (p *postgres) AddUserToAlert(alert models.Alert, user models.User) error {

	alertReceiver := models.NewAlertReceiver(&user, &alert)

	result, err := p.db.NamedExec(`INSERT INTO "ALERT_RECEIVER" (alert_id, user_id, created_at)
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

func (p *postgres) RemoveUserFromAlert(alert models.Alert, user models.User) error {

	result, err := p.db.Exec(`DELETE FROM "ALERT_RECEIVER" AS ar
WHERE ar.user_id = $1 AND ar.alert_id = $2
AND ar.deleted_at IS NULL;`, user.ID, alert.ID)

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
func (p *postgres) CreateUser(user models.User) (models.User, error) {

	user.CreatedAt.Scan(time.Now())

	var id uint

	err := p.db.Get(&id, `INSERT INTO "USER" (created_at, username, telegram_id)
VALUES($1, $2, $3)
RETURNING id;`, user.CreatedAt, user.Username, user.TelegramID)

	if err != nil {
		return models.User{}, err
	}

	return p.GetUser(id)
}

func (p *postgres) GetUser(userID uint) (models.User, error) {
	var user models.User

	err := p.db.Get(&user, `SELECT u.* FROM "USER" AS u
WHERE u.id = $1
AND u.deleted_at IS NULL`, userID)

	return user, err
}

func (p *postgres) GetUserTelegram(telegramID int64) (models.User, error) {
	var user models.User

	err := p.db.Get(&user, `SELECT u.* FROM "USER" AS u
WHERE u.telegram_id = $1
AND u.deleted_at IS NULL`, telegramID)

	return user, err
}

func (p *postgres) DeleteUser(userID uint) error {

	result, err := p.db.Exec(`UPDATE "USER" AS u
SET deleted_at = $1
WHERE u.id = $2
AND u.deleted_at IS NULL`, time.Now(), userID)

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
func (p *postgres) GetDatabaseVersion() (int, error) {
	var exists bool

	err := p.db.Get(&exists, `SELECT EXISTS (
SELECT FROM information_schema.tables 
WHERE table_name = 'VERSION'
);`)

	if err != nil {
		return 0, err
	}

	if !exists {
		return -1, nil
	}

	var version int

	err = p.db.Get(&version, `SELECT v.version FROM "VERSION" AS v
ORDER BY v.version desc
LIMIT 1;`)

	if err != nil {
		return 0, err
	}

	return version, nil
}

func (p *postgres) BumpVersion(newVersion int) error {

	result, err := p.db.Exec(`INSERT INTO "VERSION" (version, updated_at)
VALUES($1, $2);`, newVersion, time.Now())

	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return sql.ErrNoRows
	}

	return nil
}

func (p *postgres) SetMuteAlert(alert models.Alert, mute bool) error {

	alert.UpdatedAt.Scan(time.Now())

	result, err := p.db.Exec(`UPDATE "ALERT" AS a
SET notify_owner = $1, updated_at = $2
WHERE a.id = $3
AND a.deleted_at IS NULL;`, mute, alert.UpdatedAt, alert.ID)

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

	return err
}
