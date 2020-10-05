package postgres

import (
	"io/ioutil"
	"strings"
	"time"
)

func (r *sqlxdata) runMigration(filePath string) error {

	file, err := ioutil.ReadFile(filePath)

	if err != nil {
		return err
	}

	fileContent := string(file)

	requests := strings.Split(fileContent, ";")

	for _, request := range requests {
		_, err := r.db.Exec(strings.TrimSpace(request))
		if err != nil {
			return err
		}
	}

	_, err = r.db.Exec(`INSERT INTO VERSION (updated_at, version)
VALUES ($1, $2)`, time.Now(), 1)

	if err != nil {
		return err
	}

	return nil
}
