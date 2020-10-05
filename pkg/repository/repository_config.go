package repository

// Configuration stores the database configuration
type Configuration struct {
	ConnectionString   string
	Database           string
	MigrationDirectory string
}
