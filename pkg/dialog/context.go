package dialog

// Context stores values for a dialog
type Context struct {
	values map[interface{}]interface{}
	Dialog *Dialog
}

// NewContext creates a new context element
func NewContext(dialog *Dialog) *Context {
	return &Context{
		values: make(map[interface{}]interface{}),
		Dialog: dialog,
	}
}

// Value returns the value associated with the key
func (c *Context) Value(key interface{}) interface{} {
	if val, ok := c.values[key]; ok {
		return val
	}
	return nil
}

// Set the value for a key
func (c *Context) Set(key, value interface{}) {
	c.values[key] = value
}
