package dialog

// ValueStore is the interface for retrieving and setting values
type ValueStore interface {
	Value(key interface{}) interface{}
	Set(key, value interface{})
}
