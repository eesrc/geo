package postgresqlstore

import (
	"database/sql"
	"fmt"
	"strings"

	// PostgreSQL driver for production servers and Real Backends (tm)
	_ "github.com/lib/pq"
	// Loads the SQLite3 driver
	_ "github.com/mattn/go-sqlite3"
)

// sqlStore is a generic implementation of the Store
type sqlStore struct {
	db *sql.DB

	collectionStatements
	geoSubscriptionStatements
	movementStatements
	positionStatements
	shapeCollectionStatements
	shapeStatements
	subscriptionStatements
	teamStatements
	tokenStatements
	trackerStatements
	userStatements

	authStatements
}

// New creates a new Store backed by given driver.
func New(connectionString string, create bool) (*sqlStore, error) {
	var store *sqlStore

	d, err := sql.Open("postgres", connectionString)
	if err != nil {
		return nil, err
	}

	if err = d.Ping(); err != nil {
		return nil, err
	}

	if create {
		createSchema(d)
	}

	store = &sqlStore{db: d}

	if err := store.initAuthStatements(); err != nil {
		return store, fmt.Errorf("Failed to initialize auth statements: %v", err)
	}

	if err := store.initGeoSubscriptionStatements(); err != nil {
		return store, fmt.Errorf("Failed to initialize geo subscription statements: %v", err)
	}

	if err := store.initCollectionStatements(); err != nil {
		return store, fmt.Errorf("Failed to initialize collection statements: %v", err)
	}

	if err := store.initMovementStatements(); err != nil {
		return store, fmt.Errorf("Failed to initialize movement statements: %v", err)
	}

	if err := store.initPositionStatements(); err != nil {
		return store, fmt.Errorf("Failed to initialize position statements: %v", err)
	}

	if err := store.initShapeCollectionStatements(); err != nil {
		return store, fmt.Errorf("Failed to initialize shape collection statements: %v", err)
	}

	if err := store.initShapeStatements(); err != nil {
		return store, fmt.Errorf("Failed to initialize shape statements: %v", err)
	}

	if err := store.initSubscriptionStatements(); err != nil {
		return store, fmt.Errorf("Failed to initialize subscription statements: %v", err)
	}

	if err := store.initTeamStatements(); err != nil {
		return store, fmt.Errorf("Failed to initialize team statements: %v", err)
	}

	if err := store.initTokenStatements(); err != nil {
		return store, fmt.Errorf("Failed to initialize token statements: %v", err)
	}

	if err := store.initTrackerStatements(); err != nil {
		return store, fmt.Errorf("Failed to initialize tracker statements: %v", err)
	}

	if err := store.initUserStatements(); err != nil {
		return store, fmt.Errorf("Failed to initialize user statements: %v", err)
	}

	return store, nil
}

// Close closes the database
func (s *sqlStore) Close() error {
	return s.db.Close()
}

// createSchema creates the database schema for Geo.  If this fails we
// cannot meaningfully proceed so failure to create the schema will
// cause a panic.
func createSchema(db *sql.DB) {
	for n, statement := range strings.Split(postgresSchema, ";") {
		if _, err := db.Exec(statement); err != nil {
			panic(fmt.Sprintf("Statement %d failed: \"%s\" : %s", n+1, statement, err))
		}
	}
}

// rowScanner implements Scan - ie read from both sql.Row and sql.Rows. i'm not sure
// why golang doesn't implement this interface
type rowScanner interface {
	Scan(...interface{}) error
}
