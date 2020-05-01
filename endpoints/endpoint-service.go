package endpoints

// EndpointService represents an endpoint that can send notifications to all receivers
type EndpointService interface {
	NotifyAll(message string) error
	Run() error
}
