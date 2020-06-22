package postgresqlstore

import (
	"database/sql"
	"fmt"

	"github.com/eesrc/geo/pkg/model"
	"github.com/eesrc/geo/pkg/store/errors"
)

type tokenStatements struct {
	create       *sql.Stmt
	update       *sql.Stmt
	get          *sql.Stmt
	getByUserID  *sql.Stmt
	delete       *sql.Stmt
	list         *sql.Stmt
	listByUserID *sql.Stmt
}

func (s *sqlStore) initTokenStatements() error {
	var err error

	if s.tokenStatements.create, err = s.db.Prepare(`
	INSERT INTO tokens (
		token,
		resource,
		user_id,
		perm_write,
		created
	) VALUES (
		$1,
		$2,
		$3,
		$4,
		$5
	)
	`); err != nil {
		return err
	}

	if s.tokenStatements.get, err = s.db.Prepare(`
	SELECT
		token,
		resource,
		user_id,
		perm_write,
		created
	FROM tokens
	WHERE token = $1
	`); err != nil {
		return err
	}

	if s.tokenStatements.getByUserID, err = s.db.Prepare(`
	SELECT
		token,
		resource,
		user_id,
		perm_write,
		created
	FROM tokens
	WHERE
		token = $1
		AND
		user_id = $2
	`); err != nil {
		return err
	}

	if s.tokenStatements.update, err = s.db.Prepare(`
	UPDATE tokens
	SET
		resource = $1,
		perm_write = $2
	WHERE
		token = $3
		AND
		user_id = $4
	`); err != nil {
		return err
	}

	if s.tokenStatements.delete, err = s.db.Prepare(`
	DELETE FROM tokens
	WHERE
		token = $1
		AND
		user_id = $2
	`); err != nil {
		return err
	}

	if s.tokenStatements.list, err = s.db.Prepare(`
	SELECT
		token,
		resource,
		user_id,
		perm_write,
		created
	FROM tokens
	ORDER BY
		created ASC
	LIMIT $1
	OFFSET $2
	`); err != nil {
		return err
	}

	if s.tokenStatements.listByUserID, err = s.db.Prepare(`
	SELECT
		token,
		resource,
		user_id,
		perm_write,
		created
	FROM tokens
	WHERE user_id = $1
	ORDER BY
		created ASC
	LIMIT $2
	OFFSET $3
	`); err != nil {
		return err
	}

	return err
}

func (s *sqlStore) CreateToken(token *model.Token) (string, error) {
	_, err := s.tokenStatements.create.Exec(
		token.Token,
		token.Resource,
		token.UserID,
		token.PermWrite,
		token.Created,
	)

	return token.Token, errors.NewStorageErrorFromError(err)
}

func (s *sqlStore) UpdateToken(token *model.Token) error {
	res, err := s.tokenStatements.update.Exec(
		token.Resource,
		token.PermWrite,
		token.Token,
		token.UserID,
	)

	if err != nil {
		return errors.NewStorageErrorFromError(err)
	}

	rowsAffected, err := res.RowsAffected()

	if err == nil && rowsAffected == 0 {
		return errors.NewStorageError(errors.AccessDeniedError, fmt.Errorf("No token updated"))
	}

	return errors.NewStorageErrorFromError(err)
}

func (s *sqlStore) GetToken(token string) (*model.Token, error) {
	row := s.tokenStatements.get.QueryRow(
		token,
	)

	tokenModel, err := scanTokenRow(row)

	return &tokenModel, errors.NewStorageErrorFromError(err)
}

func (s *sqlStore) GetTokenByUserID(token string, userID int64) (*model.Token, error) {
	row := s.tokenStatements.getByUserID.QueryRow(
		token,
		userID,
	)

	tokenModel, err := scanTokenRow(row)

	if err == sql.ErrNoRows {
		return &tokenModel, errors.NewStorageError(errors.AccessDeniedError, err)
	}

	return &tokenModel, errors.NewStorageErrorFromError(err)
}

func (s *sqlStore) DeleteToken(token string, userID int64) error {
	res, err := s.tokenStatements.delete.Exec(
		token,
		userID,
	)

	if err != nil {
		return errors.NewStorageErrorFromError(err)
	}

	rowsAffected, err := res.RowsAffected()

	if err == nil && rowsAffected == 0 {
		return errors.NewStorageError(errors.AccessDeniedError, fmt.Errorf("No token deleted"))
	}

	return errors.NewStorageErrorFromError(err)
}

func (s *sqlStore) ListTokens(offset int64, limit int64) ([]model.Token, error) {
	var tokens []model.Token
	rows, err := s.tokenStatements.list.Query(
		limit,
		offset,
	)

	if err != nil {
		return tokens, errors.NewStorageErrorFromError(err)
	}

	defer rows.Close()

	for rows.Next() {
		token, err := scanTokenRow(rows)

		if err != nil {
			return tokens, errors.NewStorageErrorFromError(err)
		}

		tokens = append(tokens, token)
	}

	return tokens, nil
}

func (s *sqlStore) ListTokensByUserID(userID int64, offset int64, limit int64) ([]model.Token, error) {
	var tokens []model.Token
	rows, err := s.tokenStatements.listByUserID.Query(
		userID,
		limit,
		offset,
	)

	if err != nil {
		return tokens, errors.NewStorageErrorFromError(err)
	}

	defer rows.Close()

	for rows.Next() {
		token, err := scanTokenRow(rows)

		if err != nil {
			return tokens, errors.NewStorageErrorFromError(err)
		}

		tokens = append(tokens, token)
	}

	return tokens, nil
}

func scanTokenRow(row rowScanner) (model.Token, error) {
	token := model.Token{}

	err := row.Scan(
		&token.Token,
		&token.Resource,
		&token.UserID,
		&token.PermWrite,
		&token.Created,
	)

	return token, err
}
