// Package sqlitestore is the SQLite implementation of the Store type.
//
// Note that we will use a mutex to ensure consistency when using
// SQLite 3.  You can compile SQLite 3 so that it is multithread safe,
// and you might be tempted to try to do that, but eventually you will
// find that this is more trouble than it is worth and change your
// mind.  Besides, this implementation of Store actually benefits from
// lack of performance because it will make you think twice about your
// interaction with the persistence layer.
package sqlitestore

import (
	"database/sql"
	"fmt"
	"os"
	"strings"
	"sync"

	_ "github.com/mattn/go-sqlite3" // loads the SQLite3 driver
)

// sqliteStore is an implementation of the Store for SQLite3.
type sqliteStore struct {
	mu sync.Mutex
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

// New creates a new Store backed by SQLite 3.  It will create any
// missing directories.
func New(dbFile string, create bool) (*sqliteStore, error) {
	var databaseFileExisted = false
	if _, err := os.Stat(dbFile); err == nil {
		databaseFileExisted = true
	}

	d, err := sql.Open("sqlite3", dbFile)
	if err != nil {
		return nil, err
	}

	if err = d.Ping(); err != nil {
		return nil, err
	}

	if !databaseFileExisted || create {
		createSchema(d, dbFile)
	}

	store := &sqliteStore{db: d}

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
func (s *sqliteStore) Close() error {
	return s.db.Close()
}

// createSchema creates the database schema for Geo.  If this fails we
// cannot meaningfully proceed so failure to create the schema will
// cause a panic.
func createSchema(db *sql.DB, fileName string) {
	for n, statement := range strings.Split(schema, ";") {
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
