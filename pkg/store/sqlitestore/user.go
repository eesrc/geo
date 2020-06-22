package sqlitestore

import (
	"database/sql"

	"github.com/eesrc/geo/pkg/model"
	"github.com/eesrc/geo/pkg/store/errors"
)

type userStatements struct {
	create         *sql.Stmt
	get            *sql.Stmt
	getByGithubID  *sql.Stmt
	getByConnectID *sql.Stmt
	update         *sql.Stmt
	delete         *sql.Stmt
	list           *sql.Stmt
}

func (s *sqliteStore) initUserStatements() error {
	var err error

	if s.userStatements.create, err = s.db.Prepare(`
	INSERT INTO users (
		name,
		email,
		email_verified,
		phone,
		phone_verified,
		deleted,
		admin,
		created,
		github_id,
		connect_id
	) VALUES (
		$1,
		$2,
		$3,
		$4,
		$5,
		$6,
		$7,
		$8,
		$9,
		$10
	)`); err != nil {
		return err
	}

	if s.userStatements.get, err = s.db.Prepare(`
	SELECT
		id,
		name,
		email,
		email_verified,
		phone,
		phone_verified,
		deleted,
		admin,
		created,
		github_id,
		connect_id
	FROM users
	WHERE id = $1
	`); err != nil {
		return err
	}

	if s.userStatements.getByGithubID, err = s.db.Prepare(`
	SELECT
		id,
		name,
		email,
		email_verified,
		phone,
		phone_verified,
		deleted,
		admin,
		created,
		github_id,
		connect_id
	FROM users
	WHERE github_id = $1
	`); err != nil {
		return err
	}

	if s.userStatements.getByConnectID, err = s.db.Prepare(`
	SELECT
		id,
		name,
		email,
		email_verified,
		phone,
		phone_verified,
		deleted,
		admin,
		created,
		github_id,
		connect_id
	FROM users
	WHERE connect_id = $1
	`); err != nil {
		return err
	}

	if s.userStatements.update, err = s.db.Prepare(`
	UPDATE users
	SET
		name = $1,
		email = $2,
		email_verified = $3,
		phone = $4,
		phone_verified = $5,
		deleted = $6,
		admin = $7,
		github_id = $8,
		connect_id = $9
	WHERE id = $10
	`); err != nil {
		return err
	}

	if s.userStatements.delete, err = s.db.Prepare(`
	DELETE FROM users
	WHERE id = $1
	`); err != nil {
		return err
	}

	if s.userStatements.list, err = s.db.Prepare(`
	SELECT
		id,
		name,
		email,
		email_verified,
		phone,
		phone_verified,
		deleted,
		admin,
		created,
		github_id,
		connect_id
	FROM users
	ORDER BY
		id ASC
	LIMIT $1
	OFFSET $2
	`); err != nil {
		return err
	}

	return err
}

func (s *sqliteStore) CreateUser(user *model.User) (int64, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	r, err := s.userStatements.create.Exec(
		user.Name,
		user.Email,
		user.EmailVerified,
		user.Phone,
		user.PhoneVerified,
		user.Deleted,
		user.Admin,
		user.Created,
		user.GithubID,
		user.ConnectID,
	)

	if err != nil {
		return -1, errors.NewStorageErrorFromError(err)
	}

	lastInsertID, err := r.LastInsertId()

	return lastInsertID, errors.NewStorageErrorFromError(err)
}

func (s *sqliteStore) GetUser(userID int64) (*model.User, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	row := s.userStatements.get.QueryRow(
		userID,
	)

	user, err := scanUserRow(row)

	return &user, errors.NewStorageErrorFromError(err)
}

func (s *sqliteStore) GetUserByGithubID(userID string) (*model.User, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	row := s.userStatements.getByGithubID.QueryRow(
		userID,
	)

	user, err := scanUserRow(row)

	return &user, errors.NewStorageErrorFromError(err)
}

func (s *sqliteStore) GetUserByConnectID(userID string) (*model.User, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	row := s.userStatements.getByConnectID.QueryRow(
		userID,
	)

	user, err := scanUserRow(row)

	return &user, errors.NewStorageErrorFromError(err)
}

func (s *sqliteStore) UpdateUser(user *model.User) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	_, err := s.userStatements.update.Exec(
		user.Name,
		user.Email,
		user.EmailVerified,
		user.Phone,
		user.PhoneVerified,
		user.Deleted,
		user.Admin,
		user.GithubID,
		user.ConnectID,
		user.ID,
	)
	return errors.NewStorageErrorFromError(err)
}

func (s *sqliteStore) DeleteUser(userID int64) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	_, err := s.userStatements.delete.Exec(
		userID,
	)

	return errors.NewStorageErrorFromError(err)
}

func (s *sqliteStore) ListUsers(offset int64, limit int64) ([]model.User, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	var users []model.User
	rows, err := s.userStatements.list.Query(
		limit,
		offset,
	)

	if err != nil {
		return users, errors.NewStorageErrorFromError(err)
	}

	defer rows.Close()

	for rows.Next() {
		user, err := scanUserRow(rows)

		if err != nil {
			return users, errors.NewStorageErrorFromError(err)
		}

		users = append(users, user)
	}

	return users, nil
}

func scanUserRow(row rowScanner) (model.User, error) {
	user := model.User{}

	err := row.Scan(
		&user.ID,
		&user.Name,
		&user.Email,
		&user.EmailVerified,
		&user.Phone,
		&user.PhoneVerified,
		&user.Deleted,
		&user.Admin,
		&user.Created,
		&user.GithubID,
		&user.ConnectID,
	)

	return user, err
}
