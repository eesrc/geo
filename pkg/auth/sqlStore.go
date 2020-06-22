package auth

import (
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"errors"
	"io"
	"sync"

	// SQLite3 driver for testing, local instances and in-memory database
	_ "github.com/mattn/go-sqlite3"
	//PostgreSQL driver for production servers and Real Backends (tm)
	_ "github.com/lib/pq"
)

func newSqliteStore(driver string, connectionString string) (*sqliteSessionStore, error) {
	var err error

	ret := sqliteSessionStore{
		mutex: &sync.Mutex{},
	}

	if ret.db, err = sql.Open(driver, connectionString); err != nil {
		return &ret, err
	}

	if err := ret.db.Ping(); err != nil {
		return &ret, err
	}

	if err := ret.createSchema(); err != nil {
		return &ret, err
	}

	if err := ret.init(); err != nil {
		return &ret, err
	}

	return &ret, nil
}

// NewSQLSessionStore creates a sql-backed session streo
func NewSQLSessionStore(driver, connectionString string) (SessionStore, error) {
	// SQLite in-memory is a special little flower and needs its mutexes
	if driver == "sqlite3" {
		return newSqliteStore(driver, connectionString)
	}

	var err error
	ret := sqlSessionStore{}
	if ret.db, err = sql.Open(driver, connectionString); err != nil {
		return nil, err
	}

	if err := ret.db.Ping(); err != nil {
		return nil, err
	}

	if err := ret.createSchema(); err != nil {
		return nil, err
	}

	if err := ret.init(); err != nil {
		return nil, err
	}

	return &ret, nil
}

type sqlSessionStore struct {
	db *sql.DB
	s  statements
}

func (s *sqlSessionStore) createSchema() error {
	_, err := s.db.Exec(createOauthStateTable)
	if err != nil {
		return err
	}

	_, err = s.db.Exec(createOauthSessionTable)
	if err != nil {
		return err
	}
	_, err = s.db.Exec(createOauthSessionTableIndex)
	if err != nil {
		return err
	}

	return nil
}

func (s *sqlSessionStore) init() error {

	statements, err := initStatements(s.db)

	if err != nil {
		return err
	}

	s.s = statements

	return nil
}
func (s *sqlSessionStore) PutState(state string) error {
	rows, err := s.s.createState.Exec(state)
	if err != nil {
		return err
	}
	if count, err := rows.RowsAffected(); err != nil || count == 0 {
		if err != nil {
			return err
		}
		return errors.New("state not stored")
	}
	return nil
}

func (s *sqlSessionStore) RemoveState(state string) error {
	rows, err := s.s.removeState.Exec(state)
	if err != nil {
		return err
	}
	count, err := rows.RowsAffected()
	if err != nil {
		return err
	}
	if count == 0 {
		return errors.New("state not found")
	}
	return nil
}

func (s *sqlSessionStore) CreateSession(accessToken string, expires int64, profile Profile) (string, error) {
	sessionID := createNewSessionID(s)
	rows, err := s.s.createSession.Exec(&sessionID, &accessToken, &expires, &profile)
	if err != nil {
		return "", err
	}
	count, err := rows.RowsAffected()
	if err != nil {
		return "", err
	}
	if count == 0 {
		return "", errors.New("no session created")
	}
	return sessionID, nil
}

func (s *sqlSessionStore) GetSession(sessionID string, expireTime int64) (Session, error) {
	row := s.s.retrieveSession.QueryRow(sessionID, expireTime)
	if row == nil {
		return Session{}, errors.New("not found")
	}
	ret := Session{}
	return ret, row.Scan(&ret.ID, &ret.AccessToken, &ret.Expires, &ret.Profile)
}

func (s *sqlSessionStore) RemoveSession(sessionID string) error {
	res, err := s.s.removeSession.Exec(sessionID)
	if err != nil {
		return err
	}
	count, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if count == 0 {
		return errors.New("session not found")
	}
	return nil
}

func (s *sqlSessionStore) GetSessions(time int64) ([]Session, error) {
	rows, err := s.s.listSessions.Query(time)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	ret := make([]Session, 0)
	for rows.Next() {
		sess := Session{}
		if err := rows.Scan(&sess.ID, &sess.AccessToken, &sess.Expires, &sess.Profile); err != nil {
			return ret, err
		}
		ret = append(ret, sess)
	}
	return ret, nil

}

func (s *sqlSessionStore) RefreshSession(sessionID string, checkInterval int64) error {
	res, err := s.s.refreshSession.Exec(checkInterval, sessionID)
	if err != nil {
		return err
	}
	count, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if count == 0 {
		return errors.New("session not found")
	}
	return nil
}

// createNewSessionID generates a random 512 bit session id. Also
// check for collisions so we don't accidentally assign an existing
// session id.  If we cannot generate a sessionID we eventually panic
// because this means something is very wrong.
func createNewSessionID(s SessionStore) string {
	numCollisionsBeforeWeGiveUp := 5
	for i := 0; i < numCollisionsBeforeWeGiveUp; i++ {
		b := make([]byte, 64)
		if _, err := io.ReadFull(rand.Reader, b); err != nil {
			panic("Unable to generate random data for session id")
		}

		encoded := base64.URLEncoding.EncodeToString(b)
		if _, err := s.GetSession(encoded, 0); err != nil {
			return encoded
		}
	}

	// We end up here if we have given up creating a session id.
	panic("Unable to create new session id")
}
