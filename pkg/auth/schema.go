package auth

import "database/sql"

const (
	createOauthStateTable = `
	CREATE TABLE IF NOT EXISTS oauth_state (
		state          VARCHAR(128)   NOT NULL,
		CONSTRAINT oauthstate_pk PRIMARY KEY (state)
	)`
	createOauthSessionTable = `
	CREATE TABLE IF NOT EXISTS oauth_session (
		session_id     VARCHAR(128)  NOT NULL,
		access_token   VARCHAR(128)  NOT NULL,
		expires        BIGINT        NOT NULL,
		profile        JSON          NOT NULL,
		CONSTRAINT oauthsession_pk PRIMARY KEY (session_id)
	)`
	createOauthSessionTableIndex = `CREATE INDEX IF NOT EXISTS oauthsession_expires ON oauth_session(expires)`
)

type statements struct {
	createState     *sql.Stmt
	removeState     *sql.Stmt
	createSession   *sql.Stmt
	retrieveSession *sql.Stmt
	removeSession   *sql.Stmt
	listSessions    *sql.Stmt
	refreshSession  *sql.Stmt
}

func initStatements(db *sql.DB) (statements, error) {
	var err error
	var statements statements

	if statements.createState, err = db.Prepare(`
		INSERT INTO oauth_state (state)
			VALUES ($1)`); err != nil {
		return statements, err
	}
	if statements.removeState, err = db.Prepare(`
		DELETE FROM oauth_state
			WHERE state = $1`); err != nil {
		return statements, err
	}
	if statements.createSession, err = db.Prepare(`
		INSERT INTO oauth_session (session_id, access_token, expires, profile)
			VALUES ($1, $2, $3, $4)
	`); err != nil {
		return statements, err
	}

	if statements.retrieveSession, err = db.Prepare(`
		SELECT session_id, access_token, expires, profile
			FROM oauth_session
			WHERE session_id = $1 AND expires > $2`); err != nil {
		return statements, err
	}
	if statements.removeSession, err = db.Prepare(`
		DELETE FROM oauth_session
			WHERE session_id = $1`); err != nil {
		return statements, err
	}
	if statements.listSessions, err = db.Prepare(`
		SELECT session_id, access_token, expires, profile
			FROM oauth_session WHERE expires < $1`); err != nil {
		return statements, err
	}
	if statements.refreshSession, err = db.Prepare(`
		UPDATE oauth_session
			SET expires = expires + $1
			WHERE session_id = $2`); err != nil {
		return statements, err
	}
	return statements, nil
}
