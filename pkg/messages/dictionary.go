package messages

import "errors"

var (
	ErrMessageNotFound = errors.New("there was no message found with that key")
)

type Dictionary interface {
	Lookup(key string) (string, error)
}

type dictionary struct {
	dict map[string]string
}

func NewDictionary() Dictionary {
	return &dictionary{
		dict: make(map[string]string, 0),
	}
}

func (d *dictionary) Lookup(key string) (string, error) {
	value, ok := d.dict[key]
	if !ok {
		return "", ErrMessageNotFound
	}
	return value, nil
}
