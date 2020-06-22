package auth

type AuthenticatorConfig struct {
	// DBDriver is the DB driver to be used for persistent sessions. Defaults to sqlite
	DBDriver string `param:"desc=Database driver;options=postgres,sqlite3;default=sqlite3"`
	// DBConnectionString is the connection string for the DB. Defaults to in-memory.
	DBConnectionString string `param:"desc=Database connection string;default=:memory:"`
}
