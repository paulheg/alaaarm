package repository

// Configuration stores the database configuration
type Configuration struct {
	ConnectionString   string `env:"CONNECTION_STRING"`
	MigrationDirectory string `env:"MIGRATION_DIR"`
}
