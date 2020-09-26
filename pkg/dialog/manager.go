package dialog

// Manager manages the state of all ongoing conversations
type Manager struct {
	convState map[interface{}]*Context
	Root      *Dialog
}

// NewManager creates a new manager instance
func NewManager(root *Dialog) *Manager {
	return &Manager{
		convState: make(map[interface{}]*Context),
		Root:      root,
	}
}

// Reset the conversation for given id
func (m *Manager) Reset(id interface{}) {
	context := m.convState[id]

	if context != nil {
		context.Dialog = m.Root
		context.values = make(map[interface{}]interface{})
	}
}

// Next step in the conversation chain
func (m *Manager) Next(i interface{}, id interface{}) error {
	// get current context
	ctx := m.convState[id]

	// if this is the first interaction with the manager
	// it has to create a context struct to store the current dialog
	// and a persistant map that can store values across multiple dialogs
	if ctx == nil {
		ctx = &Context{
			Dialog: m.Root,
			values: make(map[interface{}]interface{}),
		}

		// store the context with the user/conversation id
		m.convState[id] = ctx
	}

	// reset the dialog to the root if there are no child dialogs left
	if len(ctx.Dialog.Children) == 0 {
		m.Reset(id)
	}

	// execute all children of the current dialog
	for _, child := range ctx.Dialog.Children {
		if child != nil {
			status, err := child.Function(i, ctx)
			if err != nil {
				m.Reset(id)
				return err
			}

			switch status {
			case Next:
				ctx.Dialog = child
				return m.Next(i, id)
			case Success:
				ctx.Dialog = child
				return nil
			case Reset:
				m.Reset(id)
				return nil
			case Retry:
				return nil
			case NoMatch:
				// if it did not match
			}
		}
	}

	return ErrNoMatch
}
