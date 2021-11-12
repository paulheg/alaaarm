package migration

import (
	"database/sql"
	"io/ioutil"
	"path"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/sirupsen/logrus"
)

type Migrator interface {
	Migrate() error
}

type migrator struct {
	db                     *sql.DB
	version                VersionRepository
	migrationFileDirectory string
	log                    *logrus.Logger
}

func New(db *sql.DB, version VersionRepository, migrationFileDirectory string, log *logrus.Logger) Migrator {
	return &migrator{
		db:                     db,
		version:                version,
		migrationFileDirectory: migrationFileDirectory,
		log:                    log,
	}
}

func (m *migrator) Migrate() error {
	m.log.Debug("Starting database migration")

	version, err := m.version.GetDatabaseVersion()
	if err != nil {
		m.log.WithError(err).Debug("Error while getting the database version")
		return err
	}

	files, err := ioutil.ReadDir(m.migrationFileDirectory)
	if err != nil {
		return err
	}

	var filenames []string

	for _, file := range files {
		name := file.Name()
		if !file.IsDir() && filepath.Ext(name) == ".sql" {
			fileVersion := m.getFileVersion(name)

			if fileVersion <= version {
				continue
			}

			filenames = append(filenames, name)
		}
	}

	if len(filenames) == 0 {
		m.log.Debug("No new migration files where found")
	} else {
		sort.Strings(filenames)

		for _, filename := range filenames {
			fileVersion := m.getFileVersion(filename)

			migrationFile := path.Join(m.migrationFileDirectory, filename)

			m.log.WithField("file", migrationFile).WithField("version", version).Debug("Running migration file")

			err = m.runMigration(migrationFile, fileVersion)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (m *migrator) getFileVersion(fileName string) int {

	index := strings.Index(fileName, "_")

	// string that only contains the version number
	versionString := fileName[:index]

	version, err := strconv.Atoi(versionString)
	if err != nil {
		panic("migration file does not contain version in filename")
	}

	return version
}

func (m *migrator) runMigration(filePath string, fileVersion int) error {

	file, err := ioutil.ReadFile(filePath)

	if err != nil {
		return err
	}

	fileContent := string(file)

	requests := strings.Split(fileContent, ";")

	tx, err := m.db.Begin()
	if err != nil {
		return err
	}

	for _, request := range requests {
		_, err := m.db.Exec(strings.TrimSpace(request))
		if err != nil {
			tx.Rollback()
			return err
		}
	}

	err = m.version.BumpVersion(fileVersion)
	if err != nil {
		tx.Rollback()
		return err
	}
	err = tx.Commit()

	return err
}
