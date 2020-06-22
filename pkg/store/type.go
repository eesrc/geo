package store

type DBDriver string

const (
	SQLite     DBDriver = "sqlite3"
	PostgreSQL DBDriver = "postgres"
)
