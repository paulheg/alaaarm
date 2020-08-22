package entrypoints

type Entrypoint interface {
	Trigger(token string) error
}
