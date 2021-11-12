package migration

// VersionRepository defines a method to check the version of the database
type VersionRepository interface {

	// Gets the current version of the database
	// If this is the first setup of the database, this call could fail because of missing structures in the database
	// In this case the version should be -1 and no error should be returned
	// Be careful not to hide implementation mistakes that way
	GetDatabaseVersion() (int, error)
	BumpVersion(newVersion int) error
}
