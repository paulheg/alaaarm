package messages

import (
	"errors"
	"io/ioutil"
	"path"
)

var (
	ErrDefaultLanguageNotFound = errors.New("the default language was not found")
)

type Library interface {
	// Get the dictionary for a selected language
	// if the language does not exist, return the default language
	Get(lang string) Dictionary

	// Get the default language dictionary
	Default() Dictionary
}

// Load all languages yaml files from a base directory
// for example:
// en.yaml
// de.yaml
//
// Using the base directory localization, the library
// will load languages en, de.
func NewLibrary(baseDirectory, defaultLanguage string) (Library, error) {

	lib := &library{
		dicts: make(map[string]Dictionary, 5),
	}

	files, err := ioutil.ReadDir(baseDirectory)
	if err != nil {
		return nil, err
	}

	for _, f := range files {

		// do not read directories
		if f.IsDir() {
			continue
		}

		fileName := f.Name()
		key := fileName[len(fileName)-len(path.Ext(fileName)):]

		dict, err := NewDictionary(path.Join(baseDirectory, fileName), key)
		if err != nil {
			return nil, err
		}

		lib.dicts[key] = dict

		if key == defaultLanguage {
			lib.defaultDictionary = dict
		}
	}

	if lib.defaultDictionary == nil {
		return nil, ErrDefaultLanguageNotFound
	}

	return lib, nil
}

type library struct {
	dicts             map[string]Dictionary
	defaultDictionary Dictionary
}

func (l *library) Get(lang string) Dictionary {
	if dict, ok := l.dicts[lang]; ok {
		return dict
	} else {
		return l.defaultDictionary
	}
}

func (l *library) Default() Dictionary {
	return l.defaultDictionary
}
