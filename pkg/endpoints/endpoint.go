package endpoints

import "mime/multipart"

// Endpoint represents an endpoint that can send notifications to all receivers
type Endpoint interface {
	NotifyAll(token string, message string, file *multipart.FileHeader) error
}
