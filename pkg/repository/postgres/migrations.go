package postgres

import (
	"io/ioutil"
	"strings"
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

	return nil
}
