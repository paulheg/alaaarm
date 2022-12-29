package messages

import (
	"errors"
	"io/ioutil"

	"gopkg.in/yaml.v3"
)

var (
	ErrMessageNotFound = errors.New("there was no message found with that key")
)

type Dictionary interface {
	// Lookup a word from the dictionary
	Lookup(key string) (string, error)

	// Get the key of the dictionary (language key for example)
	Key() string
}

type dictionary struct {
	dict map[string]string
	key  string
}

func NewDictionary(filePath, dictKey string) (Dictionary, error) {
	var mapping map[string]string

	buffer, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	err = yaml.Unmarshal(buffer, &mapping)
	if err != nil {
		return nil, err
	}

	return &dictionary{
		dict: mapping,
		key:  dictKey,
	}, nil
}

func (d *dictionary) Lookup(key string) (string, error) {
	value, ok := d.dict[key]
	if !ok {
		return "", ErrMessageNotFound
	}
	return value, nil
}

func (d *dictionary) Key() string {
	return d.key
}
