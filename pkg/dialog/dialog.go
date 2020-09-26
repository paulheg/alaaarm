package dialog

import (
	"errors"
)

// Failable represents a function that can fail
type Failable func(i interface{}, ctx ValueStore) (Status, error)

// Dialog represents a chain of steps
type Dialog struct {
	Parent   *Dialog
	Children []*Dialog
	Function Failable
}

// Status represents the return of a failable function
type Status int

const (
	// NoMatch indicates that the function did not match to the input
	NoMatch = iota

	// Reset indicates to the dialog.Manager that the dialog has to be reset
	Reset

	// Next indicates to the dialog.Manager to run the next function with the same input
	Next

	// Success indicates that the function executed as expected
	Success

	// Retry indicates that the manager has to retry the current step with the next input
	Retry
)

var (
	// ErrNoMatch is returned if the dialog can't continue because
	// there are no matching childdialogs
	ErrNoMatch = errors.New("there was no match for the given input")

	// ErrReset is returned when the dialog chain has to be resetted
	ErrReset = errors.New("there was an error that resets the dialog")
)

// NewRoot creates a root dialog
func NewRoot() *Dialog {

	emptyFunc := func(i interface{}, ctx ValueStore) (Status, error) {
		return Next, nil
	}

	return &Dialog{
		Parent:   nil,
		Function: emptyFunc,
	}

}

// Chain a root dialog
func Chain(f Failable) *Dialog {
	if f == nil {
		panic("failable f is nil")
	}

	newDialog := &Dialog{
		Parent:   nil,
		Function: f,
	}

	return newDialog
}

// Root of the dialog
func (d *Dialog) Root() *Dialog {
	current := d
	for current.Parent != nil {
		current = current.Parent
	}

	return current
}

// Chain a new dialog
func (d *Dialog) Chain(f Failable) *Dialog {

	if f == nil {
		panic("failable f is nil")
	}

	newDialog := &Dialog{
		Parent:   d,
		Function: f,
	}

	d.Children = append(d.Children, newDialog)

	return newDialog
}

// Append an existing dialog chain
func (d *Dialog) Append(dialog *Dialog) *Dialog {

	d.Children = append(d.Children, dialog.Root())
	dialog.Root().Parent = d

	return dialog
}

// Branch adds Multiple Dialog options to the tree
func (d *Dialog) Branch(dialogs ...*Dialog) {
	for _, dialog := range dialogs {
		// dialogs added will be in reversed order
		// therefore chain local root

		d.Children = append(d.Children, dialog.Root())

	}
}
