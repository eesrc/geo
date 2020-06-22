package auth

import (
	"database/sql"
	"errors"
	"sync"
)

type sqliteSessionStore struct {
	mutex *sync.Mutex
	db    *sql.DB
	s     statements
}

func (s *sqliteSessionStore) createSchema() error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

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

func (s *sqliteSessionStore) init() error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	statements, err := initStatements(s.db)

	if err != nil {
		return err
	}

	s.s = statements

	return nil
}
func (s *sqliteSessionStore) PutState(state string) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

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

func (s *sqliteSessionStore) RemoveState(state string) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

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

func (s *sqliteSessionStore) CreateSession(accessToken string, expires int64, profile Profile) (string, error) {
	sessionID := createNewSessionID(s)

	s.mutex.Lock()
	defer s.mutex.Unlock()
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

func (s *sqliteSessionStore) GetSession(sessionID string, expireTime int64) (Session, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	row := s.s.retrieveSession.QueryRow(sessionID, expireTime)
	if row == nil {
		return Session{}, errors.New("not found")
	}
	ret := Session{}
	return ret, row.Scan(&ret.ID, &ret.AccessToken, &ret.Expires, &ret.Profile)

}

func (s *sqliteSessionStore) RemoveSession(sessionID string) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

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

func (s *sqliteSessionStore) GetSessions(time int64) ([]Session, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

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

func (s *sqliteSessionStore) RefreshSession(sessionID string, checkInterval int64) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

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
