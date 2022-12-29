package telegram

type Scope byte

const (
	PRIVATE_SCOPE Scope = 0x01
	GROUP_SCOPE   Scope = 0x02
	ADMIN_SCOPE   Scope = 0x04

	EVERYWHERE Scope = 0xFF
)

func (s Scope) Add(in Scope) Scope {
	return s | in
}

func (s Scope) is(scope Scope) bool {
	return s&scope == scope
}

func (s Scope) IsPrivate() bool {
	return s.is(PRIVATE_SCOPE)
}

func (s Scope) IsGroup() bool {
	return s.is(GROUP_SCOPE)
}

func (s Scope) IsAdmin() bool {
	return s.is(ADMIN_SCOPE)
}
