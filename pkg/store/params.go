package store

type StorageParams struct {
	// DBDriver is which database driver to be used for the storage. Supported drivers are Sqlite3 and PostGres
	DBDriver           string `param:"desc=Database driver;options=postgres,sqlite3;default=sqlite3"`
	DBConnectionString string `param:"desc=Connection string for database;default=:memory:"`
	CreateDBSchema     bool   `param:"desc=Create database schema;default=true"`
}
